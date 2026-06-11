package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestApifyRunnerAppliesCostLimitsAndAuthorization(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		if got := r.URL.Query().Get("maxItems"); got != "20" {
			t.Fatalf("expected maxItems=20, got %q", got)
		}
		if got := r.URL.Query().Get("maxTotalChargeUsd"); got != "0.07" {
			t.Fatalf("expected maxTotalChargeUsd=0.07, got %q", got)
		}
		if got := r.URL.Query().Get("restartOnError"); got != "false" {
			t.Fatalf("expected restartOnError=false, got %q", got)
		}

		var input map[string]any
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("decode input: %v", err)
		}
		if input["keyword"] != "beras 5 kg" {
			t.Fatalf("unexpected input: %#v", input)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"Beras 5 kg"}]`))
	}))
	defer server.Close()

	runner := &apifyRunner{baseURL: server.URL, httpClient: server.Client()}
	data, err := runner.runActor(context.Background(), "secret-token", "owner/actor", map[string]any{
		"keyword": "beras 5 kg",
	}, apifyRunOptions{MaxItems: 20, MaxChargeUSD: 0.07})
	if err != nil {
		t.Fatalf("run actor: %v", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("expected one paid request, got %d", requests.Load())
	}
	if !json.Valid(data) {
		t.Fatal("expected valid JSON response")
	}
}

func TestApifyRunnerDoesNotRetryServerErrors(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.Error(w, "temporary failure", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	runner := &apifyRunner{baseURL: server.URL, httpClient: server.Client()}
	_, err := runner.runActor(context.Background(), "secret-token", "owner~actor", map[string]any{}, apifyRunOptions{
		MaxItems:     20,
		MaxChargeUSD: 0.07,
	})
	if err == nil {
		t.Fatal("expected server error")
	}
	if requests.Load() != 1 {
		t.Fatalf("paid POST must not be retried; got %d requests", requests.Load())
	}
}

func TestMarketplaceActorInputsUseHardItemLimits(t *testing.T) {
	tokopedia, err := json.Marshal(tokopediaInput{
		Queries: []string{"beras 5 kg"},
		Limit:   20,
		ProxyConfiguration: tokopediaProxyConfiguration{
			UseApifyProxy: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	shopee, err := json.Marshal(shopeeInput{
		Mode: "keyword", Keyword: "beras 5 kg", MaxProducts: 20, Sort: "relevancy", Delay: 1.5,
	})
	if err != nil {
		t.Fatal(err)
	}

	for name, payload := range map[string][]byte{"tokopedia": tokopedia, "shopee": shopee} {
		if !json.Valid(payload) {
			t.Fatalf("%s input is invalid JSON", name)
		}
		var value map[string]any
		if err := json.Unmarshal(payload, &value); err != nil {
			t.Fatal(err)
		}
		limit := value["limit"]
		if name == "shopee" {
			limit = value["maxProducts"]
		}
		if limit != float64(20) {
			t.Fatalf("%s input has unexpected item limit %#v", name, limit)
		}
	}
}
