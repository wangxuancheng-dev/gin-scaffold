package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"

	"gin-scaffold/internal/job"
)

func TestWelcomeHandler_ProcessTask_badPayload(t *testing.T) {
	var h WelcomeHandler
	err := h.ProcessTask(context.Background(), asynq.NewTask(job.TypeWelcomeEmail, []byte("not-json")))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWelcomeHandler_ProcessTask_ok(t *testing.T) {
	var h WelcomeHandler
	payload, err := json.Marshal(job.WelcomePayload{UserID: 1, Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.ProcessTask(context.Background(), asynq.NewTask(job.TypeWelcomeEmail, payload)); err != nil {
		t.Fatal(err)
	}
}
