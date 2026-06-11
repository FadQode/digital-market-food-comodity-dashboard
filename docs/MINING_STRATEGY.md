# Marketplace Mining Strategy

## Objective

Collect a repeatable weekly sample of Indonesian staple-food listings that can
later be normalized to price per kilogram and compared with BPS/reference
prices. The first pilot commodity is rice.

## Budget Guardrails

The default paid run uses:

- 20 Tokopedia results;
- 20 Shopee results;
- one standardized query: `beras 5 kg`;
- a USD 0.07 maximum charge per Actor invocation;
- a weekly schedule;
- no automatic retry of paid Actor POST requests.

At the published pay-per-result rates checked on June 11, 2026, 20 results are
approximately USD 0.018 for Tokopedia and USD 0.060 for Shopee. The API charge
cap limits either invocation to USD 0.07. A weekly run therefore preserves most
of each USD 5 budget for validation reruns and future commodities.

Actor pricing can change. Review these pages before increasing limits:

- https://apify.com/fatihtahta/tokopedia-scraper
- https://apify.com/pumpkin_jingo/shopee-scraper-id

## Why This Query

`beras 5 kg` is more useful than `beras` because it reduces unrelated products,
small samples, sacks, bundles, and wholesale listings. The later transform
stage should still parse package size from the title and reject records whose
actual quantity cannot be established.

Use the same query, sort order, item count, and weekday for a time series. A
changing search definition can look like a price change even when the market
did not change.

## Minimum Useful Raw Record

Preserve the complete Actor payload. The transform stage should derive at least:

- source marketplace and scrape run ID;
- source product ID and canonical URL;
- product title;
- current and original price;
- parsed package quantity and unit;
- normalized price per kilogram;
- seller ID and seller name;
- seller city/province text;
- rating, review count, and sold count;
- stock or availability signal;
- sponsored/official-store flags when available;
- source query, result rank, and scrape timestamp.

Do not use displayed listing price directly as the commodity price until
quantity, variants, bundles, vouchers, and minimum-order rules are checked.

## Quality Rules

Exclude or flag:

- listings without a parseable package weight;
- prices that apply only to a low-price variant;
- bundles containing non-rice products;
- wholesale/minimum-order prices that are not comparable to retail;
- duplicate product IDs within a run;
- sponsored listings when rank analysis is not the objective;
- implausible price-per-kilogram outliers.

For reporting, prefer the median normalized price and include sample count,
interquartile range, and marketplace. Do not present a mean from fewer than ten
valid listings as a regional market price.

## Expansion Order

1. Validate one 20-result run from each marketplace and profile returned fields.
2. Build actor-specific raw-to-common-field adapters.
3. Insert each listing into `products_raw` with the complete JSON payload.
4. Add quantity parsing and data-quality flags.
5. Only then add another standardized commodity query.

This order avoids spending budget on data that the pipeline cannot yet validate
or normalize.
