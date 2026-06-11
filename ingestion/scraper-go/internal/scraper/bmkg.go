package scraper

import (
	"context"
	"fmt"

	"ingestion/scraper-go/internal/client"
	"ingestion/scraper-go/internal/model"
)

var (
	// bmkgClient is the singleton BMKG client instance
	bmkgClient client.BMKGClient
)

// init initializes the BMKG client with default configuration
func init() {
	bmkgClient = client.NewBMKGClient(client.DefaultBMKGClientConfig())
}

// ScrapeWeatherForecast retrieves weather forecast for a specific province
// provinceID: optional province ID parameter (e.g., "31" for DKI Jakarta)
// If empty, returns forecast for all provinces
func ScrapeWeatherForecast(ctx context.Context, provinceID string) (*model.WeatherForecastResponse, error) {
	forecast, err := bmkgClient.GetWeatherForecast(ctx, provinceID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape weather forecast: %w", err)
	}
	return forecast, nil
}

// ScrapeLatestEarthquake retrieves the latest earthquake information
func ScrapeLatestEarthquake(ctx context.Context) (*model.EarthquakeLatestResponse, error) {
	earthquake, err := bmkgClient.GetLatestEarthquake(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape latest earthquake: %w", err)
	}
	return earthquake, nil
}

// ScrapeRecentEarthquakes retrieves list of recent earthquakes
func ScrapeRecentEarthquakes(ctx context.Context) (*model.EarthquakeListResponse, error) {
	earthquakes, err := bmkgClient.GetRecentEarthquakes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape recent earthquakes: %w", err)
	}
	return earthquakes, nil
}

// ScrapeFeltEarthquakes retrieves list of earthquakes that were felt by people
func ScrapeFeltEarthquakes(ctx context.Context) (*model.EarthquakeListResponse, error) {
	earthquakes, err := bmkgClient.GetFeltEarthquakes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape felt earthquakes: %w", err)
	}
	return earthquakes, nil
}
