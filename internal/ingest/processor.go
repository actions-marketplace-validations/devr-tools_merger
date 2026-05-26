package ingest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mergerhq/merger/internal/checks"
	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/events"
	"github.com/mergerhq/merger/internal/github"
	"github.com/mergerhq/merger/internal/lanes"
	"github.com/mergerhq/merger/internal/mutations"
	"github.com/mergerhq/merger/internal/policy"
	"github.com/mergerhq/merger/internal/risk"
	"github.com/mergerhq/merger/internal/runtimegraph"
	"github.com/mergerhq/merger/internal/telemetry"
	"github.com/mergerhq/merger/pkg/diff"
	"github.com/mergerhq/merger/pkg/identity"
)

type Processor struct {
	logger    *telemetry.Logger
	tracer    telemetry.Tracer
	bus       events.Bus
	github    github.Service
	mutations mutations.Engine
	risk      risk.Engine
	policy    policy.Engine
	assigner  lanes.Assigner
	checks    checks.Publisher
	runtime   runtimegraph.Resolver
}

func NewProcessor(
	logger *telemetry.Logger,
	tracer telemetry.Tracer,
	bus events.Bus,
	githubService github.Service,
	mutationEngine mutations.Engine,
	riskEngine risk.Engine,
	policyEngine policy.Engine,
	assigner lanes.Assigner,
	checkPublisher checks.Publisher,
	runtimeResolver runtimegraph.Resolver,
) *Processor {
	return &Processor{
		logger:    logger,
		tracer:    tracer,
		bus:       bus,
		github:    githubService,
		mutations: mutationEngine,
		risk:      riskEngine,
		policy:    policyEngine,
		assigner:  assigner,
		checks:    checkPublisher,
		runtime:   runtimeResolver,
	}
}

func (p *Processor) ProcessPROpened(ctx context.Context, payload github.PullRequestWebhookPayload) (*domain.ChangePacket, error) {
	ctx, span := p.tracer.Start(ctx, "ingest.process_pr_opened")
	defer span.End()

	repoOwner := payload.Repository.Owner.Login
	repoName := payload.Repository.Name
	prNumber := payload.PullRequest.Number

	if err := p.bus.Publish(ctx, events.NewEnvelope(events.EventPROpened, "ingest", map[string]any{
		"repo":     payload.Repository.FullName,
		"prNumber": prNumber,
		"action":   payload.Action,
	})); err != nil {
		return nil, err
	}

	pr, err := p.github.GetPullRequest(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		pr = github.PullRequest{
			Owner:   repoOwner,
			Repo:    repoName,
			Number:  prNumber,
			Title:   payload.PullRequest.Title,
			Body:    payload.PullRequest.Body,
			Author:  payload.PullRequest.User.Login,
			URL:     payload.PullRequest.HTMLURL,
			HeadSHA: payload.PullRequest.Head.SHA,
			BaseSHA: payload.PullRequest.Base.SHA,
		}
	}

	rawDiff, err := p.github.GetPullRequestDiff(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		rawDiff = ""
	}

	parsedFiles, err := diff.ParseUnified(rawDiff)
	if err != nil {
		return nil, err
	}

	packet := domain.ChangePacket{
		ID: identity.New("cp"),
		Repo: domain.RepoRef{
			Owner:    repoOwner,
			Name:     repoName,
			FullName: payload.Repository.FullName,
		},
		PR: domain.PullRequestRef{
			Number:  pr.Number,
			URL:     pr.URL,
			HeadSHA: pr.HeadSHA,
			BaseSHA: pr.BaseSHA,
		},
		Author: domain.Author{
			Login: pr.Author,
			Type:  payload.PullRequest.User.Type,
		},
		Title:     pr.Title,
		Summary:   strings.TrimSpace(pr.Body),
		Source:    "github.pull_request",
		Files:     mapChangedFiles(parsedFiles),
		MergeLane: domain.MergeLaneYellow,
		Decision: domain.PolicyDecision{
			Status: domain.DecisionPending,
		},
		Deployment: domain.DeploymentRequirement{
			Strategy: domain.DeployDirect,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Metadata: map[string]string{
			"correlation_id": telemetry.CorrelationID(ctx),
		},
	}

	if err := p.bus.Publish(ctx, events.NewEnvelope(events.EventChangePacketCreated, "ingest", packet)); err != nil {
		return nil, err
	}

	packet.Mutations, err = p.mutations.Classify(ctx, packet.Files)
	if err != nil {
		return nil, err
	}
	if err := p.bus.Publish(ctx, events.NewEnvelope(events.EventMutationDetected, "mutations", packet.Mutations)); err != nil {
		return nil, err
	}

	packet.Runtime, packet.Ownership, err = p.runtime.ResolveImpact(ctx, packet)
	if err != nil {
		return nil, err
	}

	packet.RiskSummary, packet.Risks, err = p.risk.Evaluate(ctx, packet)
	if err != nil {
		return nil, err
	}
	if err := p.bus.Publish(ctx, events.NewEnvelope(events.EventRiskAssigned, "risk", packet.Risks)); err != nil {
		return nil, err
	}

	policyEval, err := p.policy.Evaluate(ctx, packet)
	if err != nil {
		return nil, err
	}
	packet.Decision = policyEval.Decision
	packet.Evidence = policyEval.Evidence
	packet.Reviewers = policyEval.Reviewers
	packet.Deployment = policyEval.Deployment

	packet.MergeLane, err = p.assigner.Assign(ctx, packet)
	if err != nil {
		return nil, err
	}
	packet.UpdatedAt = time.Now().UTC()

	if len(packet.Decision.Violations) > 0 {
		if err := p.bus.Publish(ctx, events.NewEnvelope(events.EventPolicyViolationDetected, "policy", packet.Decision.Violations)); err != nil {
			return nil, err
		}
	}
	if err := p.bus.Publish(ctx, events.NewEnvelope(events.EventMergeLaneAssigned, "lanes", map[string]any{
		"changePacketId": packet.ID,
		"lane":           packet.MergeLane,
	})); err != nil {
		return nil, err
	}

	if err := p.checks.Publish(ctx, packet); err != nil {
		return nil, err
	}

	p.logger.Info("processed pull request",
		"change_packet_id", packet.ID,
		"repo", packet.Repo.FullName,
		"pr_number", packet.PR.Number,
		"lane", string(packet.MergeLane),
		"risk_score", packet.RiskSummary.Score,
	)

	return &packet, nil
}

func mapChangedFiles(files []diff.File) []domain.ChangedFile {
	mapped := make([]domain.ChangedFile, 0, len(files))
	for _, file := range files {
		mapped = append(mapped, domain.ChangedFile{
			Path:         file.Path,
			PreviousPath: file.PreviousPath,
			Status:       domain.FileStatus(file.Status),
			Language:     languageFromPath(file.Path),
			Additions:    file.Additions,
			Deletions:    file.Deletions,
			Changes:      file.Additions + file.Deletions,
			Patch:        file.Patch,
		})
	}
	return mapped
}

func languageFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".sql":
		return "sql"
	case ".yaml", ".yml":
		return "yaml"
	case ".proto":
		return "proto"
	default:
		return "unknown"
	}
}

func (p *Processor) DescribeWebhook(payload github.PullRequestWebhookPayload) string {
	return fmt.Sprintf("%s#%d action=%s", payload.Repository.FullName, payload.PullRequest.Number, payload.Action)
}
