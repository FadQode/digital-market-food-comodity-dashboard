# BMKG Parser Quick Reference

## 🚀 Quick Start

```go
import "scraper-go/internal/parser"
```

## 📊 Weather Functions

### Get Current Weather
```go
current, err := parser.ExtractCurrentWeather(forecast)
icon := parser.GetWeatherIcon(current.Weather)
fmt.Printf("%s %.1f°C\n", icon, current.Temperature)
```

### Get Weather Summary
```go
summary, err := parser.GetWeatherSummary(forecast)
fmt.Println(summary.Summary)
```

### Filter by Date
```go
date := time.Date(2026, 5, 5, 0, 0, 0, 0, time.Local)
forecasts, err := parser.FilterForecastByDate(forecast, date)
```

### Daily Summary
```go
daily, err := parser.GetDailyWeatherSummary(forecast)
for _, d := range daily {
    fmt.Printf("%s: %.1f°C - %.1f°C\n", 
        d.Date.Format("2006-01-02"), d.MinTemp, d.MaxTemp)
}
```

### Simplify Forecast
```go
simple, err := parser.ParseWeatherForecastToSimple(forecast)
for _, sf := range simple {
    fmt.Printf("%s: %.1f°C, %s\n", 
        sf.DateTime.Format("15:04"), sf.Temperature, sf.Weather)
}
```

## 🌍 Earthquake Functions

### Format for Display
```go
display, err := parser.FormatEarthquakeForDisplay(earthquake)
fmt.Printf("M%.1f at %.0f km depth\n", display.Magnitude, display.Depth)
fmt.Println(display.Description)
```

### Calculate Distance
```go
// From Jakarta (-6.2088, 106.8456)
distance, err := parser.CalculateEarthquakeDistance(earthquake, -6.2088, 106.8456)
fmt.Printf("Distance: %.2f km\n", distance)
```

### Find Nearby Earthquakes
```go
// Within 500 km of Jakarta
nearby, err := parser.FindNearestEarthquakes(earthquakes, -6.2088, 106.8456, 500.0)
for _, eq := range nearby {
    fmt.Printf("M%s at %.2f km\n", eq.Earthquake.Magnitude, eq.Distance)
}
```

### Check Significance
```go
if parser.IsSignificantEarthquake(earthquake) {
    magnitude, _ := strconv.ParseFloat(earthquake.Magnitude, 64)
    severity := parser.GetEarthquakeSeverity(magnitude)
    fmt.Printf("⚠️ %s earthquake\n", severity)
}
```

## 🎨 Weather Icons

| Code | Icon | Description |
|------|------|-------------|
| 1 | ☀️ | Sunny |
| 2 | ⛅ | Partly Cloudy |
| 3 | ☁️ | Mostly Cloudy |
| 4 | ☁️ | Overcast |
| 5 | 🌫️ | Haze |
| 10 | 🌫️ | Smoke |
| 45 | 🌫️ | Fog |
| 60 | 🌧️ | Light Rain |
| 61 | 🌧️ | Moderate Rain |
| 63 | 🌧️ | Heavy Rain |
| 80 | ⛈️ | Isolated Shower |
| 95 | ⚡ | Severe Thunderstorm |
| 97 | ⚡ | Thunderstorm |

```go
icon := parser.GetWeatherIcon(weatherCode)
```

## 📏 Earthquake Severity

| Magnitude | Severity |
|-----------|----------|
| < 3.0 | Minor |
| 3.0 - 4.9 | Light |
| 5.0 - 5.9 | Moderate |
| 6.0 - 6.9 | Strong |
| 7.0 - 7.9 | Major |
| ≥ 8.0 | Great |

```go
severity := parser.GetEarthquakeSeverity(magnitude)
```

## 🔧 Common Patterns

### Weather Dashboard
```go
func getWeatherDashboard(forecast model.WeatherForecastResponse) {
    // Current conditions
    current, _ := parser.ExtractCurrentWeather(forecast)
    icon := parser.GetWeatherIcon(current.Weather)
    
    // Summary
    summary, _ := parser.GetWeatherSummary(forecast)
    
    // Daily forecast
    daily, _ := parser.GetDailyWeatherSummary(forecast)
    
    // Display
    fmt.Printf("%s %s - %.1f°C\n", icon, current.WeatherDescEn, current.Temperature)
    fmt.Println(summary.Summary)
    for _, d := range daily {
        fmt.Printf("%s: %.1f°C - %.1f°C\n", 
            d.Date.Format("Mon"), d.MinTemp, d.MaxTemp)
    }
}
```

### Earthquake Alert System
```go
func checkEarthquakeAlerts(userLat, userLon, radiusKm float64) {
    earthquakes := fetchRecentEarthquakes()
    nearby, _ := parser.FindNearestEarthquakes(earthquakes, userLat, userLon, radiusKm)
    
    for _, eq := range nearby {
        if parser.IsSignificantEarthquake(eq.Earthquake) {
            display, _ := parser.FormatEarthquakeForDisplay(eq.Earthquake)
            fmt.Printf("⚠️ Alert: %s\n", display.Description)
            fmt.Printf("Distance: %.2f km\n", eq.Distance)
        }
    }
}
```

