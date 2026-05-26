package domain

import "time"

type EvidenceStatus string

const (
	EvidencePending   EvidenceStatus = "pending"
	EvidenceRunning   EvidenceStatus = "running"
	EvidenceSatisfied EvidenceStatus = "satisfied"
	EvidenceFailed    EvidenceStatus = "failed"
	EvidenceWaived    EvidenceStatus = "waived"
)

type EvidenceExecution struct {
	ChangePacketID string            `json:"changePacketId"`
	Name           string            `json:"name"`
	Type           EvidenceType      `json:"type"`
	Status         EvidenceStatus    `json:"status"`
	Required       bool              `json:"required"`
	Summary        string            `json:"summary,omitempty"`
	DetailsURL     string            `json:"detailsUrl,omitempty"`
	UpdatedBy      string            `json:"updatedBy,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}
