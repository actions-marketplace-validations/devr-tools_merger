package ingest

import (
	"encoding/json"
	"net/http"

	"github.com/mergerhq/merger/internal/github"
	"github.com/mergerhq/merger/internal/telemetry"
	"github.com/mergerhq/merger/pkg/identity"
)

type WebhookHandler struct {
	processor *Processor
}

func NewWebhookHandler(processor *Processor) *WebhookHandler {
	return &WebhookHandler{processor: processor}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := telemetry.WithRequestID(r.Context(), identity.New("req"))

	webhook, err := github.DecodeWebhookRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx = telemetry.WithCorrelationID(ctx, webhook.DeliveryID)
	if webhook.Event != "pull_request" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	switch webhook.Payload.Action {
	case "opened", "reopened", "synchronize":
	default:
		w.WriteHeader(http.StatusAccepted)
		return
	}

	packet, err := h.processor.ProcessPROpened(ctx, webhook.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"changePacketId": packet.ID,
		"mergeLane":      packet.MergeLane,
		"riskScore":      packet.RiskSummary.Score,
	})
}
