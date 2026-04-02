package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type JiraTracker struct {
	IssueUrlStem string
	Token        string
}

func NewJiraTracker(issueUrlStem string) *JiraTracker {
	return &JiraTracker{
		IssueUrlStem: issueUrlStem,
		Token:        os.Getenv("ATLASSIAN_TOKEN"),
	}
}

func (j *JiraTracker) FetchIssue(branchName string) (*Ticket, error) {
	issueKey := j.parseBranchName(branchName)
	if issueKey == "" {
		return nil, fmt.Errorf("could not parse Jira issue key from branch name: %s", branchName)
	}

	return j.fetchFromJira(issueKey)
}

func (j *JiraTracker) parseBranchName(branchName string) string {
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

func (j *JiraTracker) fetchFromJira(issueKey string) (*Ticket, error) {
	if j.Token == "" {
		return nil, fmt.Errorf("ATLASSIAN_TOKEN not set")
	}

	// For Jira Cloud, the API base is usually inferred from the issue URL stem.
	// But let's assume the user provides the full URL for the issue in the stem.
	// We'll try to find the host from the URL stem.
	apiHost := ""
	if j.IssueUrlStem != "" {
		// Example: https://mycompany.atlassian.net/browse/
		re := regexp.MustCompile(`https?://[^/]+`)
		apiHost = re.FindString(j.IssueUrlStem)
	}

	if apiHost == "" {
		return nil, fmt.Errorf("could not infer Jira API host from issue_url_stem: %s", j.IssueUrlStem)
	}

	url := fmt.Sprintf("%s/rest/api/2/issue/%s", apiHost, issueKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Assuming ATLASSIAN_TOKEN is used as a Bearer token or we might need email.
	// However, usually ATLASSIAN_TOKEN for personal tokens can be Bearer.
	// If it's an API Token, it should be Basic Auth with email.
	// Since no email is provided, let's try Bearer token first.
	// Note: ATLASSIAN_TOKEN in some contexts is actually the base64 encoded "email:token".
	req.Header.Set("Authorization", "Bearer "+j.Token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Try Basic Auth if Bearer fails?
		// Actually let's try it as the full token first.
		// If the user provided the base64 encoded "email:token", it should be "Basic " + token.
		req.Header.Set("Authorization", "Basic "+j.Token)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
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

	issueUrl := ""
	if j.IssueUrlStem != "" {
		issueUrl = j.IssueUrlStem + result.Key
	}

	return &Ticket{
		ID:    result.Key,
		URL:   issueUrl,
		Title: result.Fields.Summary,
	}, nil
}
