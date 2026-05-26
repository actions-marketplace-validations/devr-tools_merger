package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service   ServiceConfig   `yaml:"service"`
	Logging   LoggingConfig   `yaml:"logging"`
	GitHub    GitHubConfig    `yaml:"github"`
	Events    EventsConfig    `yaml:"events"`
	Policy    PolicyConfig    `yaml:"policy"`
	Lanes     LanesConfig     `yaml:"lanes"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
}

type ServiceConfig struct {
	IngestAddress       string `yaml:"ingest_address"`
	ControlPlaneAddress string `yaml:"controlplane_address"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type GitHubConfig struct {
	WebhookSecret  string `yaml:"webhook_secret"`
	AppID          string `yaml:"app_id"`
	InstallationID int64  `yaml:"installation_id"`
	PrivateKeyPath string `yaml:"private_key_path"`
}

type EventsConfig struct {
	Backend       string `yaml:"backend"`
	SubjectPrefix string `yaml:"subject_prefix"`
}

type PolicyConfig struct {
	Path string `yaml:"path"`
}

type LanesConfig struct {
	GreenMax  int `yaml:"green_max"`
	YellowMax int `yaml:"yellow_max"`
	RedMax    int `yaml:"red_max"`
}

type TelemetryConfig struct {
	ServiceName string `yaml:"service_name"`
	Environment string `yaml:"environment"`
}

func Defaults() Config {
	return Config{
		Service: ServiceConfig{
			IngestAddress:       ":8080",
			ControlPlaneAddress: ":8081",
		},
		Logging: LoggingConfig{Level: "info"},
		Events: EventsConfig{
			Backend:       "memory",
			SubjectPrefix: "merger",
		},
		Policy: PolicyConfig{
			Path: "config/policies/default.yaml",
		},
		Lanes: LanesConfig{
			GreenMax:  20,
			YellowMax: 55,
			RedMax:    85,
		},
		Telemetry: TelemetryConfig{
			ServiceName: "merger",
			Environment: "dev",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Defaults()

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
