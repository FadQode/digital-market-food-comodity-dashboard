SELECT
    id,
    scrape_run_id,
    source,
    source_product_id,
    product_name,
    price_text,
    price_value,
    currency,
    shop_name,
    seller_location_text,
    result_rank,
    scraped_at
FROM products_raw
ORDER BY created_at DESC
LIMIT 50;
