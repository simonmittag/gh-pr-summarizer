package prtype

import "testing"

func TestPrType_Helpers(t *testing.T) {
	t.Run("DetectTypeFromBranch", func(t *testing.T) {
		tests := []struct {
			branch   string
			expected string
		}{
			{"feat/sim-1", "feat"},
			{"FEAT/sim-1", "feat"},
			{"feature/sim-1", "feature"},
			{"fix/fis-1", "fix"},
			{"bugfix/fis-1", "bugfix"},
			{"terraform/tf-1", "terraform"},
			{"tf/tf-1", "tf"},
			{"infra/tf-1", "infra"},
			{"polish/sim-9", "polish"},
			{"docs/sim-9", "docs"},
			{"doc/sim-9", "doc"},
			{"hotfix/1", "hotfix"},
			{"chore/1", "chore"},
			{"unknown/1", ""},
			{"sim-1", ""},
		}
		for _, tt := range tests {
			if got := DetectTypeFromBranch(tt.branch); got != tt.expected {
				t.Errorf("DetectTypeFromBranch(%q) = %q, want %q", tt.branch, got, tt.expected)
			}
		}
	})

	t.Run("InferTypeFromTitle", func(t *testing.T) {
		tests := []struct {
			title    string
			expected string
		}{
			{"Fix auth fallback", "fix"},
			{"Add feature X", "feature"}, // matches 'feature' because it's longer than 'feat'
			{"Add feat X", "feat"},
			{"Update docs", "docs"},
			{"Refactor infra", "infra"},
			{"Bugfix for Y", "bugfix"},
			{"No keyword", ""},
		}
		for _, tt := range tests {
			if got := InferTypeFromTitle(tt.title); got != tt.expected {
				t.Errorf("InferTypeFromTitle(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		}
	})

	t.Run("StripPrefix", func(t *testing.T) {
		tests := []struct {
			branch   string
			expected string
		}{
			{"feat/sim-1", "sim-1"},
			{"feature/sim-1", "sim-1"},
			{"fix/fis-1", "fis-1"},
			{"bugfix/fis-1", "fis-1"},
			{"terraform/tf-1", "tf-1"},
			{"tf/tf-1", "tf-1"},
			{"infra/tf-1", "tf-1"},
			{"polish/sim-9", "sim-9"},
			{"docs/sim-9", "sim-9"},
			{"doc/sim-9", "sim-9"},
			{"hotfix/1", "1"},
			{"chore/1", "1"},
			{"unknown/1", "unknown/1"},
			{"sim-1", "sim-1"},
			{"FEAT/SIM-1", "SIM-1"},
		}
		for _, tt := range tests {
			if got := StripPrefix(tt.branch); got != tt.expected {
				t.Errorf("StripPrefix(%q) = %q, want %q", tt.branch, got, tt.expected)
			}
		}
	})
}
