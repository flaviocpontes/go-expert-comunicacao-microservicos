package config

import (
	"os"
)

type Config struct {
	Port                 string
	OtelExporterEndpoint string
	ServiceBURL          string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "otel-collector:4317"
	}

	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://service-b:8081"
	}

	return &Config{
		Port:                 port,
		OtelExporterEndpoint: otelExporterEndpoint,
		ServiceBURL:          serviceBURL,
	}
}
