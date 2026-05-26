package mutations

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/identity"
)

type SignalExtractor interface {
	Name() string
	Supports(domain.ChangedFile) bool
	Extract(context.Context, domain.ChangedFile) ([]domain.MutationSignal, error)
}

type ContentLoader interface {
	Load(context.Context, string) ([]byte, error)
}

type Analyzer interface {
	Name() string
	Supports(domain.ChangedFile) bool
	Analyze(context.Context, AnalysisInput) ([]domain.Mutation, error)
}

type AnalysisRequest struct {
	Repo    domain.RepoRef
	Ref     string
	Files   []domain.ChangedFile
	Content ContentLoader
}

type AnalysisInput struct {
	Repo    domain.RepoRef
	Ref     string
	File    domain.ChangedFile
	Content []byte
}

type Engine interface {
	Classify(context.Context, AnalysisRequest) ([]domain.Mutation, error)
}

type RuleBasedEngine struct {
	rules      []Rule
	extractors []SignalExtractor
	analyzers  []Analyzer
}

func NewRuleBasedEngine(rules []Rule, extractors []SignalExtractor, analyzers []Analyzer) *RuleBasedEngine {
	return &RuleBasedEngine{rules: rules, extractors: extractors, analyzers: analyzers}
}

func DefaultEngine() *RuleBasedEngine {
	return NewRuleBasedEngine(DefaultRules(), DefaultExtractors(), DefaultAnalyzers())
}

func (e *RuleBasedEngine) Classify(ctx context.Context, request AnalysisRequest) ([]domain.Mutation, error) {
	index := make(map[domain.MutationKind]*domain.Mutation)

	for _, file := range request.Files {
		signals, err := e.extractSignals(ctx, file)
		if err != nil {
			return nil, err
		}

		for _, rule := range e.rules {
			if !rule.Matches(file.Path, signals) {
				continue
			}

			addMutation(index, domain.Mutation{
				ID:          identity.New("mut"),
				Kind:        rule.Kind,
				Severity:    rule.Severity,
				Confidence:  rule.Confidence,
				Title:       rule.Title,
				Description: rule.Description,
				Files:       []string{file.Path},
				Signals:     signals,
				Detector:    rule.Name,
			})
		}

		content, _ := loadContent(ctx, request.Content, file.Path)
		for _, analyzer := range e.analyzers {
			if !analyzer.Supports(file) {
				continue
			}

			mutations, err := analyzer.Analyze(ctx, AnalysisInput{
				Repo:    request.Repo,
				Ref:     request.Ref,
				File:    file,
				Content: content,
			})
			if err != nil {
				return nil, err
			}

			for _, mutation := range mutations {
				addMutation(index, mutation)
			}
		}
	}

	if len(index) == 0 && len(request.Files) > 0 {
		fallback := domain.Mutation{
			ID:         identity.New("mut"),
			Kind:       domain.MutationUnknown,
			Severity:   domain.SeverityLow,
			Confidence: 0.35,
			Title:      "unclassified change surface",
			Detector:   "fallback",
		}
		for _, file := range request.Files {
			fallback.Files = append(fallback.Files, file.Path)
		}
		addMutation(index, fallback)
	}

	mutations := make([]domain.Mutation, 0, len(index))
	for _, mutation := range index {
		mutations = append(mutations, *mutation)
	}

	return mutations, nil
}

func (e *RuleBasedEngine) extractSignals(ctx context.Context, file domain.ChangedFile) ([]domain.MutationSignal, error) {
	signals := make([]domain.MutationSignal, 0)
	signals = append(signals, domain.MutationSignal{
		Source: "path",
		Value:  strings.ToLower(filepath.Base(file.Path)),
		Weight: 1,
	})

	for _, extractor := range e.extractors {
		if !extractor.Supports(file) {
			continue
		}

		extracted, err := extractor.Extract(ctx, file)
		if err != nil {
			return nil, err
		}
		signals = append(signals, extracted...)
	}

	return signals, nil
}

func loadContent(ctx context.Context, loader ContentLoader, path string) ([]byte, error) {
	if loader == nil {
		return nil, nil
	}
	return loader.Load(ctx, path)
}

func addMutation(index map[domain.MutationKind]*domain.Mutation, candidate domain.Mutation) {
	current := index[candidate.Kind]
	if current == nil {
		copyCandidate := candidate
		if copyCandidate.ID == "" {
			copyCandidate.ID = identity.New("mut")
		}
		index[candidate.Kind] = &copyCandidate
		return
	}

	current.Files = appendUnique(current.Files, candidate.Files...)
	current.Signals = appendUniqueSignals(current.Signals, candidate.Signals)
	if candidate.Confidence > current.Confidence {
		current.Confidence = candidate.Confidence
	}
	if severityRank(candidate.Severity) > severityRank(current.Severity) {
		current.Severity = candidate.Severity
	}
	if current.Title == "" {
		current.Title = candidate.Title
	}
	if current.Description == "" {
		current.Description = candidate.Description
	}
	if current.Detector == "" {
		current.Detector = candidate.Detector
	}
}

func appendUnique(values []string, candidates ...string) []string {
	for _, candidate := range candidates {
		found := false
		for _, value := range values {
			if value == candidate {
				found = true
				break
			}
		}
		if !found {
			values = append(values, candidate)
		}
	}
	return values
}

func appendUniqueSignals(values []domain.MutationSignal, candidates []domain.MutationSignal) []domain.MutationSignal {
	for _, candidate := range candidates {
		found := false
		for _, value := range values {
			if value.Source == candidate.Source && value.Value == candidate.Value {
				found = true
				break
			}
		}
		if !found {
			values = append(values, candidate)
		}
	}
	return values
}

func severityRank(value domain.Severity) int {
	switch value {
	case domain.SeverityCritical:
		return 4
	case domain.SeverityHigh:
		return 3
	case domain.SeverityMedium:
		return 2
	default:
		return 1
	}
}
