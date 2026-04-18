package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gin-scaffold/config"
	"gin-scaffold/pkg/httpclient"
)

// WebhookNotifier 将通知 JSON POST 到配置的 URL。
type WebhookNotifier struct {
	cfg config.WebhookNotifyConfig
}

// NewWebhookNotifier 构造 Webhook 通知器。
func NewWebhookNotifier(cfg config.WebhookNotifyConfig) *WebhookNotifier {
	return &WebhookNotifier{cfg: cfg}
}

type webhookBody struct {
	Channel string            `json:"channel"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Notify 发送 JSON；若配置 hmac_secret 则设置 X-Notify-Signature: sha256=<hex>。
func (n *WebhookNotifier) Notify(ctx context.Context, msg Message) error {
	if n == nil {
		return fmt.Errorf("notify: webhook: nil notifier")
	}
	url := strings.TrimSpace(n.cfg.URL)
	if url == "" {
		return fmt.Errorf("notify: webhook: empty url")
	}
	payload, err := json.Marshal(webhookBody{
		Channel: msg.Channel,
		Title:   msg.Title,
		Body:    msg.Body,
		Meta:    msg.Meta,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.cfg.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	if sec := strings.TrimSpace(n.cfg.HMACSecret); sec != "" {
		mac := hmac.New(sha256.New, []byte(sec))
		_, _ = mac.Write(payload)
		req.Header.Set("X-Notify-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := httpclient.Default().Do(req)
	if err != nil {
		return fmt.Errorf("notify: webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("notify: webhook: status %d body %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
