package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"scraper-go/internal/model"
	"scraper-go/internal/parser"
)

// Example program demonstrating BMKG parser utilities usage

func main() {
	fmt.Println("=== BMKG Data Parser Examples ===\n")

	// Example 1: Parse weather forecast to simple format
	fmt.Println("1. Parse Weather Forecast to Simple Format")
	fmt.Println("-------------------------------------------")
	exampleSimpleForecast()
	fmt.Println()

	// Example 2: Extract current weather
	fmt.Println("2. Extract Current Weather")
	fmt.Println("-------------------------------------------")
	exampleCurrentWeather()
	fmt.Println()

	// Example 3: Filter forecast by date
	fmt.Println("3. Filter Forecast by Date")
	fmt.Println("-------------------------------------------")
	exampleFilterByDate()
	fmt.Println()

	// Example 4: Get weather summary
	fmt.Println("4. Get Weather Summary")
	fmt.Println("-------------------------------------------")
	exampleWeatherSummary()
	fmt.Println()

	// Example 5: Get daily weather summary
	fmt.Println("5. Get Daily Weather Summary")
	fmt.Println("-------------------------------------------")
	exampleDailyWeatherSummary()
	fmt.Println()

	// Example 6: Format earthquake for display
	fmt.Println("6. Format Earthquake for Display")
	fmt.Println("-------------------------------------------")
	exampleEarthquakeDisplay()
	fmt.Println()

	// Example 7: Calculate earthquake distance
	fmt.Println("7. Calculate Earthquake Distance")
	fmt.Println("-------------------------------------------")
	exampleEarthquakeDistance()
	fmt.Println()

	// Example 8: Find nearest earthquakes
	fmt.Println("8. Find Nearest Earthquakes")
	fmt.Println("-------------------------------------------")
	exampleNearestEarthquakes()
	fmt.Println()

	// Example 9: Weather icons
	fmt.Println("9. Weather Icons")
	fmt.Println("-------------------------------------------")
	exampleWeatherIcons()
	fmt.Println()
}

func exampleSimpleForecast() {
	forecast := createSampleForecast()

	simple, err := parser.ParseWeatherForecastToSimple(forecast)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Converted %d complex forecast entries to simple format:\n\n", len(simple))
	for i, sf := range simple {
		if i >= 3 {
			fmt.Printf("... and %d more entries\n", len(simple)-3)
			break
		}
		fmt.Printf("  %s | %s, %s\n", 
			sf.DateTime.Format("2006-01-02 15:04"),
			sf.Location,
			sf.Province)
		fmt.Printf("    Temperature: %.1f°C | Humidity: %d%%\n", 
			sf.Temperature, sf.Humidity)
		fmt.Printf("    Weather: %s | Wind: %s %.1f km/h\n", 
			sf.Weather, sf.WindDir, sf.WindSpeed)
		if sf.Rainfall > 0 {
			fmt.Printf("    Rainfall: %.1f mm\n", sf.Rainfall)
		}
		fmt.Println()
	}
}

func exampleCurrentWeather() {
	forecast := createSampleForecast()

	current, err := parser.ExtractCurrentWeather(forecast)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	icon := parser.GetWeatherIcon(current.Weather)
	fmt.Printf("Current Weather (closest to now):\n")
	fmt.Printf("  %s %s\n", icon, current.WeatherDescEn)
	fmt.Printf("  Temperature: %.1f°C\n", current.Temperature)
	fmt.Printf("  Humidity: %d%%\n", current.Humidity)
	fmt.Printf("  Wind: %s at %.1f km/h\n", current.WindDirection, current.WindSpeed)
	fmt.Printf("  Cloud Cover: %d%%\n", current.TotalCloudCover)
	fmt.Printf("  Visibility: %s\n", current.VisibilityText)
}

func exampleFilterByDate() {
	forecast := createSampleForecast()

	targetDate := time.Date(2026, 5, 5, 0, 0, 0, 0, time.Local)
	filtered, err := parser.FilterForecastByDate(forecast, targetDate)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Forecasts for %s:\n\n", targetDate.Format("Monday, January 2, 2006"))
	for _, w := range filtered {
		localTime, _ := w.ParseLocalDateTime()
		icon := parser.GetWeatherIcon(w.Weather)
		fmt.Printf("  %s | %s %s | %.1f°C\n",
			localTime.Format("15:04"),
			icon,
			w.WeatherDescEn,
			w.Temperature)
	}
}

func exampleWeatherSummary() {
	forecast := createSampleForecast()

	summary, err := parser.GetWeatherSummary(forecast)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Location: %s\n", summary.Location)
	fmt.Printf("Temperature: %.1f°C\n", summary.CurrentTemp)
	fmt.Printf("Feels Like: %s\n", summary.FeelsLike)
	fmt.Printf("Condition: %s\n", summary.Condition)
	fmt.Printf("Humidity: %d%%\n", summary.Humidity)
	fmt.Printf("Wind: %s\n", summary.WindDescription)
	fmt.Printf("Rain: %s\n\n", summary.RainChance)
	fmt.Printf("Summary: %s\n", summary.Summary)
}

