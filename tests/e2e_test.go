package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-a/pkg/servicea"
	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-b/pkg/serviceb"
)

func TestE2E(t *testing.T) {
	// 1. Mock ViaCEP
	viaCepMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		cep := ""
		for i, part := range parts {
			if part == "ws" && i+1 < len(parts) {
				cep = parts[i+1]
				break
			}
		}
		if cep == "01153000" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"localidade": "São Paulo", "erro": false}`)
		} else if cep == "00000000" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"erro": true}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer viaCepMock.Close()

	// 2. Mock WeatherAPI
	weatherApiMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "São Paulo" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"current": {"temp_c": 25.0}}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer weatherApiMock.Close()

	// 3. Setup Service B
	serviceBHandler := serviceb.NewHandler("test-key", viaCepMock.URL+"/ws/%s/json/", weatherApiMock.URL+"?key=%s&q=%s")
	serviceBMock := httptest.NewServer(serviceBHandler)
	defer serviceBMock.Close()

	// 4. Setup Service A
	serviceAHandler := servicea.NewHandler(serviceBMock.URL)
	serviceAMock := httptest.NewServer(serviceAHandler)
	defer serviceAMock.Close()

	tests := []struct {
		name           string
		cep            string
		expectedStatus int
		expectedCity   string
		expectedTempC  float64
	}{
		{
			name:           "Success",
			cep:            "01153000",
			expectedStatus: http.StatusOK,
			expectedCity:   "São Paulo",
			expectedTempC:  25.0,
		},
		{
			name:           "Invalid CEP",
			cep:            "123",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "CEP Not Found",
			cep:            "00000000",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(map[string]string{"cep": tt.cep})
			resp, err := http.Post(serviceAMock.URL, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var result struct {
					City  string  `json:"city"`
					TempC float64 `json:"temp_C"`
					TempF float64 `json:"temp_F"`
					TempK float64 `json:"temp_K"`
				}
				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if result.City != tt.expectedCity {
					t.Errorf("expected city %s, got %s", tt.expectedCity, result.City)
				}
				if result.TempC != tt.expectedTempC {
					t.Errorf("expected tempC %f, got %f", tt.expectedTempC, result.TempC)
				}
				// Verify conversions
				expectedF := tt.expectedTempC*1.8 + 32
				expectedK := tt.expectedTempC + 273.15
				if result.TempF != expectedF {
					t.Errorf("expected tempF %f, got %f", expectedF, result.TempF)
				}
				if result.TempK != expectedK {
					t.Errorf("expected tempK %f, got %f", expectedK, result.TempK)
				}
			} else {
				// Verify error message format
				var result map[string]string
				json.NewDecoder(resp.Body).Decode(&result)
				if _, ok := result["message"]; !ok {
					t.Errorf("expected message field in error response")
				}
			}
		})
	}
}
