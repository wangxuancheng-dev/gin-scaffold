package commands

import (
	"context"
	"fmt"
	"io"
	"time"

	"gin-scaffold/internal/console"
)

type pingCommand struct{}

func (c *pingCommand) Name() string { return "ping" }

func (c *pingCommand) Description() string { return "simple health check command" }

func (c *pingCommand) Run(ctx context.Context, args []string, out io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, _ = fmt.Fprintf(out, "pong %s\n", time.Now().Format(time.RFC3339))
		return nil
	}
}

func init() {
	console.Register(&pingCommand{})
}
