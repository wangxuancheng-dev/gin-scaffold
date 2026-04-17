package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

// Config 定义下游 HTTP 客户端治理参数。
type Config struct {
	Timeout          time.Duration
	RetryMax         int
	RetryBackoff     time.Duration
	CircuitThreshold uint32
	CircuitOpen      time.Duration
}

// Client 统一封装超时、重试、熔断。
type Client struct {
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker
	retryMax   int
	backoff    time.Duration
}

// New 创建治理客户端，零值会被安全默认值覆盖。
func New(cfg Config) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 200 * time.Millisecond
	}
	if cfg.CircuitThreshold == 0 {
		cfg.CircuitThreshold = 10
	}
	if cfg.CircuitOpen <= 0 {
		cfg.CircuitOpen = 30 * time.Second
	}
	if cfg.RetryMax < 0 {
		cfg.RetryMax = 0
	}
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "outbound-http",
		Timeout: cfg.CircuitOpen,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.CircuitThreshold
		},
	})
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		breaker:    cb,
		retryMax:   cfg.RetryMax,
		backoff:    cfg.RetryBackoff,
	}
}

// Do 执行请求（含重试与熔断）。
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c == nil || req == nil {
		return nil, fmt.Errorf("httpclient: nil client or request")
	}
	attempts := c.retryMax + 1
	var lastErr error
	for i := 0; i < attempts; i++ {
		resp, err := c.executeOnce(req)
		if err == nil && !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}
		if err == nil {
			lastErr = fmt.Errorf("retryable status: %d", resp.StatusCode)
			_ = resp.Body.Close()
		} else {
			lastErr = err
		}
		if i == attempts-1 || !isRetryableError(lastErr) {
			break
		}
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(c.backoff):
		}
	}
	return nil, lastErr
}

func (c *Client) executeOnce(req *http.Request) (*http.Response, error) {
	body, err := snapshotBody(req)
	if err != nil {
		return nil, err
	}
	cloned := req.Clone(req.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(body))
	cloned.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.httpClient.Do(cloned)
	})
	if err != nil {
		return nil, err
	}
	resp, ok := result.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("httpclient: invalid response type")
	}
	return resp, nil
}

func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return true
}

func snapshotBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	if req.GetBody != nil {
		rc, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		buf, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	_ = req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(buf))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buf)), nil }
	return buf, nil
}
