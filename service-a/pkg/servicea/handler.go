package servicea

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/flaviocpontes/go-expert-comunicacao-microservicos/service-a/internal/clients/serviceb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type CepRequest struct {
	Cep string `json:"cep"`
}

type ServiceAHandler struct {
	ServiceBClient *serviceb.Client
}

func NewHandler(serviceBClient *serviceb.Client) http.HandlerFunc {
	h := &ServiceAHandler{
		ServiceBClient: serviceBClient,
	}
	return h.Handle
}

func (h *ServiceAHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method not allowed"})
		return
	}

	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	tr := otel.Tracer("service-a")
	ctx, span := tr.Start(ctx, "receive-request")
	defer span.End()

	var req CepRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid zipcode"})
		return
	}

	matched, _ := regexp.MatchString(`^\d{8}$`, req.Cep)
	if !matched {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid zipcode"})
		return
	}

	// Forward to Service B
	statusCode, body, err := h.ServiceBClient.ForwardRequest(ctx, req.Cep)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}
