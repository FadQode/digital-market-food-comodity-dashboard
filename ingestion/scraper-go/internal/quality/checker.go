package quality

import (
	"strconv"
	"strings"

	"ingestion/scraper-go/internal/model"
)

const (
	suspiciousLowPrice  = 1_000.0
	suspiciousHighPrice = 1_000_000.0
)

func CheckProducts(products []model.RawProduct) []Issue {
	issues := make([]Issue, 0)
	seenProductIDs := make(map[string]struct{})
	seenURLs := make(map[string]struct{})

	for index := range products {
		product := &products[index]
		productID := productIDPointer(product)
		metadata := map[string]any{"result_rank": product.ResultRank}

		if strings.TrimSpace(product.ProductName) == "" {
			issues = append(issues, productIssue(product, productID, "error", "missing_product_name", "product name is empty", "product_name", "", metadata))
		}
		if strings.TrimSpace(product.PriceText) == "" && product.PriceValue == nil {
			issues = append(issues, productIssue(product, productID, "warning", "missing_price", "price text and numeric price are missing", "price", "", metadata))
		}
		if strings.TrimSpace(product.ProductURL) == "" {
			issues = append(issues, productIssue(product, productID, "warning", "missing_product_url", "product URL is empty", "product_url", "", metadata))
		}
		if strings.TrimSpace(product.SellerLocationText) == "" {
			issues = append(issues, productIssue(product, productID, "info", "missing_seller_location", "seller location is unavailable", "seller_location_text", "", metadata))
		}
		if product.PriceValue != nil && *product.PriceValue < suspiciousLowPrice {
			issues = append(issues, productIssue(product, productID, "warning", "suspicious_low_price", "raw price is below the temporary IDR floor", "price_value", formatFloat(*product.PriceValue), metadata))
		}
		if product.PriceValue != nil && *product.PriceValue > suspiciousHighPrice {
			issues = append(issues, productIssue(product, productID, "warning", "suspicious_high_price", "raw price is above the temporary IDR ceiling", "price_value", formatFloat(*product.PriceValue), metadata))
		}

		duplicateField, duplicateValue := duplicateKey(product, seenProductIDs, seenURLs)
		if duplicateField != "" {
			issues = append(issues, productIssue(product, productID, "warning", "duplicate_product_in_run", "product appears more than once in the actor result", duplicateField, duplicateValue, metadata))
		}
	}
	return issues
}

func EmptyActorResult(scrapeRunID int64, source string) Issue {
	return Issue{
		ScrapeRunID: scrapeRunID,
		Source:      source,
		Severity:    "error",
		IssueCode:   "empty_actor_result",
		Message:     "actor returned zero product items",
		Metadata:    map[string]any{},
	}
}

func AdapterDecodeFailed(scrapeRunID int64, source string, err error) Issue {
	return Issue{
		ScrapeRunID: scrapeRunID,
		Source:      source,
		Severity:    "error",
		IssueCode:   "adapter_decode_failed",
		Message:     "actor payload could not be decoded",
		RawValue:    err.Error(),
		Metadata:    map[string]any{},
	}
}

func RawProductInsertFailed(product model.RawProduct) Issue {
	return Issue{
		ScrapeRunID: product.ScrapeRunID,
		Source:      product.Source,
		Severity:    "error",
		IssueCode:   "raw_product_insert_failed",
		Message:     "raw product row was not inserted; it may duplicate another item in this run",
		FieldName:   "source_product_id",
		RawValue:    product.SourceProductID,
		Metadata: map[string]any{
			"product_url": product.ProductURL,
			"result_rank": product.ResultRank,
		},
	}
}

func productIssue(product *model.RawProduct, productID *int64, severity, code, message, field, rawValue string, metadata map[string]any) Issue {
	return Issue{
		ScrapeRunID:  product.ScrapeRunID,
		ProductRawID: productID,
		Source:       product.Source,
		Severity:     severity,
		IssueCode:    code,
		Message:      message,
		FieldName:    field,
		RawValue:     rawValue,
		Metadata:     metadata,
	}
}

func productIDPointer(product *model.RawProduct) *int64 {
	if product.ID == 0 {
		return nil
	}
	return &product.ID
}

func duplicateKey(product *model.RawProduct, productIDs, urls map[string]struct{}) (string, string) {
	if value := strings.TrimSpace(product.SourceProductID); value != "" {
		key := product.Source + "\x00" + value
		if _, exists := productIDs[key]; exists {
			return "source_product_id", value
		}
		productIDs[key] = struct{}{}
	}
	if value := strings.TrimSpace(product.ProductURL); value != "" {
		key := product.Source + "\x00" + value
		if _, exists := urls[key]; exists {
			return "product_url", value
		}
		urls[key] = struct{}{}
	}
	return "", ""
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