### Data Export
```go
func exportWeatherData(forecast model.WeatherForecastResponse) {
    // Simplify for export
    simple, _ := parser.ParseWeatherForecastToSimple(forecast)
    
    // Convert to JSON
    data, _ := json.MarshalIndent(simple, "", "  ")
    
    // Save to file
    os.WriteFile("weather_export.json", data, 0644)
}
```

### Weather Comparison
```go
func compareWeatherByDate(forecast model.WeatherForecastResponse, date1, date2 time.Time) {
    forecasts1, _ := parser.FilterForecastByDate(forecast, date1)
    forecasts2, _ := parser.FilterForecastByDate(forecast, date2)
    
    avg1 := calculateAverageTemp(forecasts1)
    avg2 := calculateAverageTemp(forecasts2)
    
    fmt.Printf("%s: %.1f°C\n", date1.Format("2006-01-02"), avg1)
    fmt.Printf("%s: %.1f°C\n", date2.Format("2006-01-02"), avg2)
    fmt.Printf("Difference: %.1f°C\n", avg2-avg1)
}
```

## 🧪 Testing

### Run Tests
```bash
# All tests
go test -v ./internal/parser/

# Specific test
go test -v ./internal/parser/ -run TestParseWeatherForecastToSimple

# With coverage
go test -cover ./internal/parser/

# Benchmarks
go test -bench=. ./internal/parser/
```

### Example Test
```go
func TestWeatherParser(t *testing.T) {
    forecast := createTestForecast()
    
    simple, err := parser.ParseWeatherForecastToSimple(forecast)
    if err != nil {
        t.Fatalf("Error: %v", err)
    }
    
    if len(simple) == 0 {
        t.Error("Expected forecasts, got none")
    }
}
```

## 📦 Data Structures

### SimpleWeatherForecast
```go
type SimpleWeatherForecast struct {
    Location    string
    Province    string
    DateTime    time.Time
    Temperature float64
    Humidity    int
    Weather     string
    WeatherCode int
    WindSpeed   float64
    WindDir     string
    Rainfall    float64
}
```

### WeatherSummary
```go
type WeatherSummary struct {
    Location        string
    CurrentTemp     float64
    FeelsLike       string
    Condition       string
    Humidity        int
    WindDescription string
    RainChance      string
    Summary         string
}
```

### EarthquakeDisplay
```go
type EarthquakeDisplay struct {
    Time        string
    Location    string
    Magnitude   float64
    Depth       float64
    Coordinates string
    Potential   string
    ShakemapURL string
    Description string
}
```

### DailyWeatherSummary
```go
type DailyWeatherSummary struct {
    Date        time.Time
    MinTemp     float64
    MaxTemp     float64
    AvgHumidity float64
    Conditions  []string
    TotalRain   float64
}
```

### EarthquakeDistance
```go
type EarthquakeDistance struct {
    Earthquake model.EarthquakeDetail
    Distance   float64
}
```

## 🌐 Major Indonesian Cities Coordinates

```go
var cities = map[string][2]float64{
    "Jakarta":    {-6.2088, 106.8456},
    "Surabaya":   {-7.2575, 112.7521},
    "Bandung":    {-6.9175, 107.6191},
    "Medan":      {3.5952, 98.6722},
    "Semarang":   {-6.9667, 110.4167},
    "Makassar":   {-5.1477, 119.4327},
    "Palembang":  {-2.9761, 104.7754},
    "Tangerang":  {-6.1783, 106.6319},
    "Depok":      {-6.4025, 106.7942},
    "Bekasi":     {-6.2349, 106.9896},
    "Yogyakarta": {-7.7956, 110.3695},
    "Malang":     {-7.9797, 112.6304},
    "Bogor":      {-6.5950, 106.7969},
    "Batam":      {1.0456, 104.0305},
    "Denpasar":   {-8.6705, 115.2126},
}
```

## 💡 Tips

1. **Error Handling**: Always check errors returned by parser functions
2. **Time Zones**: Weather data uses local time, earthquakes use UTC
3. **Distance Calculation**: Uses Haversine formula (great-circle distance)
4. **Performance**: Parser functions are O(n) or better
5. **Pure Functions**: No side effects, safe for concurrent use
6. **Null Safety**: Check for nil before accessing pointer fields

## 🔗 Related Files

- `internal/parser/bmkg_parser.go` - Main implementation
- `internal/parser/bmkg_parser_test.go` - Unit tests
- `internal/parser/README.md` - Full documentation
- `internal/parser/IMPLEMENTATION.md` - Implementation details
- `examples/parser_example.go` - Usage examples

## 📚 Resources

- BMKG API: https://api.bmkg.go.id/
- Weather Codes: See `internal/model/weather.go`
- Haversine Formula: https://en.wikipedia.org/wiki/Haversine_formula

---

**Version**: 1.0.0  
**Last Updated**: 2026-05-04  
**Status**: ✅ Production Ready
