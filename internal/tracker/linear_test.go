package tracker

import (
	"testing"
)

func TestLinearTracker_ParseBranchName(t *testing.T) {
	tr := NewLinearTracker("https://linear.app/simonmittag/ticket")
	tests := []struct {
		branch   string
		expected string
	}{
		{"feature/FIS-123-some-task", "FIS-123"},
		{"FIS-456", "FIS-456"},
		{"fis-789_another_one", "FIS-789"},
		{"no-issue-here", ""},
		{"prefix-123-but-no-dash", "PREFIX-123"},
		{"feat/abc-123", "ABC-123"},
		{"fix/def-456", "DEF-456"},
		{"bug/ghi-789", "GHI-789"},
		{"feat-abc-123", "ABC-123"},
		{"fix-def-456", "DEF-456"},
		{"bug-ghi-789", "GHI-789"},
		{"feat/123", ""},
		{"bug/456", ""},
		{"feat-123", ""},
		{"fix-456", ""},
		{"hotfix/gh-101", "GH-101"},
		{"chore-task-123", "TASK-123"},
	}

	for _, tt := range tests {
		got := tr.parseGitBranchName(tt.branch)
		if got != tt.expected {
			t.Errorf("parseBranchName(%q) = %q; want %q", tt.branch, got, tt.expected)
		}
	}
}
