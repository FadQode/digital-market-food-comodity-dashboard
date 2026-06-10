# BMKG Data Parser

This package provides utility functions for parsing and transforming BMKG (Badan Meteorologi, Klimatologi, dan Geofisika) weather and earthquake data into more consumable formats.

## Features

### Weather Forecast Parsing

- **ParseWeatherForecastToSimple**: Convert complex nested forecast structures into a flat, simplified format
- **ExtractCurrentWeather**: Get the weather forecast closest to the current time
- **FilterForecastByDate**: Filter forecasts for a specific date
- **FilterForecastByDateRange**: Filter forecasts within a date range
- **GetWeatherSummary**: Generate human-readable weather summaries
- **GetDailyWeatherSummary**: Aggregate weather data by day with min/max temperatures

### Earthquake Data Processing

- **FormatEarthquakeForDisplay**: Format earthquake data for human-readable output
- **CalculateEarthquakeDistance**: Calculate distance from coordinates using Haversine formula
- **FindNearestEarthquakes**: Find earthquakes within a radius and sort by distance
- **IsSignificantEarthquake**: Determine if earthquake is significant (M ≥ 5.0)
- **GetEarthquakeSeverity**: Get severity level based on magnitude

### Helper Functions

- **GetWeatherIcon**: Get emoji icon for weather conditions
- **haversineDistance**: Calculate great-circle distance between two points

## Usage Examples

### Parse Weather Forecast to Simple Format

```go
import (
    "scraper-go/internal/model"
    "scraper-go/internal/parser"
)

// Assume you have a WeatherForecastResponse from BMKG API
var forecast model.WeatherForecastResponse

// Convert to simple format
simpleForecasts, err := parser.ParseWeatherForecastToSimple(forecast)
if err != nil {
    log.Fatal(err)
}

for _, sf := range simpleForecasts {
    fmt.Printf("%s: %.1f°C, %s\n", 
        sf.DateTime.Format("2006-01-02 15:04"), 
        sf.Temperature, 
        sf.Weather)
}
```

### Extract Current Weather

```go
currentWeather, err := parser.ExtractCurrentWeather(forecast)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current temperature: %.1f°C\n", currentWeather.Temperature)
fmt.Printf("Condition: %s\n", currentWeather.WeatherDescEn)
fmt.Printf("Humidity: %d%%\n", currentWeather.Humidity)
```

### Filter Forecast by Date

```go
targetDate := time.Date(2026, 5, 5, 0, 0, 0, 0, time.Local)
forecasts, err := parser.FilterForecastByDate(forecast, targetDate)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d forecasts for %s\n", 
    len(forecasts), 
    targetDate.Format("2006-01-02"))
```

### Get Weather Summary

```go
summary, err := parser.GetWeatherSummary(forecast)
if err != nil {
    log.Fatal(err)
}

fmt.Println(summary.Summary)
// Output: Currently partly cloudy in Jakarta Pusat, DKI Jakarta with temperature 28.5°C. 
//         Feels warm and comfortable. Light winds from NE at 15.5 km/h. No rain expected.
```

### Get Daily Weather Summary

```go
dailySummaries, err := parser.GetDailyWeatherSummary(forecast)
if err != nil {
    log.Fatal(err)
}

for _, daily := range dailySummaries {
    fmt.Printf("%s: %.1f°C - %.1f°C, Rain: %.1fmm\n",
        daily.Date.Format("2006-01-02"),
        daily.MinTemp,
        daily.MaxTemp,
        daily.TotalRain)
    fmt.Printf("  Conditions: %v\n", daily.Conditions)
}
```

### Format Earthquake for Display

```go
var earthquake model.EarthquakeDetail

display, err := parser.FormatEarthquakeForDisplay(earthquake)
if err != nil {
    log.Fatal(err)
}

fmt.Println(display.Description)
fmt.Printf("Magnitude: %.1f\n", display.Magnitude)
fmt.Printf("Depth: %.0f km\n", display.Depth)
fmt.Printf("Shakemap: %s\n", display.ShakemapURL)
```

### Calculate Earthquake Distance

```go
// Calculate distance from Jakarta (-6.2088, 106.8456)
jakartaLat := -6.2088
jakartaLon := 106.8456

distance, err := parser.CalculateEarthquakeDistance(earthquake, jakartaLat, jakartaLon)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Distance from Jakarta: %.2f km\n", distance)
```

### Find Nearest Earthquakes

