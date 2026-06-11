package adapter

import (
	"encoding/json"
	"testing"
)

func TestTokopediaToRawProducts(t *testing.T) {
	payload := []byte(`[
		{
			"type":"product",
			"id":5764025272,
			"url":"https://www.tokopedia.com/store/beras-premium-5kg",
			"source_context":{"scraped_time":"2026-06-11T08:00:00Z"},
			"product_core":{"product_id":"5764025272","product_title":"Beras Premium 5 kg"},
			"pricing_and_inventory":{"current_price":74500,"current_price_display":"Rp74.500","stock_value":"11"},
			"media_assets":{"primary_image_url":"https://example.com/rice.jpg"},
			"performance_and_flags":{"rating":4.8,"review_count":"6","sold_count":"29"},
			"seller_and_platform_context":{"shop_id":"1303898","shop_name":"Toko Pangan","shop_city":"Surabaya"},
			"search_listing_context":{"search_position":8,"listing_url":"https://www.tokopedia.com/store/beras-premium-5kg?src=search"}
		}
	]`)

	products, err := TokopediaToRawProducts(payload, 42, "beras 5 kg")
	if err != nil {
		t.Fatalf("adapt Tokopedia payload: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected one product, got %d", len(products))
	}
	product := products[0]
	if product.ScrapeRunID != 42 || product.Source != "tokopedia" {
		t.Fatalf("unexpected run identity: %#v", product)
	}
	if product.SourceProductID != "5764025272" || product.SourceShopID != "1303898" {
		t.Fatalf("unexpected source IDs: %#v", product)
	}
	if product.ProductName != "Beras Premium 5 kg" || product.PriceText != "Rp74.500" {
		t.Fatalf("unexpected product fields: %#v", product)
	}
	if product.ProductURL != "https://www.tokopedia.com/store/beras-premium-5kg" {
		t.Fatalf("expected canonical product URL, got %q", product.ProductURL)
	}
	if product.PriceValue == nil || *product.PriceValue != 74500 {
		t.Fatalf("unexpected price value: %#v", product.PriceValue)
	}
	if product.ResultRank == nil || *product.ResultRank != 8 {
		t.Fatalf("unexpected result rank: %#v", product.ResultRank)
	}
	if !json.Valid(product.RawPayload) || len(product.RawPayload) == 0 {
		t.Fatal("raw payload was not preserved")
	}
}

func TestTokopediaAdapterAllowsMissingOptionalFields(t *testing.T) {
	products, err := TokopediaToRawProducts([]byte(`[{"type":"product","product_core":{"product_title":"Beras"}}]`), 7, "beras")
	if err != nil {
		t.Fatalf("adapt sparse payload: %v", err)
	}
	if len(products) != 1 || products[0].PriceValue != nil || products[0].Rating != nil {
		t.Fatalf("unexpected sparse product: %#v", products)
	}
}

func TestTokopediaAdapterRejectsInvalidPayload(t *testing.T) {
	if _, err := TokopediaToRawProducts([]byte(`{"not":"an array"}`), 1, "beras"); err == nil {
		t.Fatal("expected decode error")
	}
}
