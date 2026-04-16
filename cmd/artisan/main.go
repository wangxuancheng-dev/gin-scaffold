package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"gin-scaffold/internal/console"
	_ "gin-scaffold/internal/console/commands"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "artisan",
		Short: "application console commands",
	}
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newMakeCommandCommand())
	rootCmd.AddCommand(newRunCommand())
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
