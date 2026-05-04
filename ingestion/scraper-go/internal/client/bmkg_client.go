package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ingestion/scraper-go/internal/model"

	"golang.org/x/time/rate"
)

// BMKGClient interface for dependency injection and testing
type BMKGClient interface {
	GetWeatherForecast(ctx context.Context, provinceID string) (*model.WeatherForecast, error)
	GetLatestEarthquake(ctx context.Context) (*model.AutoGempa, error)
	GetRecentEarthquakes(ctx context.Context) (*model.RecentEarthquakes, error)
	GetFeltEarthquakes(ctx context.Context) (*model.FeltEarthquakes, error)
}

// bmkgClient implements BMKGClient interface
type bmkgClient struct {
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	userAgent   string
	baseURLs    bmkgBaseURLs
}

// bmkgBaseURLs contains all BMKG API endpoints
type bmkgBaseURLs struct {
	weatherForecast    string
	latestEarthquake   string
	recentEarthquakes  string
	feltEarthquakes    string
}

// BMKGClientConfig holds configuration for BMKG client
type BMKGClientConfig struct {
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
	UserAgent       string
	RateLimit       rate.Limit // requests per second (60 req/min = 1 req/sec)
	RateBurst       int
}

// DefaultBMKGClientConfig returns default configuration
func DefaultBMKGClientConfig() *BMKGClientConfig {
	return &BMKGClientConfig{
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
		UserAgent:       "Mozilla/5.0 (compatible; BMKGDataCollector/1.0; +https://github.com/yourusername/datathon)",
		RateLimit:       rate.Limit(1.0), // 1 request per second (60/min)
		RateBurst:       2,                // allow burst of 2 requests
	}
}

// NewBMKGClient creates a new BMKG API client with proper configuration
func NewBMKGClient(config *BMKGClientConfig) BMKGClient {
	if config == nil {
		config = DefaultBMKGClientConfig()
	}

	return &bmkgClient{
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        config.MaxIdleConns,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     config.IdleConnTimeout,
				DisableCompression:  false,
			},
		},
		rateLimiter: rate.NewLimiter(config.RateLimit, config.RateBurst),
		userAgent:   config.UserAgent,
		baseURLs: bmkgBaseURLs{
			weatherForecast:   "https://api.bmkg.go.id/publik/prakiraan-cuaca",
			latestEarthquake:  "https://data.bmkg.go.id/DataMKG/TEWS/autogempa.json",
			recentEarthquakes: "https://data.bmkg.go.id/DataMKG/TEWS/gempaterkini.json",
			feltEarthquakes:   "https://data.bmkg.go.id/DataMKG/TEWS/gempadirasakan.json",
		},
	}
}

// GetWeatherForecast retrieves weather forecast for a specific province
// provinceID: optional province ID parameter (e.g., "31" for DKI Jakarta)
// If empty, returns forecast for all provinces
func (c *bmkgClient) GetWeatherForecast(ctx context.Context, provinceID string) (*model.WeatherForecast, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := c.baseURLs.weatherForecast
	if provinceID != "" {
		url = fmt.Sprintf("%s?adm4=%s", url, provinceID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create weather forecast request: %w", err)
	}

	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute weather forecast request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.validateResponse(resp); err != nil {
		return nil, fmt.Errorf("weather forecast API error: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read weather forecast response body: %w", err)
	}

	var forecast model.WeatherForecast
	if err := json.Unmarshal(body, &forecast); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weather forecast response: %w", err)
	}

	// Validate response data
	if len(forecast.Data) == 0 {
		return nil, fmt.Errorf("weather forecast response contains no data")
	}

	return &forecast, nil
}

// GetLatestEarthquake retrieves the latest earthquake information
func (c *bmkgClient) GetLatestEarthquake(ctx context.Context) (*model.AutoGempa, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURLs.latestEarthquake, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create latest earthquake request: %w", err)
	}

	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute latest earthquake request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.validateResponse(resp); err != nil {
		return nil, fmt.Errorf("latest earthquake API error: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read latest earthquake response body: %w", err)
	}

	var earthquake model.AutoGempa
	if err := json.Unmarshal(body, &earthquake); err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest earthquake response: %w", err)
	}

	// Validate response data
	if earthquake.InfoGempa.Gempa.Tanggal == "" {
		return nil, fmt.Errorf("latest earthquake response contains no valid data")
	}

	return &earthquake, nil
}

// GetRecentEarthquakes retrieves list of recent earthquakes
func (c *bmkgClient) GetRecentEarthquakes(ctx context.Context) (*model.RecentEarthquakes, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURLs.recentEarthquakes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create recent earthquakes request: %w", err)
	}

	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute recent earthquakes request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.validateResponse(resp); err != nil {
		return nil, fmt.Errorf("recent earthquakes API error: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read recent earthquakes response body: %w", err)
	}

	var earthquakes model.RecentEarthquakes
	if err := json.Unmarshal(body, &earthquakes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recent earthquakes response: %w", err)
	}

	// Validate response data
	if len(earthquakes.InfoGempa.Gempa) == 0 {
		return nil, fmt.Errorf("recent earthquakes response contains no data")
	}

	return &earthquakes, nil
}

// GetFeltEarthquakes retrieves list of earthquakes that were felt by people
func (c *bmkgClient) GetFeltEarthquakes(ctx context.Context) (*model.FeltEarthquakes, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURLs.feltEarthquakes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create felt earthquakes request: %w", err)
	}

	c.setCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute felt earthquakes request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.validateResponse(resp); err != nil {
		return nil, fmt.Errorf("felt earthquakes API error: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read felt earthquakes response body: %w", err)
	}

	var earthquakes model.FeltEarthquakes
	if err := json.Unmarshal(body, &earthquakes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal felt earthquakes response: %w", err)
	}

	// Validate response data
	if len(earthquakes.InfoGempa.Gempa) == 0 {
		return nil, fmt.Errorf("felt earthquakes response contains no data")
	}

	return &earthquakes, nil
}

// setCommonHeaders sets common HTTP headers for all requests
func (c *bmkgClient) setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
}

// validateResponse checks if the HTTP response is valid
func (c *bmkgClient) validateResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read error response body for detailed error message
	bodyBytes, _ := io.ReadAll(resp.Body)
	
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("bad request (400): %s", string(bodyBytes))
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized (401): check API credentials")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden (403): access denied")
	case http.StatusNotFound:
		return fmt.Errorf("not found (404): endpoint does not exist")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded (429): too many requests")
	case http.StatusInternalServerError:
		return fmt.Errorf("internal server error (500): %s", string(bodyBytes))
	case http.StatusServiceUnavailable:
		return fmt.Errorf("service unavailable (503): BMKG API is down")
	case http.StatusGatewayTimeout:
		return fmt.Errorf("gateway timeout (504): request took too long")
	default:
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}
}
