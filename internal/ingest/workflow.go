package ingest

import (
	"context"

	"github.com/devr-tools/merger/internal/domain"
	"github.com/devr-tools/merger/internal/github"
)

func (p *Processor) processPROpened(ctx context.Context, payload github.PullRequestWebhookPayload) (*domain.ChangePacket, error) {
	ctx, span := p.tracer.Start(ctx, "ingest.process_pr_opened")
	defer span.End()

	githubService := bindInstallation(p.github, payload.Installation.ID)
	repoOwner := payload.Repository.Owner.Login
	repoName := payload.Repository.Name
	prNumber := payload.PullRequest.Number

	if err := p.publishPROpened(ctx, payload.Repository.FullName, prNumber, payload.Action); err != nil {
		return nil, err
	}

	pr, err := p.fetchPullRequest(ctx, githubService, repoOwner, repoName, prNumber, payload)
	if err != nil {
		return nil, err
	}

	packet, err := p.buildChangePacket(ctx, payload, pr, githubService, repoOwner, repoName, prNumber)
	if err != nil {
		return nil, err
	}

	if err := p.enrichMutations(ctx, packet, githubService, repoOwner, repoName); err != nil {
		return nil, err
	}
	if err := p.enrichRuntimeImpact(ctx, packet, githubService, repoOwner, repoName); err != nil {
		return nil, err
	}
	if err := p.enrichRisk(ctx, packet); err != nil {
		return nil, err
	}
	if err := p.applyPolicy(ctx, packet); err != nil {
		return nil, err
	}
	if err := p.assignMergeLane(ctx, packet); err != nil {
		return nil, err
	}
	if err := p.finalize(ctx, packet); err != nil {
		return nil, err
	}

	return packet, nil
}

func (p *Processor) fetchPullRequest(ctx context.Context, service github.Service, repoOwner, repoName string, prNumber int, payload github.PullRequestWebhookPayload) (github.PullRequest, error) {
	pr, err := service.GetPullRequest(ctx, repoOwner, repoName, prNumber)
	if err == nil {
		return pr, nil
	}

	return github.PullRequest{
		Owner:   repoOwner,
		Repo:    repoName,
		Number:  prNumber,
		Title:   payload.PullRequest.Title,
		Body:    payload.PullRequest.Body,
		Author:  payload.PullRequest.User.Login,
		URL:     payload.PullRequest.HTMLURL,
		HeadSHA: payload.PullRequest.Head.SHA,
		BaseSHA: payload.PullRequest.Base.SHA,
	}, nil
}
