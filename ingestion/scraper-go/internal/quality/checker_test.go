package quality

import (
	"testing"

	"ingestion/scraper-go/internal/model"
)

func TestCheckProductsFlagsRequiredQualityIssues(t *testing.T) {
	low := 500.0
	high := 1_500_000.0
	products := []model.RawProduct{
		{ID: 1, ScrapeRunID: 9, Source: "tokopedia", SourceProductID: "same", ProductURL: "https://example.com/same", PriceValue: &low},
		{ID: 2, ScrapeRunID: 9, Source: "tokopedia", SourceProductID: "same", ProductURL: "https://example.com/other", ProductName: "Beras", PriceValue: &high},
	}

	issues := CheckProducts(products)
	want := map[string]bool{
		"missing_product_name":     false,
		"missing_seller_location":  false,
		"suspicious_low_price":     false,
		"suspicious_high_price":    false,
		"duplicate_product_in_run": false,
	}
	for _, issue := range issues {
		if _, tracked := want[issue.IssueCode]; tracked {
			want[issue.IssueCode] = true
		}
	}
	for code, found := range want {
		if !found {
			t.Errorf("expected issue %s, got %#v", code, issues)
		}
	}
	if issues[0].ProductRawID == nil {
		t.Fatal("expected inserted product ID on product-level issue")
	}
}

func TestCheckProductsFlagsMissingPriceAndURL(t *testing.T) {
	issues := CheckProducts([]model.RawProduct{{ScrapeRunID: 10, Source: "shopee", ProductName: "Beras"}})
	assertIssueCodes(t, issues, "missing_price", "missing_product_url", "missing_seller_location")
}

func TestRunLevelQualityIssues(t *testing.T) {
	empty := EmptyActorResult(11, "shopee")
	if empty.IssueCode != "empty_actor_result" || empty.ProductRawID != nil {
		t.Fatalf("unexpected empty result issue: %#v", empty)
	}
	decode := AdapterDecodeFailed(12, "tokopedia", errTest("bad payload"))
	if decode.IssueCode != "adapter_decode_failed" || decode.RawValue != "bad payload" {
		t.Fatalf("unexpected decode issue: %#v", decode)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func assertIssueCodes(t *testing.T, issues []Issue, codes ...string) {
	t.Helper()
	found := make(map[string]bool)
	for _, issue := range issues {
		found[issue.IssueCode] = true
	}
	for _, code := range codes {
		if !found[code] {
			t.Errorf("expected issue %s, got %#v", code, issues)
		}
	}
}
