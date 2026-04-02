package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type GitHubTracker struct {
	RepoOwner string
	RepoName  string
	Token     string
}

func NewGitHubTracker(owner, repo string) *GitHubTracker {
	token := os.Getenv("GITHUB_TOKEN")
	return &GitHubTracker{
		RepoOwner: owner,
		RepoName:  repo,
		Token:     token,
	}
}

func (g *GitHubTracker) FetchIssue(branchName string) (*Ticket, error) {
	issueNumber := g.parseBranchName(branchName)
	if issueNumber == "" {
		return nil, fmt.Errorf("could not parse GitHub issue number from branch name: %s", branchName)
	}

	return g.fetchFromGitHub(issueNumber)
}

func (g *GitHubTracker) parseBranchName(branchName string) string {
	// Common prefixes to remove
	prefixes := []string{"feat/", "fix/", "bug/", "feature/", "hotfix/", "chore/", "feat-", "fix-", "bug-", "feature-", "hotfix-", "chore-"}
	normalizedBranch := branchName
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(normalizedBranch), p) {
			normalizedBranch = normalizedBranch[len(p):]
			break
		}
	}

	// Look for a number in the branch name. GitHub issues are typically just numbers.
	// We'll look for a standalone number or a number following a hash #
	re := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(\d+)(?:$|[^a-zA-Z0-9])`)
	match := re.FindStringSubmatch(normalizedBranch)
	if len(match) > 1 {
		return match[1]
	}

	// Also try just any digits if the above failed
	re = regexp.MustCompile(`\d+`)
	return re.FindString(normalizedBranch)
}

func (g *GitHubTracker) fetchFromGitHub(issueNumber string) (*Ticket, error) {
	if g.Token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN or GH_TOKEN not set")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s", g.RepoOwner, g.RepoName, issueNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var result struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Ticket{
		ID:    fmt.Sprintf("#%d", result.Number),
		URL:   result.HTMLURL,
		Title: result.Title,
	}, nil
}
