package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GitContext holds information about the current git repository and branches.
type GitContext struct {
	CurrentBranch string
	BaseBranch    string
}

// GetContext detects the current git repository, current branch, and default base branch.
// If not inside a git repository, it returns a default context with placeholders.
func GetContext() (*GitContext, error) {
	// 1. Detect if inside a git repository.
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return &GitContext{
			CurrentBranch: "feature-branch-placeholder",
			BaseBranch:    "main",
		}, nil
	}

	// 2. Detect current branch name.
	currentBranch, err := runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to detect current branch: %w", err)
	}

	// 3. Detect default base branch (main if present, otherwise master).
	baseBranch, err := detectBaseBranch()
	if err != nil {
		return nil, err
	}

	return &GitContext{
		CurrentBranch: currentBranch,
		BaseBranch:    baseBranch,
	}, nil
}

// GetMergeBase computes the merge-base between current branch and base branch.
// If git fails (e.g., not a repository), it returns a placeholder.
func (c *GitContext) GetMergeBase() (string, error) {
	mergeBase, err := runGit("merge-base", c.BaseBranch, c.CurrentBranch)
	if err != nil {
		return "", nil
	}
	return mergeBase, nil
}

func (c *GitContext) GetRemoteOwnerRepo() (string, string, error) {
	output, err := runGit("remote", "get-url", "origin")
	if err != nil {
		return "", "", err
	}
	// Support both SSH and HTTPS formats
	// SSH: git@github.com:owner/repo.git
	// HTTPS: https://github.com/owner/repo.git
	output = strings.TrimSuffix(output, ".git")
	parts := strings.Split(output, ":")
	if len(parts) > 1 {
		// SSH format or HTTPS with port
		path := parts[len(parts)-1]
		pathParts := strings.Split(path, "/")
		if len(pathParts) >= 2 {
			return pathParts[len(pathParts)-2], pathParts[len(pathParts)-1], nil
		}
	} else {
		// HTTPS format
		parts = strings.Split(output, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-2], parts[len(parts)-1], nil
		}
	}
	return "", "", fmt.Errorf("could not parse owner and repo from remote URL: %s", output)
}

func detectBaseBranch() (string, error) {
	// Check if 'main' exists.
	if err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/main").Run(); err == nil {
		return "main", nil
	}
	// Check if 'master' exists.
	if err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/master").Run(); err == nil {
		return "master", nil
	}
	// If neither exists, error clearly.
	return "", errors.New("neither 'main' nor 'master' branch found")
}

func runGit(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git error: %s: %w", stderr.String(), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
