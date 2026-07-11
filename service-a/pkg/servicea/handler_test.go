package servicea

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-a/pkg/clients/serviceb"
)

func TestHandler(t *testing.T) {
	// Mock Service B
	serviceBMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Cep string `json:"cep"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Cep == "01153000" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"city": "São Paulo", "temp_C": 25.0, "temp_F": 77.0, "temp_K": 298.15}`))
		} else if req.Cep == "00000000" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "can not find zipcode"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer serviceBMock.Close()

	saClient := serviceb.NewClient(serviceBMock.URL)
	handler := NewHandler(saClient)

	tests := []struct {
		name           string
		cep            string
		expectedStatus int
	}{
		{
			name:           "Success",
			cep:            "01153000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid CEP (too short)",
			cep:            "123",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Invalid CEP (alpha)",
			cep:            "ABCDEFGH",
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
		})
	}
}
