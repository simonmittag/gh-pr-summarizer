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

	markdown := renderer.PRBody(subjects, ticket, "feat/KAN-1")

	if !strings.Contains(markdown, "# Title") {
		t.Errorf("expected markdown to contain # Title")
	}
	if !strings.Contains(markdown, "✨ FEAT/KAN-1: Implementation of Login") {
		t.Errorf("expected title to be correct")
	}
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

	markdown := renderer.PRBody(subjects, nil, "chore/update-deps")

	if !strings.Contains(markdown, "# Title") {
		t.Errorf("expected markdown to contain # Title")
	}
	if !strings.Contains(markdown, "🧹 CHORE/update-deps: Your title goes here") {
		t.Errorf("expected title to be correct for no ticket, got %q", markdown)
	}
	if !strings.Contains(markdown, "Why this PR? See, [ticket-management-placeholder]") {
		t.Errorf("expected markdown to contain placeholder for why")
	}
	if !strings.Contains(markdown, "A summary of what this PR changes") {
		t.Errorf("expected markdown to contain default summary")
	}
}

func TestRenderer_GenerateTitleSection(t *testing.T) {
	tests := []struct {
		name     string
		draft    bool
		branch   string
		ticket   *tracker.Ticket
		expected string
	}{
		{
			name:     "draft feat branch",
			draft:    true,
			branch:   "feat/sim-1",
			ticket:   &tracker.Ticket{ID: "SIM-1", Title: "Improve auth resolution"},
			expected: "# Title\n🚧 ✨ FEAT/SIM-1: Improve auth resolution\n",
		},
		{
			name:     "non-draft fix branch",
			draft:    false,
			branch:   "fix/fis-3163",
			ticket:   &tracker.Ticket{ID: "FIS-3163", Title: "Handle token fallback"},
			expected: "# Title\n🐛 FIX/FIS-3163: Handle token fallback\n",
		},
		{
			name:     "feature prefix alias",
			draft:    false,
			branch:   "feature/sim-1",
			ticket:   &tracker.Ticket{ID: "SIM-1", Title: "Improve auth resolution"},
			expected: "# Title\n✨ FEATURE/SIM-1: Improve auth resolution\n",
		},
		{
			name:     "terraform prefix",
			draft:    false,
			branch:   "terraform/fis-22",
			ticket:   &tracker.Ticket{ID: "FIS-22", Title: "Update infra"},
			expected: "# Title\n🚜 TERRAFORM/FIS-22: Update infra\n",
		},
		{
			name:     "tf prefix",
			draft:    false,
			branch:   "tf/fis-22",
			ticket:   &tracker.Ticket{ID: "FIS-22", Title: "Update infra"},
			expected: "# Title\n🚜 TF/FIS-22: Update infra\n",
		},
		{
			name:     "infra prefix",
			draft:    false,
			branch:   "infra/fis-22",
			ticket:   &tracker.Ticket{ID: "FIS-22", Title: "Update infra"},
			expected: "# Title\n🚜 INFRA/FIS-22: Update infra\n",
		},
		{
			name:     "bug prefix",
			draft:    false,
			branch:   "bug/fis-1",
			ticket:   &tracker.Ticket{ID: "FIS-1", Title: "Fix bug"},
			expected: "# Title\n🐛 BUG/FIS-1: Fix bug\n",
		},
		{
			name:     "bugfix prefix",
			draft:    false,
			branch:   "bugfix/fis-1",
			ticket:   &tracker.Ticket{ID: "FIS-1", Title: "Fix bug"},
			expected: "# Title\n🐛 BUGFIX/FIS-1: Fix bug\n",
		},
		{
			name:     "polish prefix",
			draft:    false,
			branch:   "polish/sim-9",
			ticket:   &tracker.Ticket{ID: "SIM-9", Title: "Clean up code"},
			expected: "# Title\n💄 POLISH/SIM-9: Clean up code\n",
		},
		{
			name:     "docs prefix",
			draft:    false,
			branch:   "docs/sim-9",
			ticket:   &tracker.Ticket{ID: "SIM-9", Title: "Update readme"},
			expected: "# Title\n💄 DOCS/SIM-9: Update readme\n",
		},
		{
			name:     "doc prefix",
			draft:    false,
			branch:   "doc/sim-9",
			ticket:   &tracker.Ticket{ID: "SIM-9", Title: "Update readme"},
			expected: "# Title\n💄 DOC/SIM-9: Update readme\n",
		},
		{
			name:     "tracker-confirmed issue key preserved exactly",
			draft:    false,
			branch:   "feat/sim-1",
			ticket:   &tracker.Ticket{ID: "sim-1", Title: "Mixed Case Key"},
			expected: "# Title\n✨ FEAT/sim-1: Mixed Case Key\n",
		},
		{
			name:     "GitHub numeric issue key",
			draft:    false,
			branch:   "feat/1",
			ticket:   &tracker.Ticket{ID: "1", Title: "GitHub issue"},
			expected: "# Title\n✨ FEAT/1: GitHub issue\n",
		},
		{
			name:     "no-prefix branch infers both type and emoji from issue title keyword",
			draft:    false,
			branch:   "sim-1",
			ticket:   &tracker.Ticket{ID: "SIM-1", Title: "Fix auth fallback"},
			expected: "# Title\n🐛 FIX/SIM-1: Fix auth fallback\n",
		},
		{
			name:     "branch prefix overrides issue title keyword",
			draft:    false,
			branch:   "feat/sim-1",
			ticket:   &tracker.Ticket{ID: "SIM-1", Title: "Fix auth fallback"},
			expected: "# Title\n✨ FEAT/SIM-1: Fix auth fallback\n",
		},
		{
			name:     "fallback to feat only when both fail",
			draft:    false,
			branch:   "sim-1",
			ticket:   &tracker.Ticket{ID: "SIM-1", Title: "Unknown task"},
			expected: "# Title\n✨ FEAT/SIM-1: Unknown task\n",
		},
		{
			name:     "missing issue title fallback",
			draft:    false,
			branch:   "feat/sim-1",
			ticket:   &tracker.Ticket{ID: "SIM-1", Title: ""},
			expected: "# Title\n✨ FEAT/SIM-1: Your title goes here\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(nil)
			r.Draft = tt.draft
			got := r.generateTitleSection(tt.ticket, tt.branch)
			if got != tt.expected {
				t.Errorf("generateTitleSection() got = %q, want %q", got, tt.expected)
			}
		})
	}
}
