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
		title := ticket.Title
		if r.AI != nil {
			if fixedTitle := r.fixTitle(ticket.Title); fixedTitle != "" {
				title = fixedTitle
			}
		}
		sb.WriteString(fmt.Sprintf("%s, see [%s](%s)\n\n", title, ticket.ID, ticket.URL))
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
	commitsStr := strings.Join(subjects, "\n")

	prompt := fmt.Sprintf(`As a software engineer, write a 2-3 sentence markdown paragraph summarizing the "WHAT" 
of this Pull Request based on the provided issue and commits. 

Strictly follow these rules:
1. Write in a factual, engineering-oriented tone.
2. Focus on the core objective and the actual implementation.
3. Use only the provided data; do not add filler or boilerplate.
4. DO NOT mention the PR "aims to", "seeks to", or "is intended to".
5. DO NOT use third-person self-references like "This Pull Request", "This PR", or "This change".
6. DO NOT include any issue IDs (e.g., SIM-8), ticket numbers, or URLs.
7. DO NOT include headers, titles, or introductory phrases.
8. Start the paragraph directly with the content.

Issue: %s
Commits: %s`, ticket.Title, commitsStr)

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

func (r *Renderer) fixTitle(title string) string {
	prompt := fmt.Sprintf(`As a grammar editor at large, fix the capitalization, spelling, and grammar of the following title. 
Focus strictly on making the spelling and English of the title correct. 
Do not add any other text, explanation, or prefixes. 
You're only an editor, you have no specific knowledge of the title. 
Return only the edited title content itself. Never prefix the result with "Title:".

%s`, title)

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
