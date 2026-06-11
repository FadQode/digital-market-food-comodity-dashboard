package adapter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"ingestion/scraper-go/internal/model"
)

type shopeeRecord struct {
	ShopID      json.RawMessage `json:"shop_id"`
	ItemID      json.RawMessage `json:"item_id"`
	ProductName string          `json:"nama_produk"`
	Price       json.RawMessage `json:"harga"`
	Rating      json.RawMessage `json:"rating"`
	RatingCount json.RawMessage `json:"jumlah_rating"`
	Sold        json.RawMessage `json:"terjual"`
	Stock       json.RawMessage `json:"stok"`
	Location    string          `json:"lokasi"`
	URL         string          `json:"url"`
	ImageURL    string          `json:"gambar"`
	ShopName    string          `json:"nama_toko"`
	ScrapedAt   string          `json:"scraped_at"`
}

func ShopeeToRawProducts(payload []byte, scrapeRunID int64, query string) ([]model.RawProduct, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, fmt.Errorf("decode Shopee actor payload: %w", err)
	}

	now := time.Now()
	products := make([]model.RawProduct, 0, len(items))
	for index, raw := range items {
		var item shopeeRecord
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("decode Shopee item %d: %w", index+1, err)
		}
		price := rawFloat(item.Price)
		priceText := ""
		if price != nil {
			priceText = strconv.FormatFloat(*price, 'f', -1, 64)
		}

		products = append(products, model.RawProduct{
			ScrapeRunID:        scrapeRunID,
			Source:             "shopee",
			SourceProductID:    rawString(item.ItemID),
			SourceShopID:       rawString(item.ShopID),
			ProductURL:         item.URL,
			ImageURL:           item.ImageURL,
			ProductName:        item.ProductName,
			ShopName:           item.ShopName,
			SellerLocationText: item.Location,
			PriceText:          priceText,
			PriceValue:         price,
			Currency:           "IDR",
			Rating:             rawFloat(item.Rating),
			ReviewCount:        rawInt(item.RatingCount),
			SoldCount:          rawInt(item.Sold),
			StockText:          rawString(item.Stock),
			ResultRank:         rankPointer(index + 1),
			SourceQuery:        query,
			ScrapedAt:          parseTimeOrNow(item.ScrapedAt, now),
			RawPayload:         append(json.RawMessage(nil), raw...),
		})
	}
	return products, nil
}