func exampleDailyWeatherSummary() {
	forecast := createSampleForecast()

	daily, err := parser.GetDailyWeatherSummary(forecast)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Daily Weather Summary (%d days):\n\n", len(daily))
	for _, d := range daily {
		fmt.Printf("📅 %s\n", d.Date.Format("Monday, January 2, 2006"))
		fmt.Printf("   Temperature: %.1f°C - %.1f°C\n", d.MinTemp, d.MaxTemp)
		fmt.Printf("   Avg Humidity: %.0f%%\n", d.AvgHumidity)
		fmt.Printf("   Total Rain: %.1f mm\n", d.TotalRain)
		fmt.Printf("   Conditions: %v\n\n", d.Conditions)
	}
}

func exampleEarthquakeDisplay() {
	earthquake := createSampleEarthquake()

	display, err := parser.FormatEarthquakeForDisplay(earthquake)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("🌍 Earthquake Information\n\n")
	fmt.Printf("Time: %s\n", display.Time)
	fmt.Printf("Location: %s\n", display.Location)
	fmt.Printf("Magnitude: %.1f\n", display.Magnitude)
	fmt.Printf("Depth: %.0f km\n", display.Depth)
	fmt.Printf("Coordinates: %s\n", display.Coordinates)
	fmt.Printf("Potential: %s\n", display.Potential)
	fmt.Printf("Shakemap: %s\n\n", display.ShakemapURL)
	fmt.Printf("Description:\n%s\n", display.Description)

	// Check severity
	severity := parser.GetEarthquakeSeverity(display.Magnitude)
	fmt.Printf("\nSeverity Level: %s\n", severity)

	if parser.IsSignificantEarthquake(earthquake) {
		fmt.Println("⚠️  This is a significant earthquake (M ≥ 5.0)")
	}
}

func exampleEarthquakeDistance() {
	earthquake := createSampleEarthquake()

	// Major Indonesian cities coordinates
	cities := map[string][2]float64{
		"Jakarta":   {-6.2088, 106.8456},
		"Bandung":   {-6.9175, 107.6191},
		"Surabaya":  {-7.2575, 112.7521},
		"Medan":     {3.5952, 98.6722},
		"Makassar":  {-5.1477, 119.4327},
		"Palembang": {-2.9761, 104.7754},
	}

	fmt.Println("Distance from earthquake to major cities:\n")
	for city, coords := range cities {
		distance, err := parser.CalculateEarthquakeDistance(earthquake, coords[0], coords[1])
		if err != nil {
			log.Printf("Error calculating distance to %s: %v\n", city, err)
			continue
		}
		fmt.Printf("  %s: %.2f km\n", city, distance)
	}
}

func exampleNearestEarthquakes() {
	earthquakes := createSampleEarthquakeList()

	// Find earthquakes within 500 km of Jakarta
	jakartaLat := -6.2088
	jakartaLon := 106.8456
	radiusKm := 500.0

	nearby, err := parser.FindNearestEarthquakes(earthquakes, jakartaLat, jakartaLon, radiusKm)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Earthquakes within %.0f km of Jakarta:\n\n", radiusKm)
	if len(nearby) == 0 {
		fmt.Println("  No earthquakes found in this radius")
		return
	}

	for i, eq := range nearby {
		magnitude := eq.Earthquake.Magnitude
		fmt.Printf("%d. M%s - %.2f km away\n", i+1, magnitude, eq.Distance)
		fmt.Printf("   %s\n", eq.Earthquake.Region)
		fmt.Printf("   %s\n\n", eq.Earthquake.Time)
	}
}

func exampleWeatherIcons() {
	weatherCodes := []struct {
		code int
		name string
	}{
		{model.WeatherCodeSunny, "Sunny"},
		{model.WeatherCodePartlyCloudy, "Partly Cloudy"},
		{model.WeatherCodeMostlyCloudy, "Mostly Cloudy"},
		{model.WeatherCodeOvercast, "Overcast"},
		{model.WeatherCodeHaze, "Haze"},
		{model.WeatherCodeFog, "Fog"},
		{model.WeatherCodeLightRain, "Light Rain"},
		{model.WeatherCodeModerateRain, "Moderate Rain"},
		{model.WeatherCodeHeavyRain, "Heavy Rain"},
		{model.WeatherCodeThunderstorm, "Thunderstorm"},
	}

	fmt.Println("Weather condition icons:\n")
	for _, wc := range weatherCodes {
		icon := parser.GetWeatherIcon(wc.code)
		fmt.Printf("  %s  %s\n", icon, wc.name)
	}
}

// Helper functions to create sample data

