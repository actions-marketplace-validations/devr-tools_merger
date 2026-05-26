package controlplane_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mergerhq/merger/internal/controlplane"
	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/store"
)

func TestHTTPHandlerReturnsChangePacketView(t *testing.T) {
	repo := seedRepository(t)
	handler := controlplane.NewHTTPHandler(controlplane.NewService(repo))
	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/change-packets/cp_1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	packet := payload["packet"].(map[string]any)
	if packet["id"] != "cp_1" {
		t.Fatalf("unexpected packet payload: %#v", payload)
	}
}

func TestHTTPHandlerUpdatesEvidenceExecution(t *testing.T) {
	repo := seedRepository(t)
	handler := controlplane.NewHTTPHandler(controlplane.NewService(repo))
	mux := http.NewServeMux()
	handler.Register(mux)

	body := bytes.NewBufferString(`{"status":"satisfied","summary":"tests passed","updatedBy":"ci","type":"integration_tests","required":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/change-packets/cp_1/evidence/integration_tests", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	executions, err := repo.ListEvidenceExecutions(context.Background(), "cp_1")
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	if executions[0].Status != domain.EvidenceSatisfied {
		t.Fatalf("expected satisfied evidence, got %s", executions[0].Status)
	}
}

func seedRepository(t *testing.T) *store.MemoryRepository {
	t.Helper()

	repo := store.NewMemoryRepository()
	packet := domain.ChangePacket{
		ID:        "cp_1",
		Repo:      domain.RepoRef{FullName: "acme/repo"},
		PR:        domain.PullRequestRef{Number: 42},
		MergeLane: domain.MergeLaneYellow,
		RiskSummary: domain.RiskSummary{
			Score: 30,
		},
		UpdatedAt: time.Now().UTC(),
		Evidence: []domain.EvidenceRequirement{
			{Name: "integration_tests", Type: domain.EvidenceIntegrationTests, Required: true},
		},
	}
	if err := repo.SaveChangePacket(context.Background(), packet); err != nil {
		t.Fatalf("seed change packet: %v", err)
	}
	return repo
}
