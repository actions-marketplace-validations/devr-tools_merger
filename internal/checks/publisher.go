package checks

import (
	"context"
	"fmt"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/github"
)

type Publisher interface {
	Publish(context.Context, domain.ChangePacket) error
}

type GitHubCheckPublisher struct {
	client github.CheckRunPublisher
}

func NewGitHubCheckPublisher(client github.CheckRunPublisher) *GitHubCheckPublisher {
	return &GitHubCheckPublisher{client: client}
}

func (p *GitHubCheckPublisher) Publish(ctx context.Context, packet domain.ChangePacket) error {
	if p.client == nil {
		return nil
	}

	status := "completed"
	conclusion := "neutral"
	switch packet.MergeLane {
	case domain.MergeLaneGreen:
		conclusion = "success"
	case domain.MergeLaneBlack:
		conclusion = "action_required"
	case domain.MergeLaneRed:
		conclusion = "neutral"
	}

	return p.client.PublishCheckRun(ctx, github.CheckRunInput{
		RepoOwner:  packet.Repo.Owner,
		RepoName:   packet.Repo.Name,
		HeadSHA:    packet.PR.HeadSHA,
		Name:       "merger/change-control",
		Status:     status,
		Conclusion: conclusion,
		Summary:    fmt.Sprintf("lane=%s risk=%d mutations=%d policies=%d", packet.MergeLane, packet.RiskSummary.Score, len(packet.Mutations), len(packet.Decision.AppliedPolicies)),
	})
}
