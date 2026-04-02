package tracker

import (
	"testing"
)

func TestGitHubTracker_ParseBranchName(t *testing.T) {
	tr := NewGitHubTracker("owner", "repo")
	tests := []struct {
		branch   string
		expected string
	}{
		{"feature/123-some-task", "123"},
		{"123", "123"},
		{"fix/456", "456"},
		{"bug/789-issue", "789"},
		{"feature-101", "101"},
		{"hotfix-202", "202"},
		{"chore/303", "303"},
		{"no-issue-here", ""},
		{"prefix-123-but-no-dash", "123"},
		{"feat/abc-123", "123"},
		{"fix/def-456", "456"},
		{"bug/ghi-789", "789"},
		{"feat-abc-123", "123"},
		{"fix-def-456", "456"},
		{"bug-ghi-789", "789"},
		{"issue#123", "123"},
		{"GH-456", "456"},
	}

	for _, tt := range tests {
		got := tr.parseBranchName(tt.branch)
		if got != tt.expected {
			t.Errorf("parseBranchName(%q) = %q; want %q", tt.branch, got, tt.expected)
		}
	}
}
