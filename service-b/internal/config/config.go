package config

import (
	"os"
	"regexp"
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
	matched, _ := regexp.MatchString(`^[a-z0-9]{31}$`, weatherAPIKey)
	if !matched {
		panic("The Weather API Key is missing. Acquire on at https://www.weatherapi.com and save it a `.env` file in the root of the project or invoke docker compose up using it as an environment variable. You can do it in the following way:\n\nWEATHER_API_KEY={API_KEY} docker compose up")
	}
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
		WeatherURLTemplate:   "https://api.weatherapi.com/v1/current.json?key=%s&q=%s",
	}
}
