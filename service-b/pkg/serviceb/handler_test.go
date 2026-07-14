package serviceb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-b/pkg/clients/viacep"
	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-b/pkg/clients/weather"
)

func TestHandler(t *testing.T) {
	// Mock ViaCEP
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

	// Mock WeatherAPI
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

	vcClient := viacep.NewClient(viaCepMock.URL + "/ws/%s/json/")
	wClient := weather.NewClient("test-key", weatherApiMock.URL+"?key=%s&q=%s")
	handler := NewHandler(vcClient, wClient)

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
			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d. Body: %s", tt.name, tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.expectedStatus == http.StatusOK {
				var result SuccessResponse
				err := json.NewDecoder(rr.Body).Decode(&result)
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
				expectedK := tt.expectedTempC + 273
				if result.TempF != expectedF {
					t.Errorf("expected tempF %f, got %f", expectedF, result.TempF)
				}
				if result.TempK != expectedK {
					t.Errorf("expected tempK %f, got %f", expectedK, result.TempK)
				}
			}
		})
	}
}
