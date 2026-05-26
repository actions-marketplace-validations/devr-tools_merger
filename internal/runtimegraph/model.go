package runtimegraph

import "github.com/mergerhq/merger/internal/domain"

type NodeKind string

const (
	NodeService NodeKind = "service"
	NodeAPI     NodeKind = "api"
	NodeStore   NodeKind = "datastore"
	NodeQueue   NodeKind = "queue"
	NodeInfra   NodeKind = "infra"
	NodeTeam    NodeKind = "team"
)

type EdgeType string

const (
	EdgeCalls     EdgeType = "calls"
	EdgeReads     EdgeType = "reads"
	EdgePublishes EdgeType = "publishes"
	EdgeOwns      EdgeType = "owns"
	EdgeDeploys   EdgeType = "deploys"
)

type Node struct {
	ID          string             `json:"id"`
	Kind        NodeKind           `json:"kind"`
	Name        string             `json:"name"`
	Namespace   string             `json:"namespace,omitempty"`
	Owner       string             `json:"owner,omitempty"`
	Criticality domain.Criticality `json:"criticality,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

type Edge struct {
	From        string            `json:"from"`
	To          string            `json:"to"`
	Type        EdgeType          `json:"type"`
	Critical    bool              `json:"critical"`
	Directional bool              `json:"directional"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
