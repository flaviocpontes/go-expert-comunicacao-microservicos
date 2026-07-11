package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type Client struct {
	APIKey      string
	URLTemplate string
}

func NewClient(apiKey, urlTemplate string) *Client {
	return &Client{
		APIKey:      apiKey,
		URLTemplate: urlTemplate,
	}
}

func (c *Client) GetWeather(ctx context.Context, city string) (*WeatherResponse, error) {
	encodedCity := url.QueryEscape(city)
	weatherURL := fmt.Sprintf(c.URLTemplate, c.APIKey, encodedCity)

	req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather api returned status %d", resp.StatusCode)
	}

	var result WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
