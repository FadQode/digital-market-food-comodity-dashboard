package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultApifyBaseURL = "https://api.apify.com"

type apifyRunOptions struct {
	MaxItems     int
	MaxChargeUSD float64
}

type apifyRunner struct {
	baseURL    string
	httpClient httpDoer
}

func newApifyRunner() *apifyRunner {
	return &apifyRunner{
		baseURL:    defaultApifyBaseURL,
		httpClient: &http.Client{Timeout: 3 * time.Minute},
	}
}

// runActor intentionally does not retry. Retrying this POST could start a
// second paid Actor run when the first response is lost or returns a 5xx.
func (r *apifyRunner) runActor(ctx context.Context, token, actorID string, input any, options apifyRunOptions) ([]byte, error) {
	if token == "" {
		return nil, fmt.Errorf("Apify token is required")
	}
	if options.MaxItems <= 0 {
		return nil, fmt.Errorf("Apify max items must be greater than zero")
	}
	if options.MaxChargeUSD <= 0 {
		return nil, fmt.Errorf("Apify max charge must be greater than zero")
	}

	actorID = strings.ReplaceAll(actorID, "/", "~")
	endpoint, err := url.Parse(strings.TrimRight(r.baseURL, "/") + "/v2/acts/" + url.PathEscape(actorID) + "/run-sync-get-dataset-items")
	if err != nil {
		return nil, fmt.Errorf("build Apify endpoint: %w", err)
	}
	query := endpoint.Query()
	query.Set("maxItems", strconv.Itoa(options.MaxItems))
	query.Set("maxTotalChargeUsd", strconv.FormatFloat(options.MaxChargeUSD, 'f', 2, 64))
	query.Set("timeout", "180")
	query.Set("restartOnError", "false")
	query.Set("format", "json")
	endpoint.RawQuery = query.Encode()

	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode Apify input: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Apify request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute Apify actor %s: %w", actorID, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("read Apify actor %s response: %w", actorID, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Apify actor %s returned status %d: %s", actorID, resp.StatusCode, truncate(responseBody, 1024))
	}
	if !json.Valid(responseBody) {
		return nil, fmt.Errorf("Apify actor %s returned invalid JSON", actorID)
	}
	return responseBody, nil
}

func truncate(value []byte, limit int) string {
	if len(value) <= limit {
		return string(value)
	}
	return string(value[:limit]) + "..."
}
