package tracker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestJiraTracker_ParseBranchName(t *testing.T) {
	tr := NewJiraTracker("https://mycompany.atlassian.net/browse/")
	tests := []struct {
		branch   string
		expected string
	}{
		{"feature/PROJ-123-some-task", "PROJ-123"},
		{"PROJ-456", "PROJ-456"},
		{"proj-789_another_one", "PROJ-789"},
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
		got := tr.parseBranchName(tt.branch)
		if got != tt.expected {
			t.Errorf("parseBranchName(%q) = %q; want %q", tt.branch, got, tt.expected)
		}
	}
}

func TestJiraTracker_FetchIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/issue/PROJ-123" {
			t.Errorf("expected path /rest/api/2/issue/PROJ-123, got %s", r.URL.Path)
		}
		if !((r.Header.Get("Authorization") == "Bearer dummy-token") || (r.Header.Get("Authorization") == "Basic dummy-token")) {
			t.Errorf("expected Authorization header with dummy-token, got %s", r.Header.Get("Authorization"))
		}

		resp := struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
			} `json:"fields"`
		}{
			Key: "PROJ-123",
			Fields: struct {
				Summary string `json:"summary"`
			}{
				Summary: "Test Jira Issue",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	os.Setenv("ATLASSIAN_TOKEN", "dummy-token")
	defer os.Unsetenv("ATLASSIAN_TOKEN")

	// We use the server URL as the stem for testing so it can infer the host
	tr := NewJiraTracker(server.URL + "/browse/")
	ticket, err := tr.FetchIssue("feature/PROJ-123")
	if err != nil {
		t.Fatalf("FetchIssue failed: %v", err)
	}

	if ticket.ID != "PROJ-123" {
		t.Errorf("expected ID PROJ-123, got %s", ticket.ID)
	}
	if ticket.Title != "Test Jira Issue" {
		t.Errorf("expected Title 'Test Jira Issue', got %s", ticket.Title)
	}
	if ticket.URL != server.URL+"/browse/PROJ-123" {
		t.Errorf("expected URL %s, got %s", server.URL+"/browse/PROJ-123", ticket.URL)
	}
}
