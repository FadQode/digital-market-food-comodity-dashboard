package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultMinDelay   = 1500 * time.Millisecond
	defaultBaseDelay  = 2 * time.Second
	defaultJitter     = 500 * time.Millisecond
	defaultMaxRetries = 3
)

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// RetryClient applies marketplace-friendly request spacing and bounded retries.
type RetryClient struct {
	client     httpDoer
	logger     *slog.Logger
	minDelay   time.Duration
	baseDelay  time.Duration
	jitter     time.Duration
	maxRetries int

	mu          sync.Mutex
	lastRequest time.Time
	rand        *rand.Rand
	now         func() time.Time
	sleep       func(context.Context, time.Duration) error
}

func NewRetryClient(client httpDoer, logger *slog.Logger) *RetryClient {
	if logger == nil {
		logger = slog.Default()
	}

	return &RetryClient{
		client:     client,
		logger:     logger,
		minDelay:   defaultMinDelay,
		baseDelay:  defaultBaseDelay,
		jitter:     defaultJitter,
		maxRetries: defaultMaxRetries,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		now:        time.Now,
		sleep:      sleepContext,
	}
}

func (c *RetryClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.waitForRequestSlot(req.Context()); err != nil {
			return nil, err
		}

		attemptReq, err := cloneRequest(req, attempt)
		if err != nil {
			return nil, err
		}

		resp, err := c.client.Do(attemptReq)
		if err == nil && !shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		lastErr = err
		if attempt == c.maxRetries {
			if resp != nil {
				drainAndClose(resp.Body)
				lastErr = fmt.Errorf("HTTP status %d", resp.StatusCode)
			}
			break
		}

		delay := c.backoffDelay(attempt, resp)
		status := 0
		if resp != nil {
			status = resp.StatusCode
			drainAndClose(resp.Body)
		}

		attrs := []any{
			"event", "retry",
			"attempt", attempt + 1,
			"url", redactedURL(req),
			"status", status,
			"delay_ms", delay.Milliseconds(),
		}
		if err != nil {
			attrs = append(attrs, "reason", err.Error())
		}
		c.logger.WarnContext(req.Context(), "request retry", attrs...)

		if err := c.sleep(req.Context(), delay); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded for %s: %w", redactedURL(req), lastErr)
}

func redactedURL(req *http.Request) string {
	cloned := *req.URL
	query := cloned.Query()
	for _, key := range []string{"token", "api_key", "apikey", "key"} {
		if query.Has(key) {
			query.Set(key, "REDACTED")
		}
	}
	cloned.RawQuery = query.Encode()
	return cloned.Redacted()
}

func (c *RetryClient) waitForRequestSlot(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	wait := c.lastRequest.Add(c.jittered(c.minDelay)).Sub(now)
	if wait > 0 {
		if err := c.sleep(ctx, wait); err != nil {
			return err
		}
		now = c.now()
	}
	c.lastRequest = now
	return nil
}

func (c *RetryClient) backoffDelay(attempt int, resp *http.Response) time.Duration {
	delay := c.baseDelay * time.Duration(1<<attempt)
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if retryAfter, ok := parseRetryAfter(resp.Header.Get("Retry-After"), c.now()); ok {
			delay = retryAfter
		}
	}
	return c.jittered(delay)
}

func (c *RetryClient) jittered(delay time.Duration) time.Duration {
	if c.jitter <= 0 {
		return delay
	}
	offset := time.Duration(c.rand.Int63n(int64(2*c.jitter)+1)) - c.jitter
	if delay+offset < 0 {
		return 0
	}
	return delay + offset
}

func shouldRetry(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	delay := when.Sub(now)
	if delay < 0 {
		delay = 0
	}
	return delay, true
}

func cloneRequest(req *http.Request, attempt int) (*http.Request, error) {
	cloned := req.Clone(req.Context())
	if attempt == 0 || req.Body == nil {
		return cloned, nil
	}
	if req.GetBody == nil {
		return nil, fmt.Errorf("request body cannot be replayed for retry")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("recreate request body: %w", err)
	}
	cloned.Body = body
	return cloned, nil
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 4096))
	_ = body.Close()
}
