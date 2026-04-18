package service

import (
	"context"
	"strings"
	"testing"
)

func TestScheduledTaskService_runCommand_ShellDisabled(t *testing.T) {
	s := &ScheduledTaskService{shellCommandsEnabled: false}
	_, err := s.runCommand(context.Background(), "echo injected")
	if err == nil || !strings.Contains(err.Error(), "shell commands disabled") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestScheduledTaskService_runCommand_ShellEnabled(t *testing.T) {
	s := &ScheduledTaskService{shellCommandsEnabled: true}
	_, err := s.runCommand(context.Background(), "echo ok")
	if err != nil {
		// Windows echo semantics differ; only require no "disabled" error
		if strings.Contains(err.Error(), "shell commands disabled") {
			t.Fatal(err)
		}
	}
}
