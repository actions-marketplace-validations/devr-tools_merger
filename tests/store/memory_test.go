package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/store"
)

func TestMemoryRepositoryPersistsEvidenceFromChangePacket(t *testing.T) {
	repo := store.NewMemoryRepository()
	packet := domain.ChangePacket{
		ID:        "cp_1",
		UpdatedAt: time.Now().UTC(),
		Evidence: []domain.EvidenceRequirement{
			{Name: "integration_tests", Type: domain.EvidenceIntegrationTests, Required: true},
		},
	}

	if err := repo.SaveChangePacket(context.Background(), packet); err != nil {
		t.Fatalf("save change packet: %v", err)
	}

	executions, err := repo.ListEvidenceExecutions(context.Background(), "cp_1")
	if err != nil {
		t.Fatalf("list evidence executions: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("expected one evidence execution, got %d", len(executions))
	}
	if executions[0].Status != domain.EvidencePending {
		t.Fatalf("expected pending evidence status, got %s", executions[0].Status)
	}
}
