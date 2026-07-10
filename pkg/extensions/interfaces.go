package extensions

import (
	"context"

	"github.com/devr-tools/merger/pkg/merger"
)

type EventHandler func(context.Context, merger.Envelope) error

type EventBus interface {
	Publish(context.Context, merger.Envelope) error
	Subscribe(merger.EventType, EventHandler) error
	Close() error
}

type EventStore interface {
	SaveEvent(context.Context, merger.Envelope) error
}

type ChangePacketStore interface {
	SaveChangePacket(context.Context, merger.ChangePacket) error
}

type PullRequest struct {
	Owner   string
	Repo    string
	Number  int
	Title   string
	Body    string
	Author  string
	URL     string
	HeadSHA string
	BaseSHA string
}

type CheckRunInput struct {
	RepoOwner  string
	RepoName   string
	HeadSHA    string
	Name       string
	Status     string
	Conclusion string
	Summary    string
	DetailsURL string
}

type SCMProvider interface {
	GetPullRequest(context.Context, string, string, int) (PullRequest, error)
	GetPullRequestDiff(context.Context, string, string, int) (string, error)
	GetFileContent(context.Context, string, string, string, string) ([]byte, error)
	PublishCheckRun(context.Context, CheckRunInput) error
}

type MutationContext struct {
	Repo    merger.RepoRef
	Ref     string
	File    merger.ChangedFile
	Content []byte
}

type MutationAnalyzer interface {
	Name() string
	Supports(merger.ChangedFile) bool
	Analyze(context.Context, MutationContext) ([]merger.Mutation, error)
}

type GraphSourceInput struct {
	Repo  merger.RepoRef
	Ref   string
	Files []merger.ChangedFile
	Load  func(context.Context, string) ([]byte, error)
}

type GraphFragment struct {
	Nodes       []merger.GraphNode
	Edges       []merger.GraphEdge
	Systems     []merger.SystemRef
	Ownership   []merger.OwnershipBoundary
	Notes       []string
	Criticality merger.Criticality
}

type RuntimeGraphSource interface {
	Name() string
	Collect(context.Context, GraphSourceInput) (GraphFragment, error)
}