```go
var earthquakes []model.EarthquakeDetail

// Find earthquakes within 500 km of Jakarta
jakartaLat := -6.2088
jakartaLon := 106.8456
radiusKm := 500.0

nearby, err := parser.FindNearestEarthquakes(earthquakes, jakartaLat, jakartaLon, radiusKm)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d earthquakes within %.0f km:\n", len(nearby), radiusKm)
for _, eq := range nearby {
    fmt.Printf("  M%.1f at %.2f km - %s\n",
        eq.Earthquake.Magnitude,
        eq.Distance,
        eq.Earthquake.Region)
}
```

### Check Earthquake Significance

```go
if parser.IsSignificantEarthquake(earthquake) {
    fmt.Println("This is a significant earthquake (M ≥ 5.0)")
    
    magnitude, _ := strconv.ParseFloat(earthquake.Magnitude, 64)
    severity := parser.GetEarthquakeSeverity(magnitude)
    fmt.Printf("Severity: %s\n", severity)
}
```

### Get Weather Icon

```go
icon := parser.GetWeatherIcon(currentWeather.Weather)
fmt.Printf("%s %s\n", icon, currentWeather.WeatherDescEn)
// Output: ⛅ Partly Cloudy
```

## Data Structures

### SimpleWeatherForecast

Simplified weather forecast structure:

```go
type SimpleWeatherForecast struct {
    Location    string    // Location name
    Province    string    // Province name
    DateTime    time.Time // Forecast datetime
    Temperature float64   // Temperature in °C
    Humidity    int       // Humidity in %
    Weather     string    // Weather description
    WeatherCode int       // Weather code
    WindSpeed   float64   // Wind speed in km/h
    WindDir     string    // Wind direction
    Rainfall    float64   // Rainfall in mm
}
```

### WeatherSummary

Human-readable weather summary:

```go
type WeatherSummary struct {
    Location        string  // Location name
    CurrentTemp     float64 // Current temperature
    FeelsLike       string  // Feels like description
    Condition       string  // Weather condition
    Humidity        int     // Humidity percentage
    WindDescription string  // Wind description
    RainChance      string  // Rain probability
    Summary         string  // Full summary text
}
```

### EarthquakeDisplay

Formatted earthquake data:

```go
type EarthquakeDisplay struct {
    Time        string  // Formatted time
    Location    string  // Location description
    Magnitude   float64 // Magnitude
    Depth       float64 // Depth in km
    Coordinates string  // Coordinates string
    Potential   string  // Tsunami potential
    ShakemapURL string  // Shakemap image URL
    Description string  // Full description
}
```

### DailyWeatherSummary

Daily aggregated weather data:

```go
type DailyWeatherSummary struct {
    Date        time.Time // Date
    MinTemp     float64   // Minimum temperature
    MaxTemp     float64   // Maximum temperature
    AvgHumidity float64   // Average humidity
    Conditions  []string  // List of conditions
    TotalRain   float64   // Total rainfall in mm
}
```

### EarthquakeDistance

Earthquake with calculated distance:

```go
type EarthquakeDistance struct {
    Earthquake model.EarthquakeDetail // Earthquake details
    Distance   float64                // Distance in km
}
```

## Design Principles

1. **Pure Functions**: All functions are pure with no side effects
2. **Error Handling**: Proper error handling with descriptive messages
3. **Type Safety**: Strong typing with clear data structures
4. **Performance**: Efficient algorithms (Haversine for distance calculation)
5. **Testability**: All functions are unit-testable
6. **Documentation**: Clear documentation and examples

## Weather Codes

The parser uses BMKG weather codes:

- `1` - Sunny (☀️)
- `2` - Partly Cloudy (⛅)
- `3` - Mostly Cloudy (☁️)
- `4` - Overcast (☁️)
- `5` - Haze (🌫️)
- `10` - Smoke (🌫️)
- `45` - Fog (🌫️)
- `60` - Light Rain (🌧️)
- `61` - Moderate Rain (🌧️)
- `63` - Heavy Rain (🌧️)
- `80` - Isolated Shower (⛈️)
- `95` - Severe Thunderstorm (⚡)
- `97` - Thunderstorm (⚡)

## Earthquake Severity Levels

Based on magnitude:

- **Minor**: < 3.0
- **Light**: 3.0 - 4.9
- **Moderate**: 5.0 - 5.9
- **Strong**: 6.0 - 6.9
- **Major**: 7.0 - 7.9
- **Great**: ≥ 8.0

## Testing

Run tests with:

```bash
go test -v ./internal/parser/
```

Run benchmarks:

```bash
go test -bench=. ./internal/parser/
```

## Dependencies

- Standard library only (`math`, `time`, `strings`, `strconv`, `fmt`)
- Internal models: `scraper-go/internal/model`

## License

Part of the Datathon Dicoding project.
