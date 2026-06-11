package scraper

import (
	"context"

	"ingestion/scraper-go/internal/client"
)

func ScrapeTokopedia(ctx context.Context, keyword string) ([]byte, error) {
	return client.RunTokopediaScraper(ctx, keyword)
}
