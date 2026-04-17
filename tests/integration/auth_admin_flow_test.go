//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type integrationEnv struct {
	BaseURL  string
	Username string
	Password string
}

func TestIntegration_AdminAuthFlow(t *testing.T) {
	env, ok := loadIntegrationEnv(t)
	if !ok {
		return
	}
	client := &http.Client{Timeout: 20 * time.Second}
	token := loginAndGetAccessToken(t, client, env.BaseURL, env.Username, env.Password)
	resp := doAuthedRequest(t, client, http.MethodGet, env.BaseURL+"/api/v1/admin/users?page=1&page_size=5", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_UserExportAsyncFlow(t *testing.T) {
	env, ok := loadIntegrationEnv(t)
	if !ok {
		return
	}
	client := &http.Client{Timeout: 20 * time.Second}
	token := loginAndGetAccessToken(t, client, env.BaseURL, env.Username, env.Password)

	createURL := env.BaseURL + "/api/v1/admin/users/export/tasks?fields=id,username"
	resp := doAuthedRequest(t, client, http.MethodPost, createURL, token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create user export task expected 200, got %d", resp.StatusCode)
	}
	envBody := decodeEnvelope(t, resp)
	taskID := extractStringField(t, envBody.Data, "task_id")
	if taskID == "" {
		t.Fatal("empty task_id for user export")
	}

	statusURL := env.BaseURL + "/api/v1/admin/users/export/tasks/" + url.PathEscape(taskID)
	waitTaskSuccess(t, client, statusURL, token, 45*time.Second)
	downloadURL := env.BaseURL + "/api/v1/admin/users/export/tasks/" + url.PathEscape(taskID) + "/download"
	downloadResp := doAuthedRequest(t, client, http.MethodGet, downloadURL, token, nil)
	if downloadResp.StatusCode != http.StatusOK {
		t.Fatalf("download user export expected 200, got %d", downloadResp.StatusCode)
	}
	body, _ := io.ReadAll(downloadResp.Body)
	if len(body) == 0 || !strings.Contains(string(body), "id,username") {
		t.Fatalf("unexpected user export csv content: %q", string(body))
	}
}

func TestIntegration_AuditExportAsyncFlow(t *testing.T) {
	env, ok := loadIntegrationEnv(t)
	if !ok {
		return
	}
	client := &http.Client{Timeout: 20 * time.Second}
	token := loginAndGetAccessToken(t, client, env.BaseURL, env.Username, env.Password)

	createURL := env.BaseURL + "/api/v1/admin/audit-logs/export/tasks"
	resp := doAuthedRequest(t, client, http.MethodPost, createURL, token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create audit export task expected 200, got %d", resp.StatusCode)
	}
	envBody := decodeEnvelope(t, resp)
	taskID := extractStringField(t, envBody.Data, "task_id")
	if taskID == "" {
		t.Fatal("empty task_id for audit export")
	}

	statusURL := env.BaseURL + "/api/v1/admin/audit-logs/export/tasks/" + url.PathEscape(taskID)
	waitTaskSuccess(t, client, statusURL, token, 45*time.Second)
	downloadURL := env.BaseURL + "/api/v1/admin/audit-logs/export/tasks/" + url.PathEscape(taskID) + "/download"
	downloadResp := doAuthedRequest(t, client, http.MethodGet, downloadURL, token, nil)
	if downloadResp.StatusCode != http.StatusOK {
		t.Fatalf("download audit export expected 200, got %d", downloadResp.StatusCode)
	}
	body, _ := io.ReadAll(downloadResp.Body)
	if len(body) == 0 || !strings.Contains(string(body), "request_id") {
		t.Fatalf("unexpected audit export csv content: %q", string(body))
	}
}

func loadIntegrationEnv(t *testing.T) (integrationEnv, bool) {
	baseURL := strings.TrimSpace(os.Getenv("INTEGRATION_BASE_URL"))
	username := strings.TrimSpace(os.Getenv("INTEGRATION_ADMIN_USERNAME"))
	password := strings.TrimSpace(os.Getenv("INTEGRATION_ADMIN_PASSWORD"))
	if baseURL == "" || username == "" || password == "" {
		t.Skip("set INTEGRATION_BASE_URL, INTEGRATION_ADMIN_USERNAME, INTEGRATION_ADMIN_PASSWORD to run integration tests")
		return integrationEnv{}, false
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return integrationEnv{BaseURL: baseURL, Username: username, Password: password}, true
}

func doAuthedRequest(t *testing.T, client *http.Client, method, requestURL, token string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, requestURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeEnvelope(t *testing.T, resp *http.Response) envelope {
	t.Helper()
	defer resp.Body.Close()
	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Code != 200 {
		t.Fatalf("unexpected business code: %d msg=%s", env.Code, env.Msg)
	}
	return env
}

func extractStringField(t *testing.T, raw json.RawMessage, field string) string {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data object: %v", err)
	}
	v, _ := data[field].(string)
	return strings.TrimSpace(v)
}

func waitTaskSuccess(t *testing.T, client *http.Client, statusURL, token string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := doAuthedRequest(t, client, http.MethodGet, statusURL, token, nil)
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf("status request failed: status=%d body=%s", resp.StatusCode, string(b))
		}
		env := decodeEnvelope(t, resp)
		var data map[string]any
		if err := json.Unmarshal(env.Data, &data); err != nil {
			t.Fatalf("decode status data: %v", err)
		}
		state, _ := data["state"].(string)
		isReady, _ := data["is_ready"].(bool)
		if isReady || state == "success" {
			return
		}
		if state == "failed" {
			t.Fatalf("task failed: %+v", data)
		}
		time.Sleep(1500 * time.Millisecond)
	}
	t.Fatalf("task not ready before timeout: %s", fmt.Sprintf("%s (timeout=%s)", statusURL, timeout))
}

func loginAndGetAccessToken(t *testing.T, client *http.Client, baseURL, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := client.Post(baseURL+"/api/v1/client/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d", resp.StatusCode)
	}
	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode login data: %v", err)
	}
	if strings.TrimSpace(data.AccessToken) == "" {
		t.Fatal("empty access_token")
	}
	return data.AccessToken
}
