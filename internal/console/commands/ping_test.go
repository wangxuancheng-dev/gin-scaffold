package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPingCommand_Run(t *testing.T) {
	var c pingCommand
	var buf bytes.Buffer
	if err := c.Run(context.Background(), nil, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "pong") {
		t.Fatalf("got %q", buf.String())
	}
}
