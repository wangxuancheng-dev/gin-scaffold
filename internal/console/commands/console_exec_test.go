package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"gin-scaffold/internal/console"
)

func TestConsoleExecute_Ping(t *testing.T) {
	var buf bytes.Buffer
	if err := console.Execute(context.Background(), "ping", nil, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "pong") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestConsoleExecute_Unknown(t *testing.T) {
	err := console.Execute(context.Background(), "no-such-command-xyz", nil, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConsoleList_NotEmpty(t *testing.T) {
	if len(console.List()) == 0 {
		t.Fatal("expected registered commands")
	}
}
