package adapter

import (
	"encoding/json"
	"fmt"
	"time"

	"ingestion/scraper-go/internal/model"
)

type tokopediaRecord struct {
	Type          string          `json:"type"`
	ID            json.RawMessage `json:"id"`
	URL           string          `json:"url"`
	SourceContext struct {
		ScrapedTime string `json:"scraped_time"`
		URL         string `json:"url"`
	} `json:"source_context"`
	ProductCore struct {
		ProductID    json.RawMessage `json:"product_id"`
		ProductTitle string          `json:"product_title"`
	} `json:"product_core"`
	PricingAndInventory struct {
		CurrentPrice        json.RawMessage `json:"current_price"`
		CurrentPriceDisplay string          `json:"current_price_display"`
		StockValue          json.RawMessage `json:"stock_value"`
		StockText           string          `json:"stock_text"`
	} `json:"pricing_and_inventory"`
	MediaAssets struct {
		PrimaryImageURL string `json:"primary_image_url"`
	} `json:"media_assets"`
	PerformanceAndFlags struct {
		Rating      json.RawMessage `json:"rating"`
		ReviewCount json.RawMessage `json:"review_count"`
		SoldCount   json.RawMessage `json:"sold_count"`
	} `json:"performance_and_flags"`
	SellerAndPlatformContext struct {
		ShopID   json.RawMessage `json:"shop_id"`
		ShopName string          `json:"shop_name"`
		ShopCity string          `json:"shop_city"`
	} `json:"seller_and_platform_context"`
	SearchListingContext struct {
		SearchPosition int    `json:"search_position"`
		ListingURL     string `json:"listing_url"`
	} `json:"search_listing_context"`
}

func TokopediaToRawProducts(payload []byte, scrapeRunID int64, query string) ([]model.RawProduct, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, fmt.Errorf("decode Tokopedia actor payload: %w", err)
	}

	now := time.Now()
	products := make([]model.RawProduct, 0, len(items))
	for index, raw := range items {
		var item tokopediaRecord
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("decode Tokopedia item %d: %w", index+1, err)
		}
		if item.Type != "" && item.Type != "product" {
			continue
		}

		productID := rawString(item.ProductCore.ProductID)
		if productID == "" {
			productID = rawString(item.ID)
		}
		productURL := item.URL
		if productURL == "" {
			productURL = item.SourceContext.URL
		}
		if productURL == "" {
			productURL = item.SearchListingContext.ListingURL
		}
		rank := item.SearchListingContext.SearchPosition
		if rank <= 0 {
			rank = index + 1
		}
		stockText := item.PricingAndInventory.StockText
		if stockText == "" {
			stockText = rawString(item.PricingAndInventory.StockValue)
		}

		products = append(products, model.RawProduct{
			ScrapeRunID:        scrapeRunID,
			Source:             "tokopedia",
			SourceProductID:    productID,
			SourceShopID:       rawString(item.SellerAndPlatformContext.ShopID),
			ProductURL:         productURL,
			ImageURL:           item.MediaAssets.PrimaryImageURL,
			ProductName:        item.ProductCore.ProductTitle,
			ShopName:           item.SellerAndPlatformContext.ShopName,
			SellerLocationText: item.SellerAndPlatformContext.ShopCity,
			PriceText:          item.PricingAndInventory.CurrentPriceDisplay,
			PriceValue:         rawFloat(item.PricingAndInventory.CurrentPrice),
			Currency:           "IDR",
			Rating:             rawFloat(item.PerformanceAndFlags.Rating),
			ReviewCount:        rawInt(item.PerformanceAndFlags.ReviewCount),
			SoldCount:          rawInt(item.PerformanceAndFlags.SoldCount),
			StockText:          stockText,
			ResultRank:         rankPointer(rank),
			SourceQuery:        query,
			ScrapedAt:          parseTimeOrNow(item.SourceContext.ScrapedTime, now),
			RawPayload:         append(json.RawMessage(nil), raw...),
		})
	}
	return products, nil
}
