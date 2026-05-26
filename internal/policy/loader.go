package policy

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
