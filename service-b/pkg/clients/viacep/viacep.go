package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ViaCepResponse struct {
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro"`
}

type Client struct {
	URLTemplate string
}

func NewClient(urlTemplate string) *Client {
	return &Client{
		URLTemplate: urlTemplate,
	}
}

func (c *Client) GetCityByCep(ctx context.Context, cep string) (*ViaCepResponse, error) {
	url := fmt.Sprintf(c.URLTemplate, cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viacep api returned status %d", resp.StatusCode)
	}

	var result ViaCepResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Erro || result.Localidade == "" {
		return nil, fmt.Errorf("zipcode not found")
	}

	return &result, nil
}
