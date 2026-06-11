package storage

import (
	"encoding/json"
	"testing"
	"time"

	"ingestion/scraper-go/internal/model"
)

func TestRawProductArgsPreserveStorageContract(t *testing.T) {
	price := 74500.0
	rank := 3
	product := model.RawProduct{
		ScrapeRunID:     42,
		Source:          "tokopedia",
		SourceProductID: "product-1",
		ProductName:     "Beras Premium 5 kg",
		PriceValue:      &price,
		Currency:        "IDR",
		ResultRank:      &rank,
		RawPayload:      json.RawMessage(`{"id":"product-1"}`),
		ScrapedAt:       time.Date(2026, 6, 11, 8, 0, 0, 0, time.UTC),
	}

	args := rawProductArgs(product)
	if len(args) != 20 {
		t.Fatalf("expected 20 insert arguments, got %d", len(args))
	}
	if args[0] != int64(42) || args[1] != "tokopedia" || args[2] != "product-1" {
		t.Fatalf("unexpected identity arguments: %#v", args[:3])
	}
	if args[6] != "Beras Premium 5 kg" || args[19] == nil {
		t.Fatalf("raw product contract lost fields: %#v", args)
	}
}
