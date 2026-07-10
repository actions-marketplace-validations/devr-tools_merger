package policy

import (
	"strings"

	"github.com/devr-tools/merger/internal/domain"
)

func matchesRule(rule RuleConfig, packet domain.ChangePacket) bool {
	if len(rule.When.Mutations) > 0 && !packetHasAnyMutation(packet, rule.When.Mutations) {
		return false
	}
	if len(rule.When.Paths) > 0 && !packetTouchesAnyPath(packet, rule.When.Paths) {
		return false
	}
	if len(rule.When.Criticalities) > 0 && !containsCriticality(rule.When.Criticalities, packet.Runtime.Criticality) {
		return false
	}
	if rule.When.RiskScoreGTE > 0 && packet.RiskSummary.Score < rule.When.RiskScoreGTE {
		return false
	}
	if len(rule.When.OwnershipTeams) > 0 && !packetTouchesOwners(packet, rule.When.OwnershipTeams) {
		return false
	}

	return true
}

func packetHasAnyMutation(packet domain.ChangePacket, kinds []domain.MutationKind) bool {
	for _, mutation := range packet.Mutations {
		for _, kind := range kinds {
			if mutation.Kind == kind {
				return true
			}
		}
	}
	return false
}

func packetTouchesAnyPath(packet domain.ChangePacket, patterns []string) bool {
	for _, file := range packet.Files {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(file.Path), strings.ToLower(pattern)) {
				return true
			}
		}
	}
	return false
}

func containsCriticality(values []domain.Criticality, candidate domain.Criticality) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func packetTouchesOwners(packet domain.ChangePacket, teams []string) bool {
	for _, owner := range packet.Ownership {
		for _, team := range teams {
			if strings.EqualFold(owner.Team, team) {
				return true
			}
		}
	}
	return false
}
