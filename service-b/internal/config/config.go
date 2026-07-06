package config

import (
	"os"
)

type Config struct {
	WeatherAPIKey        string
	Port                 string
	OtelExporterEndpoint string
	ViaCepURLTemplate    string
	WeatherURLTemplate   string
}

func LoadConfig() *Config {
	weatherAPIKey := os.Getenv("WEATHER_API_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "otel-collector:4317"
	}

	return &Config{
		WeatherAPIKey:        weatherAPIKey,
		Port:                 port,
		OtelExporterEndpoint: otelExporterEndpoint,
		ViaCepURLTemplate:    "https://viacep.com.br/ws/%s/json/",
		WeatherURLTemplate:   "http://api.weatherapi.com/v1/current.json?key=%s&q=%s",
	}
}
