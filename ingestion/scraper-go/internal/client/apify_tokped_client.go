package client

import (
	"context"
	"fmt"
	"os"
)

const defaultTokopediaActorID = "fatihtahta~tokopedia-scraper"

type tokopediaInput struct {
	Queries            []string                    `json:"queries"`
	Limit              int                         `json:"limit"`
	ProxyConfiguration tokopediaProxyConfiguration `json:"proxyConfiguration"`
}

type tokopediaProxyConfiguration struct {
	UseApifyProxy bool `json:"useApifyProxy"`
}

func RunTokopediaScraper(ctx context.Context, keyword string, maxItems int, maxChargeUSD float64) ([]byte, error) {
	apiKey := os.Getenv("TOKOPEDIA_APIFY_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TOKOPEDIA_APIFY_KEY is required")
	}
	actorID := envOrDefault("TOKOPEDIA_APIFY_ACTOR", defaultTokopediaActorID)
	input := tokopediaInput{
		Queries: []string{keyword},
		Limit:   maxItems,
		ProxyConfiguration: tokopediaProxyConfiguration{
			UseApifyProxy: true,
		},
	}
	return newApifyRunner().runActor(ctx, apiKey, actorID, input, apifyRunOptions{
		MaxItems:     maxItems,
		MaxChargeUSD: maxChargeUSD,
	})
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
