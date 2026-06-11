package scraper

import (
	"context"

	"ingestion/scraper-go/internal/client"
)

func ScrapeShopee(ctx context.Context, keyword string, maxItems int, maxChargeUSD float64) ([]byte, error) {
	return client.RunShopeeScraper(ctx, keyword, maxItems, maxChargeUSD)
}
