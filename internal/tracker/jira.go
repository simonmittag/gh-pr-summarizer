package tracker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

type JiraTracker struct {
	TicketUrlStem string
	Token         string
}

func NewJiraTracker(ticketUrlStem, token string) *JiraTracker {
	return &JiraTracker{
		TicketUrlStem: ticketUrlStem,
		Token:         token,
	}
}

func (j *JiraTracker) FetchTicket(branchName string) (*Ticket, error) {
	ticketKey := j.parseGitBranchName(branchName)
	if ticketKey == "" {
		return nil, fmt.Errorf("unable to parse Jira ticket key from branch name: %s", branchName)
	}

	ticket, err := j.fetchTicketFromJira(ticketKey)
	if err != nil {
		log.Debug().Err(err).Str("ticketKey", ticketKey).Msg("failed to fetch ticket from jira, proceeding without ticket")
		return nil, err
	}
	log.Debug().Str("ticketKey", ticketKey).Msg("successfully fetched ticket from from jira")
	return ticket, nil
}

func (j *JiraTracker) parseGitBranchName(branchName string) string {
	// Common prefixes to remove
	prefixes := []string{"feat/", "fix/", "bug/", "feature/", "hotfix/", "chore/", "feat-", "fix-", "bug-", "feature-", "hotfix-", "chore-"}
	normalizedBranch := branchName
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(normalizedBranch), p) {
			normalizedBranch = normalizedBranch[len(p):]
			break
		}
	}

	// Look for something like ABC-123 or abc-123
	re := regexp.MustCompile(`(?i)([a-z]+-\d+)`)
	match := re.FindString(normalizedBranch)
	if match != "" {
		return strings.ToUpper(match)
	}
	return ""
}

func (j *JiraTracker) fetchTicketFromJira(ticketKey string) (*Ticket, error) {
	if j.Token == "" {
		return nil, fmt.Errorf("ATLASSIAN_TOKEN not set")
	}

	apiHost := ""
	if j.TicketUrlStem != "" {
		re := regexp.MustCompile(`https?://[^/]+`)
		apiHost = re.FindString(j.TicketUrlStem)
	}

	if apiHost == "" {
		return nil, fmt.Errorf("unable to infer Jira API host from ticket_url_stem: %s. please set it to something like https://your-domain.atlassian.net/browse/", j.TicketUrlStem)
	}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s", apiHost, ticketKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Logic for authentication:
	// 1. If token contains ":", assume it's "email:api_token" and encode to Basic
	// 2. If ATLASSIAN_EMAIL is set, assume token is API token and encode to Basic
	// 3. Otherwise, try Bearer (PAT) and fallback to Basic (already encoded)

	email := os.Getenv("ATLASSIAN_EMAIL")
	if strings.Contains(j.Token, ":") {
		encoded := base64.StdEncoding.EncodeToString([]byte(j.Token))
		req.Header.Set("Authorization", "Basic "+encoded)
	} else if email != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(email + ":" + j.Token))
		req.Header.Set("Authorization", "Basic "+encoded)
	} else {
		// Fallback to original logic: try Bearer then Basic
		req.Header.Set("Authorization", "Bearer "+j.Token)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If initial attempt failed (e.g. 401/403), try Basic auth with raw token as fallback
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		if !strings.Contains(j.Token, ":") && email == "" {
			req.Header.Set("Authorization", "Basic "+j.Token)
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API returned status %d for %s", resp.StatusCode, url)
	}

	var result struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
		} `json:"fields"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	ticketUrl := ""
	if j.TicketUrlStem != "" {
		ticketUrl = j.TicketUrlStem + result.Key
	}

	return &Ticket{
		ID:    result.Key,
		URL:   ticketUrl,
		Title: result.Fields.Summary,
	}, nil
}
