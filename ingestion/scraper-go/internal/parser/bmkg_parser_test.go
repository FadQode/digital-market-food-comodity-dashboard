package parser

import (
	"math"
	"testing"
	"time"

	"scraper-go/internal/model"
)

// Test data helpers

func createTestWeatherInfo(datetime string, temp float64, humidity int, weatherCode int, weatherDesc string, precipitation float64) model.WeatherInfo {
	return model.WeatherInfo{
		LocalDateTime:   datetime,
		UTCDateTime:     datetime,
		Temperature:     temp,
		Humidity:        humidity,
		Weather:         weatherCode,
		WeatherDescEn:   weatherDesc,
		Precipitation:   precipitation,
		WindSpeed:       15.5,
		WindDirection:   "NE",
		TotalCloudCover: 50,
		Visibility:      10000,
	}
}

func createTestForecastResponse() model.WeatherForecastResponse {
	location := model.Location{
		Province:  "DKI Jakarta",
		CityRegcy: "Jakarta Pusat",
		District:  "Menteng",
		Village:   "Gondangdia",
		Latitude:  -6.1944,
		Longitude: 106.8229,
	}

	weather1 := createTestWeatherInfo("2026-05-04 08:00:00", 28.5, 70, model.WeatherCodePartlyCloudy, "Partly Cloudy", 0)
	weather2 := createTestWeatherInfo("2026-05-04 14:00:00", 32.0, 65, model.WeatherCodeSunny, "Sunny", 0)
	weather3 := createTestWeatherInfo("2026-05-04 20:00:00", 26.0, 80, model.WeatherCodeLightRain, "Light Rain", 2.5)
	weather4 := createTestWeatherInfo("2026-05-05 08:00:00", 27.0, 75, model.WeatherCodeMostlyCloudy, "Mostly Cloudy", 0)

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

func createTestEarthquake(magnitude, depth string, coords string, region string) model.EarthquakeDetail {
	return model.EarthquakeDetail{
		Date:        "04 Mei 2026",
		Time:        "15:30:45 WIB",
		DateTime:    "2026-05-04T08:30:45Z",
		Coordinates: coords,
		Latitude:    "5.50 LU",
		Longitude:   "95.50 BT",
		Magnitude:   magnitude,
		Depth:       depth,
		Region:      region,
		Potential:   "Tidak berpotensi tsunami",
		Shakemap:    "20260504153045.mmi.jpg",
	}
}

// Tests for ParseWeatherForecastToSimple

func TestParseWeatherForecastToSimple(t *testing.T) {
	forecast := createTestForecastResponse()

	simple, err := ParseWeatherForecastToSimple(forecast)
	if err != nil {
		t.Fatalf("ParseWeatherForecastToSimple failed: %v", err)
	}

	if len(simple) != 4 {
		t.Errorf("Expected 4 simple forecasts, got %d", len(simple))
	}

	// Check first forecast
	if simple[0].Location != "Gondangdia, Menteng" {
		t.Errorf("Expected location 'Gondangdia, Menteng', got '%s'", simple[0].Location)
	}

	if simple[0].Temperature != 28.5 {
		t.Errorf("Expected temperature 28.5, got %.1f", simple[0].Temperature)
	}

	if simple[0].Weather != "Partly Cloudy" {
		t.Errorf("Expected weather 'Partly Cloudy', got '%s'", simple[0].Weather)
	}
}

func TestParseWeatherForecastToSimple_EmptyData(t *testing.T) {
	forecast := model.WeatherForecastResponse{
		Data: []model.ForecastData{},
	}

	simple, err := ParseWeatherForecastToSimple(forecast)
	if err != nil {
		t.Fatalf("ParseWeatherForecastToSimple failed: %v", err)
	}

	if len(simple) != 0 {
		t.Errorf("Expected 0 simple forecasts for empty data, got %d", len(simple))
	}
}

// Tests for ExtractCurrentWeather

func TestExtractCurrentWeather(t *testing.T) {
	forecast := createTestForecastResponse()

	current, err := ExtractCurrentWeather(forecast)
	if err != nil {
		t.Fatalf("ExtractCurrentWeather failed: %v", err)
	}

	if current == nil {
		t.Fatal("Expected current weather, got nil")
	}

	// Should return one of the weather entries (closest to current time)
	if current.Temperature < 20 || current.Temperature > 35 {
		t.Errorf("Temperature seems out of range: %.1f", current.Temperature)
	}
}

func TestExtractCurrentWeather_NoData(t *testing.T) {
	forecast := model.WeatherForecastResponse{
		Data: []model.ForecastData{},
	}

	_, err := ExtractCurrentWeather(forecast)
	if err == nil {
		t.Error("Expected error for empty forecast data, got nil")
	}
}

// Tests for FilterForecastByDate

func TestFilterForecastByDate(t *testing.T) {
	forecast := createTestForecastResponse()

	targetDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	filtered, err := FilterForecastByDate(forecast, targetDate)
	if err != nil {
		t.Fatalf("FilterForecastByDate failed: %v", err)
	}

	if len(filtered) != 3 {
		t.Errorf("Expected 3 forecasts for 2026-05-04, got %d", len(filtered))
	}

	// Check that all filtered items are from the target date
	for _, w := range filtered {
		localTime, _ := w.ParseLocalDateTime()
		if localTime.Year() != 2026 || localTime.Month() != 5 || localTime.Day() != 4 {
			t.Errorf("Filtered forecast has wrong date: %v", localTime)
		}
	}
}

func TestFilterForecastByDate_NoMatch(t *testing.T) {
	forecast := createTestForecastResponse()

	targetDate := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	_, err := FilterForecastByDate(forecast, targetDate)
	if err == nil {
		t.Error("Expected error for date with no data, got nil")
	}
}

// Tests for FilterForecastByDateRange

func TestFilterForecastByDateRange(t *testing.T) {
	forecast := createTestForecastResponse()

	startDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 5, 23, 59, 59, 0, time.UTC)

	filtered, err := FilterForecastByDateRange(forecast, startDate, endDate)
	if err != nil {
		t.Fatalf("FilterForecastByDateRange failed: %v", err)
	}

	if len(filtered) != 4 {
		t.Errorf("Expected 4 forecasts in range, got %d", len(filtered))
	}
}

