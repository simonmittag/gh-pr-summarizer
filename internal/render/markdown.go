package render

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/simonmittag/gh-pr-summarizer/internal/tracker"
)

type Renderer struct {
	AI *openai.Client
}

func NewRenderer(ai *openai.Client) *Renderer {
	return &Renderer{
		AI: ai,
	}
}

// PRBody generates markdown for the pull request body based on the commit subjects and optional ticket.
func (r *Renderer) PRBody(subjects []string, ticket *tracker.Ticket) string {
	if len(subjects) == 0 {
		return "no local commits detected"
	}

	var sb strings.Builder

	sb.WriteString("# Why\n")
	if ticket != nil {
		sb.WriteString(fmt.Sprintf("%s, see [%s](%s)\n\n", ticket.Title, ticket.ID, ticket.URL))
	} else {
		sb.WriteString("Why this PR? See, [issue-management-ticket-placeholder](https://example.com/issue/1)\n\n")
	}

	sb.WriteString("# What\n")
	aiSummary := ""
	if r.AI != nil && ticket != nil {
		aiSummary = r.generateAISummary(subjects, ticket)
	}

	if aiSummary != "" {
		sb.WriteString(aiSummary + "\n\n")
	} else {
		sb.WriteString("A summary of what this PR changes\n\n")
	}

	sb.WriteString("# How\n")
	for _, subject := range subjects {
		sb.WriteString(fmt.Sprintf("- [x] %s\n", subject))
	}

	return sb.String()
}

func (r *Renderer) generateAISummary(subjects []string, ticket *tracker.Ticket) string {
	ticketJSON := fmt.Sprintf(`{"id": "%s", "title": "%s", "url": "%s"}`, ticket.ID, ticket.Title, ticket.URL)
	commitsStr := strings.Join(subjects, "\n")

	prompt := fmt.Sprintf(`Generate a short paragraph in Markdown format for the "# What" section of a Pull Request.
Focus on the issue first as WHAT the PR is trying to achieve and the commits after that WHAT has actually been done.
Weight your response based on where there is more data. If the issue is long and commit comments short, 
focus on the issue content. If the issue is basic but the commit comments are long, use the available data from there.

Issue Data:
%s

Commit Subjects:
%s`, ticketJSON, commitsStr)

	resp, err := r.AI.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return ""
	}

	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content)
	}

	return ""
}
