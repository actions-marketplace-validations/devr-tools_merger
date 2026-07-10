package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mergerhq/merger/internal/config"
	"github.com/mergerhq/merger/internal/policy"
)

// configCandidates lists the config locations merger auto-discovers, in
// priority order, relative to the repository root. This mirrors the
// devr-tools convention (a root file, then a tool-named dot directory).
var configCandidates = []string{
	"merger.yaml",
	"merger.yml",
	"merger.json",
	".merger/merger.yaml",
	".merger/merger.yml",
	".merger/merger.json",
	".merger/config.yaml",
	".merger/config.yml",
	".merger/config.json",
}

var configDirCandidates = []string{
	"merger.yaml", "merger.yml", "merger.json",
	"config.yaml", "config.yml", "config.json",
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// discoverConfigPath resolves the config file to use. An explicit path is
// honored (and, if it is a directory, searched for a merger config). Otherwise
// the standard candidates under root are tried. An empty return with a nil
// error means no config was found and defaults should be used.
func discoverConfigPath(root, explicit string) (string, error) {
	if explicit != "" {
		info, err := os.Stat(explicit)
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return explicit, nil
		}
		for _, name := range configDirCandidates {
			candidate := filepath.Join(explicit, name)
			if fileExists(candidate) {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("no merger config (merger.* or config.*) found in %s", explicit)
	}

	for _, candidate := range configCandidates {
		path := filepath.Join(root, candidate)
		if fileExists(path) {
			return path, nil
		}
	}
	return "", nil
}

// loadConfig discovers and loads configuration, falling back to defaults when
// none is found. It returns the resolved path ("" when defaults were used).
func loadConfig(root, explicit string) (config.Config, string, error) {
	path, err := discoverConfigPath(root, explicit)
	if err != nil {
		return config.Config{}, "", err
	}
	if path == "" {
		return config.Defaults(), "", nil
	}
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, "", fmt.Errorf("load config %s: %w", path, err)
	}
	return cfg, path, nil
}

// resolvePolicyPath determines the policy file for a run. An explicit flag wins;
// otherwise the config's policy path is resolved relative to root. The returned
// path is empty when no policy file could be located.
func resolvePolicyPath(root, explicit string, cfg config.Config) string {
	if explicit != "" {
		return explicit
	}
	if cfg.Policy.Path == "" {
		return ""
	}
	path := cfg.Policy.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	if fileExists(path) {
		return path
	}
	return ""
}

// loadPolicy loads a policy file when one is resolvable. A missing policy is
// not an error for scanning: it yields an empty rule set and found=false.
func loadPolicy(root, explicit string, cfg config.Config) (policy.Config, string, bool, error) {
	path := resolvePolicyPath(root, explicit, cfg)
	if path == "" {
		return policy.Config{}, "", false, nil
	}
	policyConfig, err := policy.LoadConfig(path)
	if err != nil {
		return policy.Config{}, path, false, fmt.Errorf("load policy %s: %w", path, err)
	}
	return policyConfig, path, true, nil
}