// Tests for FormatEarthquakeForDisplay

func TestFormatEarthquakeForDisplay(t *testing.T) {
	eq := createTestEarthquake("5.7", "10 km", "-5.50,105.50", "Pusat gempa berada di laut 50 km barat daya Bengkulu")

	display, err := FormatEarthquakeForDisplay(eq)
	if err != nil {
		t.Fatalf("FormatEarthquakeForDisplay failed: %v", err)
	}

	if display.Magnitude != 5.7 {
		t.Errorf("Expected magnitude 5.7, got %.1f", display.Magnitude)
	}

	if display.Depth != 10 {
		t.Errorf("Expected depth 10, got %.1f", display.Depth)
	}

	if display.Location != "Pusat gempa berada di laut 50 km barat daya Bengkulu" {
		t.Errorf("Unexpected location: %s", display.Location)
	}

	if display.ShakemapURL == "" {
		t.Error("Expected shakemap URL, got empty string")
	}

	if display.Description == "" {
		t.Error("Expected description, got empty string")
	}
}

func TestFormatEarthquakeForDisplay_InvalidMagnitude(t *testing.T) {
	eq := createTestEarthquake("invalid", "10 km", "-5.50,105.50", "Test Region")

	display, err := FormatEarthquakeForDisplay(eq)
	if err != nil {
		t.Fatalf("FormatEarthquakeForDisplay failed: %v", err)
	}

	// Should default to 0 for invalid magnitude
	if display.Magnitude != 0 {
		t.Errorf("Expected magnitude 0 for invalid input, got %.1f", display.Magnitude)
	}
}

// Tests for CalculateEarthquakeDistance

func TestCalculateEarthquakeDistance(t *testing.T) {
	// Earthquake in Bengkulu area
	eq := createTestEarthquake("5.7", "10 km", "-3.8,102.3", "Bengkulu")

	// Calculate distance from Jakarta (-6.2, 106.8)
	distance, err := CalculateEarthquakeDistance(eq, -6.2, 106.8)
	if err != nil {
		t.Fatalf("CalculateEarthquakeDistance failed: %v", err)
	}

	// Distance should be roughly 500-600 km
	if distance < 400 || distance > 700 {
		t.Errorf("Distance seems incorrect: %.2f km", distance)
	}
}

func TestCalculateEarthquakeDistance_SameLocation(t *testing.T) {
	eq := createTestEarthquake("5.7", "10 km", "-6.2,106.8", "Jakarta")

	// Calculate distance from same coordinates
	distance, err := CalculateEarthquakeDistance(eq, -6.2, 106.8)
	if err != nil {
		t.Fatalf("CalculateEarthquakeDistance failed: %v", err)
	}

	// Distance should be very close to 0
	if distance > 1 {
		t.Errorf("Expected distance near 0, got %.2f km", distance)
	}
}

func TestCalculateEarthquakeDistance_InvalidCoordinates(t *testing.T) {
	eq := createTestEarthquake("5.7", "10 km", "invalid", "Test")

	_, err := CalculateEarthquakeDistance(eq, -6.2, 106.8)
	if err == nil {
		t.Error("Expected error for invalid coordinates, got nil")
	}
}

// Tests for FindNearestEarthquakes

