package scraper

import (
    "ingestion/scraper-go/internal/client"
)

func ScrapeTokopedia(keyword string) ([]byte, error) {
    return client.RunTokopediaScraper(keyword)
}