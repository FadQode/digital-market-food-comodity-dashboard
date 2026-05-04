package model

import "time"

// WeatherForecastResponse represents the complete response from BMKG weather forecast API
// API Endpoint: https://api.bmkg.go.id/publik/prakiraan-cuaca?adm4={kode_wilayah}
type WeatherForecastResponse struct {
	Location Location       `json:"lokasi"`
	Data     []ForecastData `json:"data"`
}

// Location represents the geographical location information
type Location struct {
	Adm1      string  `json:"adm1"`                // Administrative level 1 code (Province)
	Adm2      string  `json:"adm2"`                // Administrative level 2 code (City/Regency)
	Adm3      string  `json:"adm3"`                // Administrative level 3 code (District)
	Adm4      string  `json:"adm4"`                // Administrative level 4 code (Village/Sub-district)
	Province  string  `json:"provinsi"`            // Province name
	CityRegcy string  `json:"kotkab"`              // City or Regency name
	District  string  `json:"kecamatan"`           // District name
	Village   string  `json:"desa"`                // Village or Sub-district name
	Longitude float64 `json:"lon"`                 // Longitude coordinate
	Latitude  float64 `json:"lat"`                 // Latitude coordinate
	Timezone  string  `json:"timezone"`            // Timezone (e.g., "Asia/Jakarta" or "+0700")
	Type      *string `json:"type,omitempty"`      // Location type (e.g., "adm4")
}

// ForecastData represents weather forecast data for a location
type ForecastData struct {
	Location Location        `json:"lokasi"`
	Weather  [][]WeatherInfo `json:"cuaca"` // 3D array: [day][time_slot]weather_info
}

// WeatherInfo represents detailed weather information for a specific time
type WeatherInfo struct {
	DateTime        string   `json:"datetime"`          // UTC datetime in ISO 8601 format
	UTCDateTime     string   `json:"utc_datetime"`      // UTC datetime (YYYY-MM-DD HH:mm:ss)
	LocalDateTime   string   `json:"local_datetime"`    // Local datetime (YYYY-MM-DD HH:mm:ss)
	Temperature     float64  `json:"t"`                 // Temperature in Celsius (°C)
	Humidity        int      `json:"hu"`                // Humidity in percentage (%)
	TotalCloudCover int      `json:"tcc"`               // Total cloud cover in percentage (%)
	Precipitation   float64  `json:"tp"`                // Total precipitation in mm
	Weather         int      `json:"weather"`           // Weather code
	WeatherDesc     string   `json:"weather_desc"`      // Weather description in Indonesian
	WeatherDescEn   string   `json:"weather_desc_en"`   // Weather description in English
	WindDegree      int      `json:"wd_deg"`            // Wind direction in degrees
	WindDirection   string   `json:"wd"`                // Wind direction from (N, NE, E, SE, S, SW, W, NW)
	WindDirectionTo string   `json:"wd_to"`             // Wind direction to
	WindSpeed       float64  `json:"ws"`                // Wind speed in km/h
	Visibility      int      `json:"vs"`                // Visibility in meters
	VisibilityText  string   `json:"vs_text"`           // Visibility description (e.g., "> 10 km")
	TimeIndex       string   `json:"time_index"`        // Time index range (e.g., "7-8")
	AnalysisDate    string   `json:"analysis_date"`     // Analysis/production date in ISO 8601 format
	Image           string   `json:"image"`             // URL to weather icon image
}

// EarthquakeLatestResponse represents the latest earthquake data
// API Endpoint: https://data.bmkg.go.id/DataMKG/TEWS/autogempa.json
type EarthquakeLatestResponse struct {
	InfoEarthquake InfoEarthquake `json:"Infogempa"`
}

// InfoEarthquake wraps earthquake information
type InfoEarthquake struct {
	Earthquake EarthquakeDetail `json:"gempa"`
}

// EarthquakeDetail represents detailed earthquake information
type EarthquakeDetail struct {
	Date        string  `json:"Tanggal"`     // Date in Indonesian format (DD MMM YYYY)
	Time        string  `json:"Jam"`         // Time in WIB (HH:mm:ss WIB)
	DateTime    string  `json:"DateTime"`    // ISO 8601 datetime in UTC
	Coordinates string  `json:"Coordinates"` // Coordinates as "lat,lon" string
	Latitude    string  `json:"Lintang"`     // Latitude with direction (e.g., "4.74 LU")
	Longitude   string  `json:"Bujur"`       // Longitude with direction (e.g., "96.77 BT")
	Magnitude   string  `json:"Magnitude"`   // Earthquake magnitude
	Depth       string  `json:"Kedalaman"`   // Depth in km
	Region      string  `json:"Wilayah"`     // Region/area description
	Potential   string  `json:"Potensi"`     // Tsunami potential or felt status
	Felt        *string `json:"Dirasakan,omitempty"` // Areas where earthquake was felt (MMI scale)
	Shakemap    string  `json:"Shakemap"`    // Shakemap image filename
}

// EarthquakeListResponse represents a list of recent earthquakes
// API Endpoints:
// - https://data.bmkg.go.id/DataMKG/TEWS/gempaterkini.json (M 5.0+)
// - https://data.bmkg.go.id/DataMKG/TEWS/gempadirasakan.json (Felt earthquakes)
type EarthquakeListResponse struct {
	InfoEarthquake InfoEarthquakeList `json:"Infogempa"`
}

// InfoEarthquakeList wraps a list of earthquakes
type InfoEarthquakeList struct {
	Earthquakes []EarthquakeDetail `json:"gempa"`
}

