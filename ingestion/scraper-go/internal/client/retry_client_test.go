package client

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRetryClientRetriesWithExponentialBackoff(t *testing.T) {
	attempts := 0
	client := newTestRetryClient(doerFunc(func(_ *http.Request) (*http.Response, error) {
		attempts++
		return response(http.StatusServiceUnavailable), nil
	}))

	var delays []time.Duration
	client.sleep = func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/products", nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("expected max retries error")
	}
	if attempts != 4 {
		t.Fatalf("expected 4 attempts, got %d", attempts)
	}
	want := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}
	if len(delays) != len(want) {
		t.Fatalf("expected %d delays, got %v", len(want), delays)
	}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delay %d: expected %s, got %s", i, want[i], delays[i])
		}
	}
}

func TestRetryClientHonorsRetryAfter(t *testing.T) {
	attempts := 0
	client := newTestRetryClient(doerFunc(func(_ *http.Request) (*http.Response, error) {
		attempts++
		resp := response(http.StatusOK)
		if attempts == 1 {
			resp = response(http.StatusTooManyRequests)
			resp.Header.Set("Retry-After", "7")
		}
		return resp, nil
	}))

	var delay time.Duration
	client.sleep = func(_ context.Context, value time.Duration) error {
		delay = value
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/products", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if attempts != 2 || delay != 7*time.Second {
		t.Fatalf("expected 2 attempts and 7s delay, got %d attempts and %s", attempts, delay)
	}
}

func TestRetryClientReplaysRequestBody(t *testing.T) {
	attempts := 0
	client := newTestRetryClient(doerFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if string(body) != `{"search":"beras"}` {
			t.Fatalf("attempt %d received body %q", attempts, body)
		}
		if attempts == 1 {
			return response(http.StatusInternalServerError), nil
		}
		return response(http.StatusOK), nil
	}))

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/products", strings.NewReader(`{"search":"beras"}`))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRedactedURLHidesAPIToken(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/run?token=secret&query=beras", nil)
	value := redactedURL(req)
	if strings.Contains(value, "secret") || !strings.Contains(value, "REDACTED") {
		t.Fatalf("token was not redacted: %s", value)
	}
}

func newTestRetryClient(doer httpDoer) *RetryClient {
	client := NewRetryClient(doer, slog.New(slog.NewTextHandler(io.Discard, nil)))
	client.minDelay = 0
	client.jitter = 0
	return client
}

func response(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("response")),
	}
}
