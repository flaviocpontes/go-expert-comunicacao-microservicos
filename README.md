# Distributed Tracing with OpenTelemetry and Zipkin

This project consists of two microservices (Service A and Service B) that work together to query the weather of a city based on its zip code (CEP).

## Architecture

- **Service A (Input)**: Receives the user's request, validates the CEP, and forwards it to Service B.
- **Service B (Orchestration)**: Receives the CEP, identifies the city (via ViaCEP), queries the current temperature (via WeatherAPI), and performs temperature conversions (Celsius, Fahrenheit, Kelvin).
- **OTEL Collector**: Collects traces from both services and sends them to Zipkin.
- **Zipkin**: Visualizes the distributed traces.

## Prerequisites

- Docker and Docker Compose
- [WeatherAPI](https://www.weatherapi.com/) Key

## How to Run

1. Clone the repository.
2. Create a `.env` file at the root or set the `WEATHER_API_KEY` environment variable.
   ```bash
   export WEATHER_API_KEY=your_api_key_here
   ```
3. Run the services using Docker Compose:
   ```bash
   docker-compose up --build
   ```

## How to Test

### Service A

Send a POST request to Service A at `http://localhost:8080/` with a JSON payload containing the CEP.

**Valid CEP:**
```bash
curl -X POST http://localhost:8080/ -d '{"cep": "01153000"}'
```
Response:
```json
{"city": "São Paulo", "temp_C": 25.0, "temp_F": 77.0, "temp_K": 298.15}
```

**Invalid CEP (Format):**
```bash
curl -i -X POST http://localhost:8080/ -d '{"cep": "123"}'
```
Response: `422 Unprocessable Entity` with message `{"message": "invalid zipcode"}`.

**CEP Not Found:**
```bash
curl -i -X POST http://localhost:8080/ -d '{"cep": "00000000"}'
```
Response: `404 Not Found` with message `{"message": "can not find zipcode"}`.

### Zipkin Tracing

Access Zipkin at `http://localhost:9411` to visualize the traces.
Search for traces to see the flow from `service-a` to `service-b`, including the spans for external API calls (`get-location` and `get-weather`).

## Automated Tests

The project includes automated end-to-end tests that mock external dependencies (ViaCEP and WeatherAPI).

To run the tests, ensure you are at the project root and execute:

```bash
go test ./service-a/... ./service-b/...
```

These tests verify the entire flow from Service A to Service B, including:
- Success cases with valid CEP.
- Error handling for invalid CEP formats (422).
- Error handling for non-existent CEPs (404).

## Temperature Conversions

- **Fahrenheit:** `F = C * 1.8 + 32`
- **Kelvin:** `K = C + 273`
