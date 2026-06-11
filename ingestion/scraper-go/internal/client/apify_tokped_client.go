package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

const tokopediaActorID = "jupri~tokopedia-scraper"

type apifyInput struct {
	Search   string `json:"search"`
	MaxItems int    `json:"maxItems"`
}

func RunTokopediaScraper(ctx context.Context, keyword string) ([]byte, error) {
	apiKey := os.Getenv("APIFY_TOKOPEDIA_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("APIFY_TOKOPEDIA_TOKEN is required")
	}

	endpoint := &url.URL{
		Scheme: "https",
		Host:   "api.apify.com",
		Path:   "/v2/acts/" + tokopediaActorID + "/run-sync-get-dataset-items",
	}
	query := endpoint.Query()
	query.Set("token", apiKey)
	endpoint.RawQuery = query.Encode()

	body, err := json.Marshal(apifyInput{Search: keyword, MaxItems: 10})
	if err != nil {
		return nil, fmt.Errorf("encode Apify input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Apify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	baseClient := &http.Client{Timeout: 2 * time.Minute}
	resp, err := NewRetryClient(baseClient, slog.Default()).Do(req)
	if err != nil {
		return nil, fmt.Errorf("run Tokopedia actor: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Tokopedia actor response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Tokopedia actor returned status %d: %s", resp.StatusCode, string(responseBody))
	}
	return responseBody, nil
}
