package console

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

type Command interface {
	Name() string
	Description() string
	Run(ctx context.Context, args []string, out io.Writer) error
}

var (
	mu       sync.RWMutex
	commands = map[string]Command{}
)

func Register(cmd Command) {
	if cmd == nil {
		return
	}
	name := strings.TrimSpace(cmd.Name())
	if name == "" {
		return
	}
	mu.Lock()
	commands[name] = cmd
	mu.Unlock()
}

func Execute(ctx context.Context, name string, args []string, out io.Writer) error {
	mu.RLock()
	cmd, ok := commands[name]
	mu.RUnlock()
	if !ok {
		return fmt.Errorf("artisan command not found: %s", name)
	}
	return cmd.Run(ctx, args, out)
}

func List() []Command {
	mu.RLock()
	items := make([]Command, 0, len(commands))
	for _, cmd := range commands {
		items = append(items, cmd)
	}
	mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name() < items[j].Name()
	})
	return items
}
