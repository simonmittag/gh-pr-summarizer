package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeConfig(t *testing.T) {
	global := Config{
		Tracker: "linear",
		Linear: TrackerConfig{
			TokenEnv:      "GLOBAL_LINEAR_TOKEN",
			TicketUrlStem: "https://linear.app/global/issue",
		},
		GitHub: TrackerConfig{
			TokenEnv: "GLOBAL_GITHUB_TOKEN",
		},
	}

	local := Config{
		Linear: TrackerConfig{
			TicketUrlStem: "https://linear.app/local/issue",
		},
		Jira: TrackerConfig{
			TokenEnv: "LOCAL_JIRA_TOKEN",
		},
	}

	merged := mergeConfig(global, local)

	// Scalar field from global (because local was empty)
	if merged.Tracker != "linear" {
		t.Errorf("expected tracker 'linear', got %s", merged.Tracker)
	}

	// Nested field: override global with local
	if merged.Linear.TicketUrlStem != "https://linear.app/local/issue" {
		t.Errorf("expected Linear.TicketUrlStem 'https://linear.app/local/issue', got %s", merged.Linear.TicketUrlStem)
	}

	// Nested field: retain global when local is missing
	if merged.Linear.TokenEnv != "GLOBAL_LINEAR_TOKEN" {
		t.Errorf("expected Linear.TokenEnv 'GLOBAL_LINEAR_TOKEN', got %s", merged.Linear.TokenEnv)
	}

	// Root-level field override if local provided it
	localWithTracker := local
	localWithTracker.Tracker = "jira"
	mergedWithTracker := mergeConfig(global, localWithTracker)
	if mergedWithTracker.Tracker != "jira" {
		t.Errorf("expected tracker 'jira', got %s", mergedWithTracker.Tracker)
	}

	// Entirely new tracker from local
	if merged.Jira.TokenEnv != "LOCAL_JIRA_TOKEN" {
		t.Errorf("expected Jira.TokenEnv 'LOCAL_JIRA_TOKEN', got %s", merged.Jira.TokenEnv)
	}
}

func TestApplyTrackerDefaults(t *testing.T) {
	cfg := &Config{}
	applyTrackerDefaults(cfg)

	if cfg.Linear.TokenEnv != "LINEAR_API_KEY" {
		t.Errorf("expected Linear default, got %s", cfg.Linear.TokenEnv)
	}
	if cfg.GitHub.TokenEnv != "GITHUB_TOKEN" {
		t.Errorf("expected GitHub default, got %s", cfg.GitHub.TokenEnv)
	}
	if cfg.Jira.TokenEnv != "ATLASSIAN_TOKEN" {
		t.Errorf("expected Jira default, got %s", cfg.Jira.TokenEnv)
	}

	// Ensure it doesn't override existing values
	cfg2 := &Config{
		Linear: TrackerConfig{TokenEnv: "CUSTOM_LINEAR"},
	}
	applyTrackerDefaults(cfg2)
	if cfg2.Linear.TokenEnv != "CUSTOM_LINEAR" {
		t.Errorf("expected custom Linear token, got %s", cfg2.Linear.TokenEnv)
	}
}

func TestEnsureGlobalConfig(t *testing.T) {
	// Mock HOME for testing
	tempHome, err := os.MkdirTemp("", "config-test-home")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	err = EnsureGlobalConfig()
	if err != nil {
		t.Fatalf("EnsureGlobalConfig failed: %v", err)
	}

	configDir := filepath.Join(tempHome, ".config", "gh-pr-summarizer")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}

	configFile := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Test idempotency: call again, it should not fail or change content
	err = os.WriteFile(configFile, []byte("tracker = 'linear'"), 0644)
	if err != nil {
		t.Fatalf("failed to write to config file: %v", err)
	}

	err = EnsureGlobalConfig()
	if err != nil {
		t.Fatalf("EnsureGlobalConfig failed on second call: %v", err)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if string(content) != "tracker = 'linear'" {
		t.Error("config file was overwritten")
	}
}
