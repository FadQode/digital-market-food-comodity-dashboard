INSERT INTO regions (region_name, latitude, longitude)
SELECT 'Surabaya', -7.25, 112.75
WHERE NOT EXISTS (
    SELECT 1 FROM regions WHERE region_name = 'Surabaya'
);

INSERT INTO commodities (commodity_name, category_lvl1, category_lvl2)
SELECT 'Beras', 'Sembako', 'Beras'
WHERE NOT EXISTS (
    SELECT 1 FROM commodities WHERE commodity_name = 'Beras'
);
