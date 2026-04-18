package notify

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-scaffold/config"
)

func TestChain_Notify_joinsErrors(t *testing.T) {
	e1 := errors.New("a")
	e2 := errors.New("b")
	chain := Chain{
		NotifierFunc(func(context.Context, Message) error { return e1 }),
		NotifierFunc(func(context.Context, Message) error { return e2 }),
	}
	err := chain.Notify(context.Background(), Message{Channel: "c"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("got %v", err)
	}
}

func TestChain_Notify_skipsNil(t *testing.T) {
	var chain Chain
	chain = append(chain, nil, Noop{})
	if err := chain.Notify(context.Background(), Message{}); err != nil {
		t.Fatal(err)
	}
}

func TestSetDefault_nilRestoresLogNotifier(t *testing.T) {
	SetDefault(Noop{})
	SetDefault(nil)
	if _, ok := Default().(LogNotifier); !ok {
		t.Fatalf("got %T", Default())
	}
}

func TestWebhookNotifier_nil(t *testing.T) {
	var n *WebhookNotifier
	if err := n.Notify(context.Background(), Message{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestWebhookNotifier_emptyURL(t *testing.T) {
	n := NewWebhookNotifier(config.WebhookNotifyConfig{})
	if err := n.Notify(context.Background(), Message{Channel: "x"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestWebhookNotifier_ok(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	n := NewWebhookNotifier(config.WebhookNotifyConfig{URL: srv.URL})
	if err := n.Notify(context.Background(), Message{Channel: "c", Title: "t", Body: "b"}); err != nil {
		t.Fatal(err)
	}
}

func TestSMTPNotifier_nil(t *testing.T) {
	var n *SMTPNotifier
	if err := n.Notify(context.Background(), Message{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestSMTPNotifier_missingRecipient(t *testing.T) {
	n := NewSMTPNotifier(config.SMTPNotifyConfig{Host: "127.0.0.1", Port: 25, From: "a@b.c"})
	if err := n.Notify(context.Background(), Message{Title: "x", Body: "y"}); err == nil {
		t.Fatal("expected error")
	}
}

// NotifierFunc adapts a function to Notifier (test helper).
type NotifierFunc func(context.Context, Message) error

func (f NotifierFunc) Notify(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}
