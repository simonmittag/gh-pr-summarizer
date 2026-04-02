package render

import (
	"fmt"
	"strings"

	"github.com/simonmittag/gh-pr-summarizer/internal/tracker"
)

// PRBody generates markdown for the pull request body based on the commit subjects and optional ticket.
func PRBody(subjects []string, ticket *tracker.Ticket) string {
	if len(subjects) == 0 {
		return "no local commits detected"
	}

	var sb strings.Builder

	sb.WriteString("# Why\n")
	if ticket != nil {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", ticket.ID, ticket.URL))
		sb.WriteString(fmt.Sprintf("Rationale: %s\n\n", ticket.Title))
	} else {
		sb.WriteString("[issue-management-ticket-placeholder](https://example.com/issue/1)\n")
		sb.WriteString("Rationale: Why this PR? \n\n")
	}

	sb.WriteString("# What\n")
	sb.WriteString("A summary of what this PR changes\n\n")

	sb.WriteString("# How\n")
	if len(subjects) == 0 {
		sb.WriteString("Itemise how this PR achieves the above\n\n")
	}
	for _, subject := range subjects {
		sb.WriteString(fmt.Sprintf("- [x] %s\n", subject))
	}

	return sb.String()
}