func createSampleForecast() model.WeatherForecastResponse {
	location := model.Location{
		Province:  "DKI Jakarta",
		CityRegcy: "Jakarta Pusat",
		District:  "Menteng",
		Village:   "Gondangdia",
		Latitude:  -6.1944,
		Longitude: 106.8229,
		Timezone:  "Asia/Jakarta",
	}

	weather1 := model.WeatherInfo{
		LocalDateTime:   "2026-05-04 08:00:00",
		UTCDateTime:     "2026-05-04 01:00:00",
		Temperature:     28.5,
		Humidity:        70,
		TotalCloudCover: 40,
		Precipitation:   0,
		Weather:         model.WeatherCodePartlyCloudy,
		WeatherDescEn:   "Partly Cloudy",
		WindSpeed:       15.5,
		WindDirection:   "NE",
		Visibility:      10000,
		VisibilityText:  "> 10 km",
	}

	weather2 := model.WeatherInfo{
		LocalDateTime:   "2026-05-04 14:00:00",
		UTCDateTime:     "2026-05-04 07:00:00",
		Temperature:     32.0,
		Humidity:        65,
		TotalCloudCover: 20,
		Precipitation:   0,
		Weather:         model.WeatherCodeSunny,
		WeatherDescEn:   "Sunny",
		WindSpeed:       12.0,
		WindDirection:   "E",
		Visibility:      10000,
		VisibilityText:  "> 10 km",
	}

	weather3 := model.WeatherInfo{
		LocalDateTime:   "2026-05-04 20:00:00",
		UTCDateTime:     "2026-05-04 13:00:00",
		Temperature:     26.0,
		Humidity:        80,
		TotalCloudCover: 70,
		Precipitation:   2.5,
		Weather:         model.WeatherCodeLightRain,
		WeatherDescEn:   "Light Rain",
		WindSpeed:       18.0,
		WindDirection:   "W",
		Visibility:      8000,
		VisibilityText:  "8 km",
	}

	weather4 := model.WeatherInfo{
		LocalDateTime:   "2026-05-05 08:00:00",
		UTCDateTime:     "2026-05-05 01:00:00",
		Temperature:     27.0,
		Humidity:        75,
		TotalCloudCover: 60,
		Precipitation:   0,
		Weather:         model.WeatherCodeMostlyCloudy,
		WeatherDescEn:   "Mostly Cloudy",
		WindSpeed:       10.0,
		WindDirection:   "N",
		Visibility:      10000,
		VisibilityText:  "> 10 km",
	}

	return model.WeatherForecastResponse{
		Location: location,
		Data: []model.ForecastData{
			{
				Location: location,
				Weather: [][]model.WeatherInfo{
					{weather1, weather2, weather3},
					{weather4},
				},
			},
		},
	}
}

func createSampleEarthquake() model.EarthquakeDetail {
	return model.EarthquakeDetail{
		Date:        "04 Mei 2026",
		Time:        "15:30:45 WIB",
		DateTime:    "2026-05-04T08:30:45Z",
		Coordinates: "-3.80,102.30",
		Latitude:    "3.80 LS",
		Longitude:   "102.30 BT",
		Magnitude:   "5.7",
		Depth:       "10 km",
		Region:      "Pusat gempa berada di laut 50 km barat daya Bengkulu",
		Potential:   "Tidak berpotensi tsunami",
		Shakemap:    "20260504153045.mmi.jpg",
	}
}

func createSampleEarthquakeList() []model.EarthquakeDetail {
	return []model.EarthquakeDetail{
		{
			Date:        "04 Mei 2026",
			Time:        "15:30:45 WIB",
			DateTime:    "2026-05-04T08:30:45Z",
			Coordinates: "-3.80,102.30",
			Magnitude:   "5.7",
			Depth:       "10 km",
			Region:      "50 km barat daya Bengkulu",
			Potential:   "Tidak berpotensi tsunami",
		},
		{
			Date:        "04 Mei 2026",
			Time:        "10:15:30 WIB",
			DateTime:    "2026-05-04T03:15:30Z",
			Coordinates: "-6.50,107.00",
			Magnitude:   "4.2",
			Depth:       "15 km",
			Region:      "25 km tenggara Bandung",
			Potential:   "Dirasakan II-III MMI",
		},
		{
			Date:        "03 Mei 2026",
			Time:        "22:45:12 WIB",
			DateTime:    "2026-05-03T15:45:12Z",
			Coordinates: "1.50,97.50",
			Magnitude:   "6.1",
			Depth:       "20 km",
			Region:      "100 km barat laut Sibolga",
			Potential:   "Tidak berpotensi tsunami",
		},
		{
			Date:        "03 Mei 2026",
			Time:        "08:20:00 WIB",
			DateTime:    "2026-05-03T01:20:00Z",
			Coordinates: "-8.50,115.50",
			Magnitude:   "5.0",
			Depth:       "8 km",
			Region:      "30 km timur laut Denpasar",
			Potential:   "Dirasakan III MMI",
		},
	}
}

// SaveToJSON demonstrates saving parsed data to JSON file
func SaveToJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
