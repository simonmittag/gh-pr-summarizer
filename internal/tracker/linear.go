package tracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/simonmittag/gh-pr-summarizer/internal/prtype"
)

type LinearTracker struct {
	TicketUrlStem string
	ApiKey        string
}

func NewLinearTracker(ticketUrlStem, apiKey string) *LinearTracker {
	return &LinearTracker{
		TicketUrlStem: ticketUrlStem,
		ApiKey:        apiKey,
	}
}

func (l *LinearTracker) FetchTicket(branchName string) (*Ticket, error) {
	ticketKey := l.parseGitBranchName(branchName)
	if ticketKey == "" {
		// If parsing fails, try to infer from a valid ticket (as per requirement)
		// But first we try to fetch if we have something that looks like a key
		return nil, fmt.Errorf("unable to parse ticket key from branch name: %s", branchName)
	}

	ticket, err := l.fetchTicketFromLinear(ticketKey)
	if err != nil {
		log.Debug().Err(err).Str("ticketKey", ticketKey).Msg("unable to fetch ticket from linear, proceeding without")
		// If fetch fails, try to "load one valid ticket from linear and infer naming scheme"
		// Requirement: "if this fails, it loads one valid ticket from linear and infers the naming scheme for the branch from there."
		inferredKey, inferErr := l.inferTicketKeyFromLinear(branchName)
		if inferErr != nil {
			log.Debug().Err(inferErr).Msg("unable to infer ticket key from linear")
			return nil, fmt.Errorf("unable to determine ticket key, proceeding without ticket")
		}
		log.Debug().Str("inferredKey", inferredKey).Msg("successfully inferred ticket key from linear")
		ticket, err = l.fetchTicketFromLinear(inferredKey)
		if err != nil {
			log.Debug().Err(err).Str("inferredKey", inferredKey).Msg("unable  to fetch inferred ticket from linear")
			return nil, err
		}
	}

	log.Debug().Str("ticketKey", ticket.ID).Msg("successfully fetched from linear")
	return ticket, nil
}

func (l *LinearTracker) parseGitBranchName(branchName string) string {
	normalizedBranch := prtype.StripPrefix(branchName)

	// 2. Look for something like ABC-123 or abc-123
	re := regexp.MustCompile(`(?i)([a-z]+-\d+)`)
	match := re.FindString(normalizedBranch)
	if match != "" {
		return strings.ToUpper(match)
	}
	return ""
}

func (l *LinearTracker) fetchTicketFromLinear(ticketKey string) (*Ticket, error) {
	if l.ApiKey == "" {
		return nil, fmt.Errorf("LINEAR_API_KEY not set")
	}

	query := fmt.Sprintf(`{ "query": "{ issue(id: \"%s\") { id identifier title url } }" }`, ticketKey)
	req, err := http.NewRequest("POST", "https://api.linear.app/graphql", bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", l.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linear api returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Issue struct {
				ID         string `json:"id"`
				Identifier string `json:"identifier"`
				Title      string `json:"title"`
				URL        string `json:"url"`
			} `json:"issue"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Data.Issue.Identifier == "" {
		return nil, fmt.Errorf("ticket not found: %s", ticketKey)
	}

	return &Ticket{
		ID:    result.Data.Issue.Identifier,
		URL:   result.Data.Issue.URL,
		Title: result.Data.Issue.Title,
	}, nil
}

func (l *LinearTracker) inferTicketKeyFromLinear(branchName string) (string, error) {
	if l.ApiKey == "" {
		return "", fmt.Errorf("LINEAR_API_KEY not set")
	}

	// Fetch one recent ticket to see its identifier format
	query := `{ "query": "{ issues(first: 1) { nodes { identifier } } }" }`
	req, err := http.NewRequest("POST", "https://api.linear.app/graphql", bytes.NewBufferString(query))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", l.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Issues struct {
				Nodes []struct {
					Identifier string `json:"identifier"`
				} `json:"nodes"`
			} `json:"issues"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Data.Issues.Nodes) == 0 {
		return "", fmt.Errorf("no tickets found to infer from")
	}

	sampleIdentifier := result.Data.Issues.Nodes[0].Identifier
	// Extract prefix (e.g., "FIS" from "FIS-123")
	parts := strings.Split(sampleIdentifier, "-")
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected identifier format: %s", sampleIdentifier)
	}
	prefix := parts[0]

	// Try to find any number in the branch name and combine it with the prefix
	re := regexp.MustCompile(`\d+`)
	numberMatch := re.FindString(branchName)
	if numberMatch == "" {
		return "", fmt.Errorf("unable to find ticket number in branch name: %s", branchName)
	}

	return fmt.Sprintf("%s-%s", prefix, numberMatch), nil
}
