package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"ingestion/scraper-go/internal/model"
)

const insertRawProductSQL = `
	INSERT INTO products_raw (
		scrape_run_id, source,
		external_product_id, source_product_id, source_shop_id,
		product_url, image_url,
		product_title, product_name,
		seller_name, shop_name, seller_location_text,
		price_text, price_amount, price_value, currency,
		rating, review_count, sold_count, stock_text,
		result_rank, source_query, scraped_at, raw_payload
	) VALUES (
		$1, $2,
		NULLIF($3, ''), NULLIF($3, ''), NULLIF($4, ''),
		NULLIF($5, ''), NULLIF($6, ''),
		NULLIF($7, ''), NULLIF($7, ''),
		NULLIF($8, ''), NULLIF($8, ''), NULLIF($9, ''),
		NULLIF($10, ''), $11, $11, COALESCE(NULLIF($12, ''), 'IDR'),
		$13, $14, $15, NULLIF($16, ''),
		$17, NULLIF($18, ''), $19, $20
	)
	ON CONFLICT DO NOTHING
	RETURNING id
`

func (r *Repository) InsertRawProducts(ctx context.Context, products []model.RawProduct) (inserted int, failed int, err error) {
	if len(products) == 0 {
		return 0, 0, nil
	}
	defer func() {
		if err != nil {
			for index := range products {
				products[index].ID = 0
			}
		}
	}()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, len(products), fmt.Errorf("begin raw product transaction: %w", err)
	}
	defer tx.Rollback()

	for index := range products {
		product := &products[index]
		if product.ScrapedAt.IsZero() {
			product.ScrapedAt = time.Now().UTC()
		}
		savepoint := fmt.Sprintf("raw_product_%d", index)
		if _, err := tx.ExecContext(ctx, "SAVEPOINT "+savepoint); err != nil {
			return inserted, len(products) - inserted, fmt.Errorf("create product savepoint: %w", err)
		}

		err := tx.QueryRowContext(ctx, insertRawProductSQL, rawProductArgs(*product)...).Scan(&product.ID)
		switch {
		case err == nil:
			inserted++
		case errors.Is(err, sql.ErrNoRows):
			failed++
		case err != nil:
			failed++
			if _, rollbackErr := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint); rollbackErr != nil {
				return inserted, failed + len(products) - index - 1, fmt.Errorf("insert raw product and recover transaction: %w", errors.Join(err, rollbackErr))
			}
			slog.WarnContext(ctx, "raw product insert failed",
				"event", "product_insert_failed",
				"run_id", product.ScrapeRunID,
				"platform", product.Source,
				"result_rank", product.ResultRank,
				"reason", err,
			)
		}
		if _, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepoint); err != nil {
			return inserted, failed + len(products) - index - 1, fmt.Errorf("release product savepoint: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, len(products), fmt.Errorf("commit raw products: %w", err)
	}
	return inserted, failed, nil
}

func rawProductArgs(product model.RawProduct) []any {
	return []any{
		product.ScrapeRunID,
		product.Source,
		product.SourceProductID,
		product.SourceShopID,
		product.ProductURL,
		product.ImageURL,
		product.ProductName,
		product.ShopName,
		product.SellerLocationText,
		product.PriceText,
		product.PriceValue,
		product.Currency,
		product.Rating,
		product.ReviewCount,
		product.SoldCount,
		product.StockText,
		product.ResultRank,
		product.SourceQuery,
		product.ScrapedAt,
		product.RawPayload,
	}
}
