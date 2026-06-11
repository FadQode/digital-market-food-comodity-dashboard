package adapter

import (
	"encoding/json"
	"testing"
)

func TestShopeeToRawProducts(t *testing.T) {
	payload := []byte(`[
		{
			"shop_id":21874626,
			"item_id":25432764626,
			"nama_produk":"Beras Premium 5 kg",
			"harga":69300,
			"rating":4.92,
			"jumlah_rating":12,
			"terjual":48,
			"stok":1,
			"lokasi":"Kab. Bandung Barat",
			"url":"https://shopee.co.id/product-i.21874626.25432764626",
			"gambar":"https://example.com/rice.jpg"
		}
	]`)

	products, err := ShopeeToRawProducts(payload, 43, "beras 5 kg")
	if err != nil {
		t.Fatalf("adapt Shopee payload: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected one product, got %d", len(products))
	}
	product := products[0]
	if product.SourceProductID != "25432764626" || product.SourceShopID != "21874626" {
		t.Fatalf("unexpected source IDs: %#v", product)
	}
	if product.ProductName != "Beras Premium 5 kg" || product.PriceText != "69300" {
		t.Fatalf("unexpected common fields: %#v", product)
	}
	if product.PriceValue == nil || *product.PriceValue != 69300 {
		t.Fatalf("unexpected price value: %#v", product.PriceValue)
	}
	if product.ReviewCount == nil || *product.ReviewCount != 12 || product.SoldCount == nil || *product.SoldCount != 48 {
		t.Fatalf("unexpected performance fields: %#v", product)
	}
	if !json.Valid(product.RawPayload) {
		t.Fatal("raw payload was not preserved")
	}
}

func TestShopeeAdapterAllowsMissingOptionalFields(t *testing.T) {
	products, err := ShopeeToRawProducts([]byte(`[{"nama_produk":"Beras"}]`), 8, "beras")
	if err != nil {
		t.Fatalf("adapt sparse payload: %v", err)
	}
	if len(products) != 1 || products[0].PriceValue != nil || products[0].ProductURL != "" {
		t.Fatalf("unexpected sparse product: %#v", products)
	}
}

func TestShopeeAdapterRejectsInvalidPayload(t *testing.T) {
	if _, err := ShopeeToRawProducts([]byte(`not-json`), 1, "beras"); err == nil {
		t.Fatal("expected decode error")
	}
}
