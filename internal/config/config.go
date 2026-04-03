package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

type TrackerConfig struct {
	TicketUrlStem string `toml:"ticket_url_stem"`
	TokenEnv      string `toml:"token_env"`
}

type Config struct {
	Tracker string        `toml:"tracker"`
	Linear  TrackerConfig `toml:"linear"`
	GitHub  TrackerConfig `toml:"github"`
	Jira    TrackerConfig `toml:"jira"`
}

func EnsureGlobalConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := fmt.Sprintf("%s/.config/gh-pr-summarizer", home)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Debug().Str("path", configDir).Msg("initializing global config directory")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	configFile := fmt.Sprintf("%s/config.toml", configDir)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create empty config file: %w", err)
		}
	}

	return nil
}

func mergeTrackerConfig(global, local TrackerConfig) TrackerConfig {
	result := global
	if local.TicketUrlStem != "" {
		result.TicketUrlStem = local.TicketUrlStem
	}
	if local.TokenEnv != "" {
		result.TokenEnv = local.TokenEnv
	}
	return result
}

func mergeConfig(global, local Config) Config {
	result := global
	if local.Tracker != "" {
		result.Tracker = local.Tracker
	}
	result.Linear = mergeTrackerConfig(global.Linear, local.Linear)
	result.GitHub = mergeTrackerConfig(global.GitHub, local.GitHub)
	result.Jira = mergeTrackerConfig(global.Jira, local.Jira)
	return result
}

func applyTrackerDefaults(cfg *Config) {
	if cfg.Linear.TokenEnv == "" {
		cfg.Linear.TokenEnv = "LINEAR_API_KEY"
	}
	if cfg.GitHub.TokenEnv == "" {
		cfg.GitHub.TokenEnv = "GITHUB_TOKEN"
	}
	if cfg.Jira.TokenEnv == "" {
		cfg.Jira.TokenEnv = "ATLASSIAN_TOKEN"
	}
}

func LoadConfig() (*Config, error) {
	if err := EnsureGlobalConfig(); err != nil {
		return nil, err
	}

	home, _ := os.UserHomeDir()
	globalPath := fmt.Sprintf("%s/.config/gh-pr-summarizer/config.toml", home)
	localPath := ".ghpr.toml"

	var globalCfg Config
	if _, err := os.Stat(globalPath); err == nil {
		if _, err := toml.DecodeFile(globalPath, &globalCfg); err != nil {
			return nil, fmt.Errorf("failed to decode global config: %w", err)
		}
	}

	var localCfg Config
	if _, err := os.Stat(localPath); err == nil {
		if _, err := toml.DecodeFile(localPath, &localCfg); err != nil {
			return nil, fmt.Errorf("failed to decode local config: %w", err)
		}
	}

	merged := mergeConfig(globalCfg, localCfg)
	applyTrackerDefaults(&merged)

	if merged.Tracker == "" {
		merged.Tracker = "none"
	}

	return &merged, nil
}
