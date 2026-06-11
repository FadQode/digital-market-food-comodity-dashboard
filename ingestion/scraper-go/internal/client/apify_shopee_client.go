package client

import (
	"context"
	"fmt"
	"os"
)

const defaultShopeeActorID = "pumpkin_jingo~shopee-scraper-id"

type shopeeInput struct {
	Mode        string  `json:"mode"`
	Keyword     string  `json:"keyword"`
	MaxProducts int     `json:"maxProducts"`
	Sort        string  `json:"sort"`
	Delay       float64 `json:"delay"`
}

func RunShopeeScraper(ctx context.Context, keyword string, maxItems int, maxChargeUSD float64) ([]byte, error) {
	apiKey := os.Getenv("SHOPEE_APIFY_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SHOPEE_APIFY_KEY is required")
	}
	actorID := envOrDefault("SHOPEE_APIFY_ACTOR", defaultShopeeActorID)
	input := shopeeInput{
		Mode:        "keyword",
		Keyword:     keyword,
		MaxProducts: maxItems,
		Sort:        "relevancy",
		Delay:       1.5,
	}
	return newApifyRunner().runActor(ctx, apiKey, actorID, input, apifyRunOptions{
		MaxItems:     maxItems,
		MaxChargeUSD: maxChargeUSD,
	})
}
