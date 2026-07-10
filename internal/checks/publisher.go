package checks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/github"
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

	client := p.client
	if binder, ok := p.client.(github.InstallationBinder); ok {
		if rawInstallationID := packet.Metadata["installation_id"]; rawInstallationID != "" {
			installationID, err := strconv.ParseInt(rawInstallationID, 10, 64)
			if err == nil {
				client = binder.ForInstallation(installationID)
			}
		}
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

	return client.PublishCheckRun(ctx, github.CheckRunInput{
		RepoOwner:  packet.Repo.Owner,
		RepoName:   packet.Repo.Name,
		HeadSHA:    packet.PR.HeadSHA,
		Name:       "merger/change-control",
		Status:     status,
		Conclusion: conclusion,
		Summary:    fmt.Sprintf("lane=%s risk=%d mutations=%d policies=%d", packet.MergeLane, packet.RiskSummary.Score, len(packet.Mutations), len(packet.Decision.AppliedPolicies)),
	})
}
