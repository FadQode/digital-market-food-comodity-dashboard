package model

import (
	"encoding/json"
	"time"
)

type RawProduct struct {
	ID                 int64
	ScrapeRunID        int64
	Source             string
	SourceProductID    string
	SourceShopID       string
	ProductURL         string
	ImageURL           string
	ProductName        string
	ShopName           string
	SellerLocationText string
	PriceText          string
	PriceValue         *float64
	Currency           string
	Rating             *float64
	ReviewCount        *int
	SoldCount          *int
	StockText          string
	ResultRank         *int
	SourceQuery        string
	ScrapedAt          time.Time
	RawPayload         json.RawMessage
}
