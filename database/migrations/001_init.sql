
-- +goose Up
CREATE TABLE IF NOT EXISTS regions (
    region_id SERIAL PRIMARY KEY,
    region_name TEXT,
    latitude FLOAT,
    longitude FLOAT
);

CREATE TABLE IF NOT EXISTS commodities (
    commodity_id SERIAL PRIMARY KEY,
    commodity_name TEXT,
    category_lvl1 TEXT,
    category_lvl2 TEXT
);

CREATE TABLE IF NOT EXISTS price_data (
    id SERIAL PRIMARY KEY,
    date DATE,
    region_id INT REFERENCES regions(region_id),
    commodity_id INT REFERENCES commodities(commodity_id),
    price FLOAT,
    source TEXT
);

CREATE TABLE IF NOT EXISTS weather_data (
    id SERIAL PRIMARY KEY,
    date DATE,
    region_id INT REFERENCES regions(region_id),
    rainfall FLOAT,
    temperature FLOAT,
    source TEXT
);

CREATE TABLE IF NOT EXISTS demand_data (
    id SERIAL PRIMARY KEY,
    date DATE,
    region_id INT REFERENCES regions(region_id),
    commodity_id INT REFERENCES commodities(commodity_id),
    demand_index FLOAT,
    source TEXT
);

CREATE TABLE IF NOT EXISTS combined_features (
    id SERIAL PRIMARY KEY,
    date DATE,
    region_id INT,
    commodity_id INT,
    price FLOAT,
    demand_index FLOAT,
    rainfall FLOAT,
    temperature FLOAT
);

-- +goose Down
DROP TABLE IF EXISTS combined_features;
DROP TABLE IF EXISTS demand_data;
DROP TABLE IF EXISTS weather_data;
DROP TABLE IF EXISTS price_data;
DROP TABLE IF EXISTS commodities;
DROP TABLE IF EXISTS regions;