func TestFindNearestEarthquakes(t *testing.T) {
	earthquakes := []model.EarthquakeDetail{
		createTestEarthquake("5.5", "10 km", "-6.0,106.5", "Near Jakarta"),
		createTestEarthquake("6.0", "15 km", "-3.8,102.3", "Bengkulu"),
		createTestEarthquake("4.5", "8 km", "-6.3,107.0", "Very Near Jakarta"),
	}

	// Find earthquakes within 100 km of Jakarta
	nearby, err := FindNearestEarthquakes(earthquakes, -6.2, 106.8, 100)
	if err != nil {
		t.Fatalf("FindNearestEarthquakes failed: %v", err)
	}

	// Should find 2 earthquakes (the ones near Jakarta)
	if len(nearby) != 2 {
		t.Errorf("Expected 2 nearby earthquakes, got %d", len(nearby))
	}

	// Check that results are sorted by distance
	if len(nearby) >= 2 && nearby[0].Distance > nearby[1].Distance {
		t.Error("Results should be sorted by distance (nearest first)")
	}
}

// Tests for GetWeatherSummary

func TestGetWeatherSummary(t *testing.T) {
	forecast := createTestForecastResponse()

	summary, err := GetWeatherSummary(forecast)
	if err != nil {
		t.Fatalf("GetWeatherSummary failed: %v", err)
	}

	if summary.Location == "" {
		t.Error("Expected location in summary, got empty string")
	}

	if summary.CurrentTemp < 20 || summary.CurrentTemp > 35 {
		t.Errorf("Temperature seems out of range: %.1f", summary.CurrentTemp)
	}

	if summary.Summary == "" {
		t.Error("Expected summary text, got empty string")
	}

	if summary.Condition == "" {
		t.Error("Expected condition, got empty string")
	}
}

// Tests for GetDailyWeatherSummary

func TestGetDailyWeatherSummary(t *testing.T) {
	forecast := createTestForecastResponse()

	daily, err := GetDailyWeatherSummary(forecast)
	if err != nil {
		t.Fatalf("GetDailyWeatherSummary failed: %v", err)
	}

	if len(daily) != 2 {
		t.Errorf("Expected 2 daily summaries, got %d", len(daily))
	}

	// Check first day
	if daily[0].MinTemp > daily[0].MaxTemp {
		t.Errorf("Min temp (%.1f) should be less than max temp (%.1f)", daily[0].MinTemp, daily[0].MaxTemp)
	}

	if len(daily[0].Conditions) == 0 {
		t.Error("Expected at least one condition in daily summary")
	}

	// Check that summaries are sorted by date
	if len(daily) >= 2 && daily[0].Date.After(daily[1].Date) {
		t.Error("Daily summaries should be sorted by date")
	}
}

// Tests for helper functions

func TestHaversineDistance(t *testing.T) {
	// Distance from Jakarta to Bandung (approximately 150 km)
	jakartaLat, jakartaLon := -6.2088, 106.8456
	bandungLat, bandungLon := -6.9175, 107.6191

	distance := haversineDistance(jakartaLat, jakartaLon, bandungLat, bandungLon)

	// Should be roughly 120-150 km
	if distance < 100 || distance > 180 {
		t.Errorf("Distance Jakarta-Bandung seems incorrect: %.2f km", distance)
	}
}

func TestHaversineDistance_SamePoint(t *testing.T) {
	distance := haversineDistance(-6.2, 106.8, -6.2, 106.8)

	if distance > 0.001 {
		t.Errorf("Distance between same points should be ~0, got %.6f", distance)
	}
}

func TestGetFeelsLikeDescription(t *testing.T) {
	tests := []struct {
		temp     float64
		humidity int
		expected string
	}{
		{35, 70, "Feels very hot and humid"},
		{30, 50, "Feels hot"},
		{26, 60, "Feels warm and comfortable"},
		{22, 50, "Feels pleasant"},
		{18, 40, "Feels cool"},
		{10, 30, "Feels cold"},
	}

	for _, tt := range tests {
		result := getFeelsLikeDescription(tt.temp, tt.humidity)
		if result != tt.expected {
			t.Errorf("getFeelsLikeDescription(%.1f, %d) = %s, expected %s",
				tt.temp, tt.humidity, result, tt.expected)
		}
	}
}

