package mutations

import (
	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/pkg/identity"
)

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

	current.Files = appendUniqueStrings(current.Files, candidate.Files...)
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

func appendUniqueStrings(values []string, candidates ...string) []string {
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
