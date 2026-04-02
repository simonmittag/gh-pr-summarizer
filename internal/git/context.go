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
