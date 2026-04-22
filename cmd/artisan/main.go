package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"

	"gin-scaffold/internal/config"
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
	rootCmd.AddCommand(newMakeModelCommand())
	rootCmd.AddCommand(newMakeDAOCommand())
	rootCmd.AddCommand(newGenerateEncryptionKeyCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newQueueFailedCommand(&env, &profile))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newGenerateEncryptionKeyCommand() *cobra.Command {
	var raw bool
	c := &cobra.Command{
		Use:   "key:generate",
		Short: "generate ENCRYPTION_KEY for config",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := make([]byte, 32)
			if _, err := rand.Read(buf); err != nil {
				return err
			}
			encoded := base64.StdEncoding.EncodeToString(buf)
			if raw {
				fmt.Fprintln(cmd.OutOrStdout(), encoded)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "base64:%s\n", encoded)
			return nil
		},
	}
	c.Flags().BoolVar(&raw, "raw", false, "print raw base64 without prefix")
	return c
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

func newMakeModelCommand() *cobra.Command {
	var (
		tableName      string
		force          bool
		withSoftDelete bool
	)
	c := &cobra.Command{
		Use:   "make:model <name>",
		Short: "generate model template in internal/model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateModelTemplate(args[0], tableName, withSoftDelete, force)
		},
	}
	c.Flags().StringVar(&tableName, "table", "", "custom table name (default: snake_case plural)")
	c.Flags().BoolVar(&force, "force", false, "overwrite existing model file")
	c.Flags().BoolVar(&withSoftDelete, "with-soft-delete", false, "include DeletedAt gorm.DeletedAt field")
	return c
}

func newMakeDAOCommand() *cobra.Command {
	var (
		force      bool
		withTenant bool
		withTx     bool
	)
	c := &cobra.Command{
		Use:   "make:dao <name>",
		Short: "generate dao template in internal/dao",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateDAOTemplate(args[0], withTenant, withTx, force)
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing dao file")
	c.Flags().BoolVar(&withTenant, "with-tenant", false, "include tenant.ApplyScope in query paths")
	c.Flags().BoolVar(&withTx, "tx", false, "include WithTx(tx *gorm.DB) helper")
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

func generateModelTemplate(rawName, tableName string, withSoftDelete, force bool) error {
	name := normalizeModelName(rawName)
	if name == "" {
		return fmt.Errorf("invalid model name: %s", rawName)
	}
	structName := toPascal(name)
	filePath := filepath.Join("internal", "model", name+".go")
	if _, err := os.Stat(filePath); err == nil && !force {
		return fmt.Errorf("file exists: %s (use --force)", filePath)
	}
	finalTable := strings.TrimSpace(tableName)
	if finalTable == "" {
		finalTable = pluralizeSnake(name)
	}
	nowFields := "CreatedAt time.Time `json:\"created_at\"`\n\tUpdatedAt time.Time `json:\"updated_at\"`"
	softDeleteField := ""
	gormImport := ""
	if withSoftDelete {
		gormImport = "\n\t\"gorm.io/gorm\""
		softDeleteField = "\n\tDeletedAt gorm.DeletedAt `gorm:\"index\" json:\"-\"`"
	}
	content := fmt.Sprintf(`package model

import (
	"time"%s
)

// %s 对应表 %s。
type %s struct {
	ID int64 `+"`gorm:\"primaryKey;autoIncrement\" json:\"id\"`"+`
	%s%s
}

// TableName 指定表名。
func (%s) TableName() string {
	return %q
}
`, gormImport, structName, finalTable, structName, nowFields, softDeleteField, structName, finalTable)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Println("generated:", filePath)
	return nil
}

func generateDAOTemplate(rawName string, withTenant, withTx, force bool) error {
	name := normalizeModelName(rawName)
	if name == "" {
		return fmt.Errorf("invalid dao name: %s", rawName)
	}
	structName := toPascal(name) + "DAO"
	modelName := toPascal(name)
	filePath := filepath.Join("internal", "dao", name+"_dao.go")
	if _, err := os.Stat(filePath); err == nil && !force {
		return fmt.Errorf("file exists: %s (use --force)", filePath)
	}

	tenantImport := ""
	scopeAssign := "q := d.db.WithContext(ctx)"
	if withTenant {
		tenantImport = "\n\t\"gin-scaffold/internal/pkg/tenant\""
		scopeAssign = "q := tenant.ApplyScope(ctx, d.db.WithContext(ctx), \"tenant_id\")"
	}
	withTxMethod := ""
	if withTx {
		withTxMethod = fmt.Sprintf(`
func (d *%s) WithTx(tx *gorm.DB) *%s {
	if tx == nil {
		return d
	}
	return &%s{db: tx}
}
`, structName, structName, structName)
	}

	content := fmt.Sprintf(`package dao

import (
	"context"

	"gin-scaffold/internal/model"%s
	"gorm.io/gorm"
)

type %s struct {
	db *gorm.DB
}

func New%s(db *gorm.DB) *%s {
	return &%s{db: db}
}
%s
func (d *%s) Create(ctx context.Context, in *model.%s) error {
	return d.db.WithContext(ctx).Create(in).Error
}

func (d *%s) GetByID(ctx context.Context, id int64) (*model.%s, error) {
	%s
	var row model.%s
	if err := q.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *%s) List(ctx context.Context, page, pageSize int) ([]model.%s, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	%s
	base := q.Model(&model.%s{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.%s
	if err := base.Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *%s) Update(ctx context.Context, id int64, patch map[string]any) error {
	if len(patch) == 0 {
		return nil
	}
	%s
	return q.Model(&model.%s{}).Where("id = ?", id).Updates(patch).Error
}

func (d *%s) Delete(ctx context.Context, id int64) error {
	%s
	return q.Where("id = ?", id).Delete(&model.%s{}).Error
}
`,
		tenantImport,
		structName,
		structName, structName,
		structName,
		withTxMethod,
		structName, modelName,
		structName, modelName,
		scopeAssign,
		modelName,
		structName, modelName,
		scopeAssign,
		modelName,
		modelName,
		structName,
		scopeAssign,
		modelName,
		structName,
		scopeAssign,
		modelName,
	)

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

func normalizeModelName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	parts := strings.Split(name, "_")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var b strings.Builder
		for _, r := range p {
			if unicode.IsLower(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
			}
		}
		part := b.String()
		if part == "" {
			return ""
		}
		out = append(out, part)
	}
	return strings.Join(out, "_")
}

func pluralizeSnake(s string) string {
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		prev := s[len(s)-2]
		if !strings.ContainsRune("aeiou", rune(prev)) {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
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
