package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "ingestion/scraper-go/internal/scraper"
    "ingestion/scraper-go/internal/storage"
)

func main() {
    // Create context with timeout for all scraping operations
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Scrape Tokopedia
    fmt.Println("Starting Tokopedia scraping for 'beras'...")
    if err := scrapeTokopedia(ctx); err != nil {
        log.Printf("Error scraping Tokopedia: %v", err)
    } else {
        fmt.Println("✓ Tokopedia scraping completed successfully")
    }

    // Scrape BMKG Weather Forecast
    fmt.Println("\nStarting BMKG weather forecast scraping...")
    if err := scrapeBMKGWeather(ctx); err != nil {
        log.Printf("Error scraping BMKG weather: %v", err)
    } else {
        fmt.Println("✓ BMKG weather forecast scraping completed successfully")
    }

    // Scrape BMKG Latest Earthquake
    fmt.Println("\nStarting BMKG latest earthquake scraping...")
    if err := scrapeBMKGLatestEarthquake(ctx); err != nil {
        log.Printf("Error scraping BMKG latest earthquake: %v", err)
    } else {
        fmt.Println("✓ BMKG latest earthquake scraping completed successfully")
    }

    // Scrape BMKG Recent Earthquakes
    fmt.Println("\nStarting BMKG recent earthquakes scraping...")
    if err := scrapeBMKGRecentEarthquakes(ctx); err != nil {
        log.Printf("Error scraping BMKG recent earthquakes: %v", err)
    } else {
        fmt.Println("✓ BMKG recent earthquakes scraping completed successfully")
    }

    fmt.Println("\n=== All scraping operations completed ===")
}

// scrapeTokopedia scrapes Tokopedia product data
func scrapeTokopedia(ctx context.Context) error {
    data, err := scraper.ScrapeTokopedia("beras")
    if err != nil {
        return fmt.Errorf("failed to scrape Tokopedia: %w", err)
    }

    if err := storage.SaveJSON("../../data/raw/tokopedia.json", data); err != nil {
        return fmt.Errorf("failed to save Tokopedia data: %w", err)
    }

    return nil
}

// scrapeBMKGWeather scrapes BMKG weather forecast data
func scrapeBMKGWeather(ctx context.Context) error {
    data, err := scraper.ScrapeWeatherForecast(ctx, "")
    if err != nil {
        return fmt.Errorf("failed to scrape BMKG weather forecast: %w", err)
    }

    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal BMKG weather forecast data: %w", err)
    }

    if err := storage.SaveJSON("../../data/raw/bmkg_weather_forecast.json", jsonData); err != nil {
        return fmt.Errorf("failed to save BMKG weather forecast data: %w", err)
    }

    return nil
}

// scrapeBMKGLatestEarthquake scrapes the latest earthquake data from BMKG
func scrapeBMKGLatestEarthquake(ctx context.Context) error {
    data, err := scraper.ScrapeLatestEarthquake(ctx)
    if err != nil {
        return fmt.Errorf("failed to scrape BMKG latest earthquake: %w", err)
    }

    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal BMKG latest earthquake data: %w", err)
    }

    if err := storage.SaveJSON("../../data/raw/bmkg_latest_earthquake.json", jsonData); err != nil {
        return fmt.Errorf("failed to save BMKG latest earthquake data: %w", err)
    }

    return nil
}

// scrapeBMKGRecentEarthquakes scrapes recent earthquakes data from BMKG
func scrapeBMKGRecentEarthquakes(ctx context.Context) error {
    data, err := scraper.ScrapeRecentEarthquakes(ctx)
    if err != nil {
        return fmt.Errorf("failed to scrape BMKG recent earthquakes: %w", err)
    }

    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal BMKG recent earthquakes data: %w", err)
    }

    if err := storage.SaveJSON("../../data/raw/bmkg_recent_earthquakes.json", jsonData); err != nil {
        return fmt.Errorf("failed to save BMKG recent earthquakes data: %w", err)
    }

    return nil
}