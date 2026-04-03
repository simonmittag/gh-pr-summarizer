package render

import (
	"strings"
	"testing"

	"github.com/simonmittag/gh-pr-summarizer/internal/tracker"
)

func TestRenderer_PRBody_NoAI(t *testing.T) {
	renderer := NewRenderer(nil)
	subjects := []string{"feat: add login", "fix: background color"}
	ticket := &tracker.Ticket{
		ID:    "KAN-1",
		Title: "Implementation of Login",
		URL:   "https://jira.example.com/browse/KAN-1",
	}

	markdown := renderer.PRBody(subjects, ticket)

	if !strings.Contains(markdown, "# Why") {
		t.Errorf("expected markdown to contain # Why")
	}
	if !strings.Contains(markdown, "Implementation of Login, see [KAN-1](https://jira.example.com/browse/KAN-1)") {
		t.Errorf("expected markdown to contain ticket info")
	}
	if !strings.Contains(markdown, "# What") {
		t.Errorf("expected markdown to contain # What")
	}
	if !strings.Contains(markdown, "A summary of what this PR changes") {
		t.Errorf("expected markdown to contain default summary")
	}
	if !strings.Contains(markdown, "- [x] feat: add login") {
		t.Errorf("expected markdown to contain commit subject 1")
	}
	if !strings.Contains(markdown, "- [x] fix: background color") {
		t.Errorf("expected markdown to contain commit subject 2")
	}
}

func TestRenderer_PRBody_NoTicket(t *testing.T) {
	renderer := NewRenderer(nil)
	subjects := []string{"chore: update deps"}

	markdown := renderer.PRBody(subjects, nil)

	if !strings.Contains(markdown, "Why this PR? See, [issue-management-ticket-placeholder]") {
		t.Errorf("expected markdown to contain placeholder for why")
	}
	if !strings.Contains(markdown, "A summary of what this PR changes") {
		t.Errorf("expected markdown to contain default summary")
	}
}