func TestGetWindDescription(t *testing.T) {
	tests := []struct {
		speed    float64
		dir      string
		contains string
	}{
		{3, "N", "Calm winds"},
		{15, "NE", "Light winds from NE"},
		{35, "SW", "Moderate winds from SW"},
		{55, "E", "Strong winds from E"},
		{70, "W", "Very strong winds from W"},
	}

	for _, tt := range tests {
		result := getWindDescription(tt.speed, tt.dir)
		if !contains([]string{result}, tt.contains) && len(tt.contains) > 0 {
			// Check if result contains expected substring
			found := false
			if len(result) >= len(tt.contains) {
				for i := 0; i <= len(result)-len(tt.contains); i++ {
					if result[i:i+len(tt.contains)] == tt.contains {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("getWindDescription(%.1f, %s) = %s, expected to contain %s",
					tt.speed, tt.dir, result, tt.contains)
			}
		}
	}
}

func TestGetRainChance(t *testing.T) {
	tests := []struct {
		weatherCode   int
		precipitation float64
		expected      string
	}{
		{model.WeatherCodeHeavyRain, 15, "Heavy rain expected"},
		{model.WeatherCodeModerateRain, 7, "Moderate rain expected"},
		{model.WeatherCodeLightRain, 2, "Light rain expected"},
		{model.WeatherCodeMostlyCloudy, 0, "Possible rain"},
		{model.WeatherCodeSunny, 0, "No rain expected"},
	}

	for _, tt := range tests {
		result := getRainChance(tt.weatherCode, tt.precipitation)
		if result != tt.expected {
			t.Errorf("getRainChance(%d, %.1f) = %s, expected %s",
				tt.weatherCode, tt.precipitation, result, tt.expected)
		}
	}
}

func TestIsSignificantEarthquake(t *testing.T) {
	tests := []struct {
		magnitude string
		expected  bool
	}{
		{"5.5", true},
		{"6.0", true},
		{"4.9", false},
		{"3.0", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		eq := createTestEarthquake(tt.magnitude, "10 km", "-5.0,105.0", "Test")
		result := IsSignificantEarthquake(eq)
		if result != tt.expected {
			t.Errorf("IsSignificantEarthquake(%s) = %v, expected %v",
				tt.magnitude, result, tt.expected)
		}
	}
}

func TestGetEarthquakeSeverity(t *testing.T) {
	tests := []struct {
		magnitude float64
		expected  string
	}{
		{2.5, "Minor"},
		{4.0, "Light"},
		{5.5, "Moderate"},
		{6.5, "Strong"},
		{7.5, "Major"},
		{8.5, "Great"},
	}

	for _, tt := range tests {
		result := GetEarthquakeSeverity(tt.magnitude)
		if result != tt.expected {
			t.Errorf("GetEarthquakeSeverity(%.1f) = %s, expected %s",
				tt.magnitude, result, tt.expected)
		}
	}
}

func TestGetWeatherIcon(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{model.WeatherCodeSunny, "☀️"},
		{model.WeatherCodePartlyCloudy, "⛅"},
		{model.WeatherCodeMostlyCloudy, "☁️"},
		{model.WeatherCodeFog, "🌫️"},
		{model.WeatherCodeLightRain, "🌧️"},
		{model.WeatherCodeThunderstorm, "⚡"},
		{999, "🌡️"}, // Unknown code
	}

	for _, tt := range tests {
		result := GetWeatherIcon(tt.code)
		if result != tt.expected {
			t.Errorf("GetWeatherIcon(%d) = %s, expected %s",
				tt.code, result, tt.expected)
		}
	}
}

// Benchmark tests

func BenchmarkParseWeatherForecastToSimple(b *testing.B) {
	forecast := createTestForecastResponse()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ParseWeatherForecastToSimple(forecast)
	}
}

func BenchmarkCalculateEarthquakeDistance(b *testing.B) {
	eq := createTestEarthquake("5.7", "10 km", "-3.8,102.3", "Bengkulu")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = CalculateEarthquakeDistance(eq, -6.2, 106.8)
	}
}

func BenchmarkHaversineDistance(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = haversineDistance(-6.2, 106.8, -3.8, 102.3)
	}
}

// Test edge cases

func TestHaversineDistance_Antipodes(t *testing.T) {
	// Test points on opposite sides of Earth
	distance := haversineDistance(0, 0, 0, 180)

	// Should be approximately half Earth's circumference (20,000 km)
	expectedDistance := math.Pi * 6371.0 // Half circumference
	tolerance := 100.0                   // 100 km tolerance

	if math.Abs(distance-expectedDistance) > tolerance {
		t.Errorf("Distance for antipodes: %.2f km, expected ~%.2f km", distance, expectedDistance)
	}
}

func TestFilterForecastByDateRange_EmptyRange(t *testing.T) {
	forecast := createTestForecastResponse()

	// End date before start date
	startDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	filtered, err := FilterForecastByDateRange(forecast, startDate, endDate)
	if err != nil {
		t.Fatalf("FilterForecastByDateRange failed: %v", err)
	}

	if len(filtered) != 0 {
		t.Errorf("Expected 0 forecasts for invalid range, got %d", len(filtered))
	}
}