// WeatherWarningRSSFeed represents the RSS feed for weather warnings
// API Endpoints:
// - https://www.bmkg.go.id/alerts/nowcast/id (Indonesian)
// - https://www.bmkg.go.id/alerts/nowcast/en (English)
type WeatherWarningRSSFeed struct {
	Channel WeatherWarningChannel `xml:"channel"`
}

// WeatherWarningChannel represents the RSS channel
type WeatherWarningChannel struct {
	Title         string                `xml:"title"`         // Channel title
	Link          string                `xml:"link"`          // Channel link
	Description   string                `xml:"description"`   // Channel description
	Language      string                `xml:"language"`      // Language (id/en)
	Copyright     string                `xml:"copyright"`     // Copyright information
	PubDate       string                `xml:"pubDate"`       // Publication date (RFC 1123)
	LastBuildDate string                `xml:"lastBuildDate"` // Last update date (RFC 1123)
	Items         []WeatherWarningItem  `xml:"item"`          // Warning items
}

// WeatherWarningItem represents a single weather warning in RSS feed
type WeatherWarningItem struct {
	Title       string `xml:"title"`       // Warning title (province)
	Link        string `xml:"link"`        // Link to detailed CAP XML
	Description string `xml:"description"` // Description of affected areas
	Author      string `xml:"author"`      // Author/issuer
	PubDate     string `xml:"pubDate"`     // Publication date (RFC 1123)
}

// WeatherWarningCAP represents the Common Alerting Protocol (CAP) format
// for detailed weather warnings
// API Endpoints:
// - https://www.bmkg.go.id/alerts/nowcast/id/{code}_alert.xml
// - https://www.bmkg.go.id/alerts/nowcast/en/{code}_alert.xml
type WeatherWarningCAP struct {
	Identifier  string                `xml:"identifier"`  // Unique alert identifier
	Sender      string                `xml:"sender"`      // Sender identifier
	Sent        string                `xml:"sent"`        // Sent time (ISO 8601)
	Status      string                `xml:"status"`      // Alert status (Actual, Exercise, System, Test)
	MsgType     string                `xml:"msgType"`     // Message type (Alert, Update, Cancel, etc.)
	Scope       string                `xml:"scope"`       // Alert scope (Public, Restricted, Private)
	Info        []WeatherWarningInfo  `xml:"info"`        // Alert information (can be multiple for different languages)
}

// WeatherWarningInfo represents detailed information in CAP alert
type WeatherWarningInfo struct {
	Language     string                    `xml:"language"`     // Language code (id-ID, en-US)
	Category     string                    `xml:"category"`     // Event category (Met, Geo, Safety, etc.)
	Event        string                    `xml:"event"`        // Event type description
	Urgency      string                    `xml:"urgency"`      // Urgency level (Immediate, Expected, Future, Past, Unknown)
	Severity     string                    `xml:"severity"`     // Severity level (Extreme, Severe, Moderate, Minor, Unknown)
	Certainty    string                    `xml:"certainty"`    // Certainty level (Observed, Likely, Possible, Unlikely, Unknown)
	Effective    string                    `xml:"effective"`    // Effective time (ISO 8601)
	Expires      string                    `xml:"expires"`      // Expiration time (ISO 8601)
	SenderName   string                    `xml:"senderName"`   // Sender name
	Headline     string                    `xml:"headline"`     // Alert headline
	Description  string                    `xml:"description"`  // Detailed description
	Web          string                    `xml:"web"`          // Web link to infographic
	Areas        []WeatherWarningArea      `xml:"area"`         // Affected areas
}

// WeatherWarningArea represents an affected geographical area
type WeatherWarningArea struct {
	AreaDesc string `xml:"areaDesc"` // Area description (district/sub-district name)
	Polygon  string `xml:"polygon"`  // Polygon coordinates defining the area
}

// Helper methods for parsing

// ParseDateTime parses the DateTime field to time.Time
func (e *EarthquakeDetail) ParseDateTime() (time.Time, error) {
	return time.Parse(time.RFC3339, e.DateTime)
}

// ParseLocalDateTime parses the local_datetime field to time.Time
func (w *WeatherInfo) ParseLocalDateTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", w.LocalDateTime)
}

// ParseUTCDateTime parses the utc_datetime field to time.Time
func (w *WeatherInfo) ParseUTCDateTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", w.UTCDateTime)
}

// GetShakemapURL returns the full URL to the shakemap image
func (e *EarthquakeDetail) GetShakemapURL() string {
	if e.Shakemap == "" {
		return ""
	}
	return "https://static.bmkg.go.id/" + e.Shakemap
}

// Weather code constants based on BMKG API
const (
	WeatherCodeSunny          = 1  // Cerah / Sunny
	WeatherCodePartlyCloudy   = 2  // Cerah Berawan / Partly Cloudy
	WeatherCodeMostlyCloudy   = 3  // Berawan / Mostly Cloudy
	WeatherCodeOvercast       = 4  // Berawan Tebal / Overcast
	WeatherCodeHaze           = 5  // Udara Kabur / Haze
	WeatherCodeSmoke          = 10 // Asap / Smoke
	WeatherCodeFog            = 45 // Kabut / Fog
	WeatherCodeLightRain      = 60 // Hujan Ringan / Light Rain
	WeatherCodeModerateRain   = 61 // Hujan Sedang / Moderate Rain
	WeatherCodeHeavyRain      = 63 // Hujan Lebat / Heavy Rain
	WeatherCodeIsolatedShower = 80 // Hujan Lokal / Isolated Shower
	WeatherCodeSevereThunderstorm = 95 // Hujan Petir / Severe Thunderstorm
	WeatherCodeThunderstorm   = 97 // Hujan Petir / Thunderstorm
)
