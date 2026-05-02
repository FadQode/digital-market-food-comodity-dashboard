package client 

import {
	"bytes"
	"encoding/json"
	"net/http"
	"os"
}

type ApifyTokpedClient struct {
	ActorID string `"json:"actorId"`
	input map[string]any  `json:"input"`
}

func RunTokopediaScraper(keyword string)([]bytes, error) {

	apikey := os.Getenv("APIFY_TOKOPEDIA_TOKEN")

	url := "https://api.apify.com/v2/acts/jupri~tokopedia-scraper/run-sync-get-dataset-items?token=" + apikey
	
	payload := ApifyTokpedClient{
		ActorID: "jupri~tokopedia-scraper",
		input: map[string]any{
			"search": keyword,
			"maxItems": 10,
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonPayload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return jsonPayload, nil
}