package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"scraper-go/internal/model"
)

// SimpleWeatherForecast represents a simplified weather forecast structure
type SimpleWeatherForecast struct {
	Location    string    `json:"location"`
	Province    string    `json:"province"`
	DateTime    time.Time `json:"datetime"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	Weather     string    `json:"weather"`
	WeatherCode int       `json:"weather_code"`
	WindSpeed   float64   `json:"wind_speed"`
	WindDir     string    `json:"wind_direction"`
	Rainfall    float64   `json:"rainfall"`
}

// WeatherSummary represents a human-readable weather summary
type WeatherSummary struct {
	Location        string  `json:"location"`
	CurrentTemp     float64 `json:"current_temp"`
	FeelsLike       string  `json:"feels_like"`
	Condition       string  `json:"condition"`
	Humidity        int     `json:"humidity"`
	WindDescription string  `json:"wind_description"`
	RainChance      string  `json:"rain_chance"`
	Summary         string  `json:"summary"`
}

// EarthquakeDisplay represents formatted earthquake data for display
type EarthquakeDisplay struct {
	Time        string  `json:"time"`
	Location    string  `json:"location"`
	Magnitude   float64 `json:"magnitude"`
	Depth       float64 `json:"depth_km"`
	Coordinates string  `json:"coordinates"`
	Potential   string  `json:"potential"`
	ShakemapURL string  `json:"shakemap_url"`
	Description string  `json:"description"`
}

// EarthquakeDistance represents earthquake with calculated distance
type EarthquakeDistance struct {
	Earthquake model.EarthquakeDetail `json:"earthquake"`
	Distance   float64                `json:"distance_km"`
}

// ParseWeatherForecastToSimple converts complex forecast to simplified structure
// Returns a flat list of simple weather forecasts for easier consumption
func ParseWeatherForecastToSimple(forecast model.WeatherForecastResponse) ([]SimpleWeatherForecast, error) {
	var simpleForecasts []SimpleWeatherForecast

	for _, data := range forecast.Data {
		location := fmt.Sprintf("%s, %s", data.Location.Village, data.Location.District)
		if data.Location.Village == "" {
			location = data.Location.District
		}

		// Flatten the 3D weather array [day][time_slot]
		for _, dayForecasts := range data.Weather {
			for _, weather := range dayForecasts {
				localTime, err := weather.ParseLocalDateTime()
				if err != nil {
					// Skip invalid datetime entries
					continue
				}

				simple := SimpleWeatherForecast{
					Location:    location,
					Province:    data.Location.Province,
					DateTime:    localTime,
					Temperature: weather.Temperature,
					Humidity:    weather.Humidity,
					Weather:     weather.WeatherDescEn,
					WeatherCode: weather.Weather,
					WindSpeed:   weather.WindSpeed,
					WindDir:     weather.WindDirection,
					Rainfall:    weather.Precipitation,
				}
				simpleForecasts = append(simpleForecasts, simple)
			}
		}
	}

	return simpleForecasts, nil
}

// ExtractCurrentWeather gets the current or nearest weather from forecast
// Returns the weather info closest to the current time
func ExtractCurrentWeather(forecast model.WeatherForecastResponse) (*model.WeatherInfo, error) {
	if len(forecast.Data) == 0 {
		return nil, fmt.Errorf("no forecast data available")
	}

	now := time.Now()
	var closestWeather *model.WeatherInfo
	var minDiff time.Duration = time.Hour * 24 * 365 // Start with a large value

	for _, data := range forecast.Data {
		for _, dayForecasts := range data.Weather {
			for _, weather := range dayForecasts {
				localTime, err := weather.ParseLocalDateTime()
				if err != nil {
					continue
				}

				diff := localTime.Sub(now)
				if diff < 0 {
					diff = -diff
				}

				if diff < minDiff {
					minDiff = diff
					weatherCopy := weather
					closestWeather = &weatherCopy
				}
			}
		}
	}

	if closestWeather == nil {
		return nil, fmt.Errorf("no valid weather data found")
	}

	return closestWeather, nil
}

// FilterForecastByDate filters forecast by specific date
// Returns all weather forecasts for the given date
func FilterForecastByDate(forecast model.WeatherForecastResponse, targetDate time.Time) ([]model.WeatherInfo, error) {
	var filtered []model.WeatherInfo

	targetYear, targetMonth, targetDay := targetDate.Date()

	for _, data := range forecast.Data {
		for _, dayForecasts := range data.Weather {
			for _, weather := range dayForecasts {
				localTime, err := weather.ParseLocalDateTime()
				if err != nil {
					continue
				}

				year, month, day := localTime.Date()
				if year == targetYear && month == targetMonth && day == targetDay {
					filtered = append(filtered, weather)
				}
			}
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no forecast data found for date: %s", targetDate.Format("2006-01-02"))
	}

	return filtered, nil
}

// FilterForecastByDateRange filters forecast by date range
// Returns all weather forecasts between startDate and endDate (inclusive)
func FilterForecastByDateRange(forecast model.WeatherForecastResponse, startDate, endDate time.Time) ([]model.WeatherInfo, error) {
	var filtered []model.WeatherInfo

	for _, data := range forecast.Data {
		for _, dayForecasts := range data.Weather {
			for _, weather := range dayForecasts {
				localTime, err := weather.ParseLocalDateTime()
				if err != nil {
					continue
				}

				if (localTime.Equal(startDate) || localTime.After(startDate)) &&
					(localTime.Equal(endDate) || localTime.Before(endDate)) {
					filtered = append(filtered, weather)
				}
			}
		}
	}

	return filtered, nil
}

// FormatEarthquakeForDisplay formats earthquake data for human-readable output
func FormatEarthquakeForDisplay(eq model.EarthquakeDetail) (*EarthquakeDisplay, error) {
	// Parse magnitude
	magnitude, err := strconv.ParseFloat(eq.Magnitude, 64)
	if err != nil {
		magnitude = 0
	}

	// Parse depth (remove " km" suffix)
	depthStr := strings.TrimSuffix(strings.TrimSpace(eq.Depth), " km")
	depth, err := strconv.ParseFloat(depthStr, 64)
	if err != nil {
		depth = 0
	}

	// Format time
	eqTime, err := eq.ParseDateTime()
	timeStr := eq.Time
	if err == nil {
		timeStr = eqTime.Format("2006-01-02 15:04:05 MST")
	}

	// Create description
	description := fmt.Sprintf("Magnitude %.1f earthquake occurred at %s in %s at depth of %.0f km. %s",
		magnitude, timeStr, eq.Region, depth, eq.Potential)

	display := &EarthquakeDisplay{
		Time:        timeStr,
		Location:    eq.Region,
		Magnitude:   magnitude,
		Depth:       depth,
		Coordinates: eq.Coordinates,
		Potential:   eq.Potential,
		ShakemapURL: eq.GetShakemapURL(),
		Description: description,
	}

	return display, nil
}

// CalculateEarthquakeDistance calculates distance from given coordinates to earthquake
// Uses Haversine formula to calculate great-circle distance
// Returns distance in kilometers
func CalculateEarthquakeDistance(eq model.EarthquakeDetail, lat, lon float64) (float64, error) {
	// Parse earthquake coordinates
	coords := strings.Split(eq.Coordinates, ",")
	if len(coords) != 2 {
		return 0, fmt.Errorf("invalid coordinates format: %s", eq.Coordinates)
	}

	eqLat, err := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid latitude: %s", coords[0])
	}

	eqLon, err := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid longitude: %s", coords[1])
	}

	return haversineDistance(lat, lon, eqLat, eqLon), nil
}

// FindNearestEarthquakes finds earthquakes within a certain radius from coordinates
// Returns earthquakes sorted by distance (nearest first)
func FindNearestEarthquakes(earthquakes []model.EarthquakeDetail, lat, lon, radiusKm float64) ([]EarthquakeDistance, error) {
	var nearby []EarthquakeDistance

	for _, eq := range earthquakes {
		distance, err := CalculateEarthquakeDistance(eq, lat, lon)
		if err != nil {
			continue
		}

		if distance <= radiusKm {
			nearby = append(nearby, EarthquakeDistance{
				Earthquake: eq,
				Distance:   distance,
			})
		}
	}

	// Sort by distance (bubble sort for simplicity)
	for i := 0; i < len(nearby); i++ {
		for j := i + 1; j < len(nearby); j++ {
			if nearby[j].Distance < nearby[i].Distance {
				nearby[i], nearby[j] = nearby[j], nearby[i]
			}
		}
	}

	return nearby, nil
}

// GetWeatherSummary generates a human-readable weather summary
func GetWeatherSummary(forecast model.WeatherForecastResponse) (*WeatherSummary, error) {
	current, err := ExtractCurrentWeather(forecast)
	if err != nil {
		return nil, err
	}

	if len(forecast.Data) == 0 {
		return nil, fmt.Errorf("no location data available")
	}

	location := forecast.Data[0].Location
	locationStr := fmt.Sprintf("%s, %s", location.CityRegcy, location.Province)

	// Determine feels like description
	feelsLike := getFeelsLikeDescription(current.Temperature, current.Humidity)

	// Wind description
	windDesc := getWindDescription(current.WindSpeed, current.WindDirection)

	// Rain chance
	rainChance := getRainChance(current.Weather, current.Precipitation)

	// Generate summary text
	summary := fmt.Sprintf("Currently %s in %s with temperature %.1f°C. %s. %s. %s",
		strings.ToLower(current.WeatherDescEn),
		locationStr,
		current.Temperature,
		feelsLike,
		windDesc,
		rainChance,
	)

	return &WeatherSummary{
		Location:        locationStr,
		CurrentTemp:     current.Temperature,
		FeelsLike:       feelsLike,
		Condition:       current.WeatherDescEn,
		Humidity:        current.Humidity,
		WindDescription: windDesc,
		RainChance:      rainChance,
		Summary:         summary,
	}, nil
}

// GetDailyWeatherSummary generates daily min/max temperature and conditions
type DailyWeatherSummary struct {
	Date        time.Time `json:"date"`
	MinTemp     float64   `json:"min_temp"`
	MaxTemp     float64   `json:"max_temp"`
	AvgHumidity float64   `json:"avg_humidity"`
	Conditions  []string  `json:"conditions"`
	TotalRain   float64   `json:"total_rain_mm"`
}

// GetDailyWeatherSummary aggregates weather data by day
func GetDailyWeatherSummary(forecast model.WeatherForecastResponse) ([]DailyWeatherSummary, error) {
	dailyMap := make(map[string]*DailyWeatherSummary)

	for _, data := range forecast.Data {
		for _, dayForecasts := range data.Weather {
			for _, weather := range dayForecasts {
				localTime, err := weather.ParseLocalDateTime()
				if err != nil {
					continue
				}

				dateKey := localTime.Format("2006-01-02")

				if _, exists := dailyMap[dateKey]; !exists {
					dailyMap[dateKey] = &DailyWeatherSummary{
						Date:       time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, localTime.Location()),
						MinTemp:    weather.Temperature,
						MaxTemp:    weather.Temperature,
						Conditions: []string{},
						TotalRain:  0,
					}
				}

				summary := dailyMap[dateKey]

				// Update min/max temperature
				if weather.Temperature < summary.MinTemp {
					summary.MinTemp = weather.Temperature
				}
				if weather.Temperature > summary.MaxTemp {
					summary.MaxTemp = weather.Temperature
				}

				// Add humidity for averaging
				summary.AvgHumidity += float64(weather.Humidity)

				// Collect unique conditions
				if !contains(summary.Conditions, weather.WeatherDescEn) {
					summary.Conditions = append(summary.Conditions, weather.WeatherDescEn)
				}

				// Sum rainfall
				summary.TotalRain += weather.Precipitation
			}
		}
	}

	// Convert map to slice and calculate averages
	var dailySummaries []DailyWeatherSummary
	for _, summary := range dailyMap {
		// Calculate average humidity (rough estimate based on conditions count)
		if len(summary.Conditions) > 0 {
			summary.AvgHumidity = summary.AvgHumidity / float64(len(summary.Conditions))
		}
		dailySummaries = append(dailySummaries, *summary)
	}

	// Sort by date
	for i := 0; i < len(dailySummaries); i++ {
		for j := i + 1; j < len(dailySummaries); j++ {
			if dailySummaries[j].Date.Before(dailySummaries[i].Date) {
				dailySummaries[i], dailySummaries[j] = dailySummaries[j], dailySummaries[i]
			}
		}
	}

	return dailySummaries, nil
}

// Helper functions

// haversineDistance calculates the great-circle distance between two points
// on Earth using the Haversine formula. Returns distance in kilometers.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// getFeelsLikeDescription returns a description of how the temperature feels
func getFeelsLikeDescription(temp float64, humidity int) string {
	// Simple heat index approximation
	if temp >= 32 && humidity >= 60 {
		return "Feels very hot and humid"
	} else if temp >= 30 {
		return "Feels hot"
	} else if temp >= 25 {
		return "Feels warm and comfortable"
	} else if temp >= 20 {
		return "Feels pleasant"
	} else if temp >= 15 {
		return "Feels cool"
	}
	return "Feels cold"
}

// getWindDescription returns a human-readable wind description
func getWindDescription(speed float64, direction string) string {
	var speedDesc string

	if speed < 5 {
		speedDesc = "Calm winds"
	} else if speed < 20 {
		speedDesc = "Light winds"
	} else if speed < 40 {
		speedDesc = "Moderate winds"
	} else if speed < 60 {
		speedDesc = "Strong winds"
	} else {
		speedDesc = "Very strong winds"
	}

	if direction != "" {
		return fmt.Sprintf("%s from %s at %.1f km/h", speedDesc, direction, speed)
	}
	return fmt.Sprintf("%s at %.1f km/h", speedDesc, speed)
}

// getRainChance returns a description of rain probability
func getRainChance(weatherCode int, precipitation float64) string {
	if weatherCode >= model.WeatherCodeLightRain && weatherCode <= model.WeatherCodeThunderstorm {
		if precipitation > 10 {
			return "Heavy rain expected"
		} else if precipitation > 5 {
			return "Moderate rain expected"
		}
		return "Light rain expected"
	}

	if weatherCode == model.WeatherCodeMostlyCloudy || weatherCode == model.WeatherCodeOvercast {
		return "Possible rain"
	}

	return "No rain expected"
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsSignificantEarthquake determines if an earthquake is significant based on magnitude
func IsSignificantEarthquake(eq model.EarthquakeDetail) bool {
	magnitude, err := strconv.ParseFloat(eq.Magnitude, 64)
	if err != nil {
		return false
	}
	return magnitude >= 5.0
}

// GetEarthquakeSeverity returns severity level based on magnitude
func GetEarthquakeSeverity(magnitude float64) string {
	if magnitude < 3.0 {
		return "Minor"
	} else if magnitude < 5.0 {
		return "Light"
	} else if magnitude < 6.0 {
		return "Moderate"
	} else if magnitude < 7.0 {
		return "Strong"
	} else if magnitude < 8.0 {
		return "Major"
	}
	return "Great"
}

// GetWeatherIcon returns a simple text icon for weather condition
func GetWeatherIcon(weatherCode int) string {
	switch weatherCode {
	case model.WeatherCodeSunny:
		return "☀️"
	case model.WeatherCodePartlyCloudy:
		return "⛅"
	case model.WeatherCodeMostlyCloudy, model.WeatherCodeOvercast:
		return "☁️"
	case model.WeatherCodeFog, model.WeatherCodeHaze, model.WeatherCodeSmoke:
		return "🌫️"
	case model.WeatherCodeLightRain, model.WeatherCodeModerateRain:
		return "🌧️"
	case model.WeatherCodeHeavyRain, model.WeatherCodeIsolatedShower:
		return "⛈️"
	case model.WeatherCodeThunderstorm, model.WeatherCodeSevereThunderstorm:
		return "⚡"
	default:
		return "🌡️"
	}
}
