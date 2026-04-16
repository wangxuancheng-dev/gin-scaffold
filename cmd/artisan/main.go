package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"

	"gin-scaffold/config"
	"gin-scaffold/internal/console"
	_ "gin-scaffold/internal/console/commands"
)

func main() {
	var env string
	var profile string
	rootCmd := &cobra.Command{
		Use:   "artisan",
		Short: "application console commands",
	}
	rootCmd.PersistentFlags().StringVar(&env, "env", "dev", "配置环境: dev|test|prod")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "配置画像: 多实例标识，如 order/crm")
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newMakeCommandCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newQueueFailedCommand(&env, &profile))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list registered artisan commands",
		Run: func(cmd *cobra.Command, args []string) {
			for _, item := range console.List() {
				fmt.Printf("%-24s %s\n", item.Name(), item.Description())
			}
		},
	}
}

func newRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> [args...]",
		Short: "run an artisan command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return console.Execute(context.Background(), args[0], args[1:], os.Stdout)
		},
	}
}

func newMakeCommandCommand() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "make:command <name>",
		Short: "generate artisan command template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateCommandTemplate(args[0], force)
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing command file")
	return c
}

func newQueueFailedCommand(env, profile *string) *cobra.Command {
	root := &cobra.Command{
		Use:   "queue:failed",
		Short: "inspect and retry failed(asynq archived) tasks",
	}
	var queueOverride string
	root.Flags().StringVarP(&queueOverride, "queue", "q", "", "asynq queue name, default from config")
	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list archived tasks from queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*env, *profile)
			if err != nil {
				return err
			}
			ins := asynq.NewInspector(asynq.RedisClientOpt{
				Addr:     cfg.Asynq.RedisAddr,
				Password: cfg.Asynq.RedisPassword,
				DB:       cfg.Asynq.RedisDB,
			})
			defer func() { _ = ins.Close() }()
			queueName := cfg.Asynq.Queue
			if queueOverride != "" {
				queueName = queueOverride
			}
			tasks, err := ins.ListArchivedTasks(queueName, asynq.PageSize(20))
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Printf("no archived tasks in queue=%s\n", queueName)
				return nil
			}
			for _, t := range tasks {
				fmt.Printf("id=%s type=%s retries=%d/%d\n", t.ID, t.Type, t.Retried, t.MaxRetry)
				if len(t.Payload) > 0 {
					var payload any
					if err := json.Unmarshal(t.Payload, &payload); err == nil {
						b, _ := json.Marshal(payload)
						fmt.Printf("  payload=%s\n", string(b))
					}
				}
				if t.LastErr != "" {
					fmt.Printf("  error=%s\n", t.LastErr)
				}
			}
			return nil
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "retry <task_id>",
		Short: "retry an archived task by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*env, *profile)
			if err != nil {
				return err
			}
			ins := asynq.NewInspector(asynq.RedisClientOpt{
				Addr:     cfg.Asynq.RedisAddr,
				Password: cfg.Asynq.RedisPassword,
				DB:       cfg.Asynq.RedisDB,
			})
			defer func() { _ = ins.Close() }()
			queueName := cfg.Asynq.Queue
			if queueOverride != "" {
				queueName = queueOverride
			}
			if err := ins.RunTask(queueName, args[0]); err != nil {
				return err
			}
			fmt.Printf("retry task ok: queue=%s id=%s\n", queueName, args[0])
			return nil
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "retry-all [n]",
		Short: "retry first n archived tasks (default 20)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := 20
			if len(args) == 1 {
				n, err := strconv.Atoi(args[0])
				if err != nil || n <= 0 {
					return fmt.Errorf("invalid n: %s", args[0])
				}
				limit = n
			}
			cfg, err := config.Load(*env, *profile)
			if err != nil {
				return err
			}
			queueName := cfg.Asynq.Queue
			if queueOverride != "" {
				queueName = queueOverride
			}
			ins := asynq.NewInspector(asynq.RedisClientOpt{
				Addr:     cfg.Asynq.RedisAddr,
				Password: cfg.Asynq.RedisPassword,
				DB:       cfg.Asynq.RedisDB,
			})
			defer func() { _ = ins.Close() }()
			tasks, err := ins.ListArchivedTasks(queueName, asynq.PageSize(limit))
			if err != nil {
				return err
			}
			okCount := 0
			for _, t := range tasks {
				if err := ins.RunTask(queueName, t.ID); err == nil {
					okCount++
				}
			}
			fmt.Printf("retry-all done: queue=%s success=%d total=%d\n", queueName, okCount, len(tasks))
			return nil
		},
	})
	return root
}

func generateCommandTemplate(rawName string, force bool) error {
	name := normalizeCommandName(rawName)
	if name == "" {
		return fmt.Errorf("invalid command name: %s", rawName)
	}
	structName := toPascal(strings.ReplaceAll(name, ":", "_")) + "Command"
	filePath := filepath.Join("internal", "console", "commands", strings.ReplaceAll(name, ":", "_")+".go")
	if _, err := os.Stat(filePath); err == nil && !force {
		return fmt.Errorf("file exists: %s (use --force)", filePath)
	}
	content := fmt.Sprintf(`package commands

import (
	"context"
	"fmt"
	"io"

	"gin-scaffold/internal/console"
)

type %s struct{}

func (c *%s) Name() string { return %q }

func (c *%s) Description() string { return "TODO: describe command" }

func (c *%s) Run(ctx context.Context, args []string, out io.Writer) error {
	_, _ = fmt.Fprintf(out, "command %%s args=%%v\n", c.Name(), args)
	return nil
}

func init() {
	console.Register(&%s{})
}
`, structName, structName, name, structName, structName, structName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Println("generated:", filePath)
	return nil
}

func normalizeCommandName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, " ", ":")
	parts := strings.Split(name, ":")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var b strings.Builder
		for _, r := range p {
			if unicode.IsLower(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
				b.WriteRune(r)
			}
		}
		part := strings.Trim(b.String(), "-_")
		if part == "" {
			return ""
		}
		out = append(out, part)
	}
	return strings.Join(out, ":")
}

func toPascal(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ':' || r == ' '
	})
	var b strings.Builder
	for _, f := range fields {
		runes := []rune(f)
		if len(runes) == 0 {
			continue
		}
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}
