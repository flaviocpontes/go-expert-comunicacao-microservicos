package serviceb

import (
	"encoding/json"
	"net/http"

	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-b/pkg/clients/viacep"
	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-b/pkg/clients/weather"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type CepRequest struct {
	Cep string `json:"cep"`
}

type SuccessResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ServiceBHandler struct {
	ViaCepClient  *viacep.Client
	WeatherClient *weather.Client
}

func NewHandler(viaCepClient *viacep.Client, weatherClient *weather.Client) http.HandlerFunc {
	h := &ServiceBHandler{
		ViaCepClient:  viaCepClient,
		WeatherClient: weatherClient,
	}
	return h.Handle
}

func (h *ServiceBHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	tr := otel.Tracer("service-b")
	ctx, span := tr.Start(ctx, "process-request")
	defer span.End()

	var req CepRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	if len(req.Cep) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// 1. Get Location
	locationSpanCtx, locationSpan := tr.Start(ctx, "get-location")
	viaCepResp, err := h.ViaCepClient.GetCityByCep(locationSpanCtx, req.Cep)
	locationSpan.End()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "can not find zipcode"})
		return
	}

	// 2. Get Weather
	weatherSpanCtx, weatherSpan := tr.Start(ctx, "get-weather")
	weatherResp, err := h.WeatherClient.GetWeather(weatherSpanCtx, viaCepResp.Localidade)
	weatherSpan.End()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "failed to get weather"})
		return
	}

	// 3. Convert and Respond
	tempC := weatherResp.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	res := SuccessResponse{
		City:  viaCepResp.Localidade,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
