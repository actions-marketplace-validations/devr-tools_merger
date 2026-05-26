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

type Engine interface {
	Classify(context.Context, []domain.ChangedFile) ([]domain.Mutation, error)
}

type RuleBasedEngine struct {
	rules      []Rule
	extractors []SignalExtractor
}

func NewRuleBasedEngine(rules []Rule, extractors []SignalExtractor) *RuleBasedEngine {
	return &RuleBasedEngine{rules: rules, extractors: extractors}
}

func DefaultEngine() *RuleBasedEngine {
	return NewRuleBasedEngine(DefaultRules(), DefaultExtractors())
}

func (e *RuleBasedEngine) Classify(ctx context.Context, files []domain.ChangedFile) ([]domain.Mutation, error) {
	index := make(map[domain.MutationKind]*domain.Mutation)

	for _, file := range files {
		signals, err := e.extractSignals(ctx, file)
		if err != nil {
			return nil, err
		}

		for _, rule := range e.rules {
			if !rule.Matches(file.Path, signals) {
				continue
			}

			mutation := index[rule.Kind]
			if mutation == nil {
				mutation = &domain.Mutation{
					ID:          identity.New("mut"),
					Kind:        rule.Kind,
					Severity:    rule.Severity,
					Confidence:  rule.Confidence,
					Title:       rule.Title,
					Description: rule.Description,
					Detector:    rule.Name,
				}
				index[rule.Kind] = mutation
			}

			mutation.Files = appendUnique(mutation.Files, file.Path)
			mutation.Signals = appendUniqueSignals(mutation.Signals, signals)
		}
	}

	if len(index) == 0 && len(files) > 0 {
		index[domain.MutationUnknown] = &domain.Mutation{
			ID:         identity.New("mut"),
			Kind:       domain.MutationUnknown,
			Severity:   domain.SeverityLow,
			Confidence: 0.35,
			Title:      "unclassified change surface",
			Detector:   "fallback",
		}
		for _, file := range files {
			index[domain.MutationUnknown].Files = append(index[domain.MutationUnknown].Files, file.Path)
		}
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

func appendUnique(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
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
