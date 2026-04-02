package git

import (
	"fmt"
	"strings"
)

// GetCommitSubjects collects commit subjects from merge-base to HEAD in chronological order.
func GetCommitSubjects(mergeBase string) ([]string, error) {
	// 'git log' with --reverse ensures chronological order (oldest first).
	// %s is the format for commit subjects.
	output, err := runGit("log", mergeBase+"..HEAD", "--format=%s", "--reverse")
	if err != nil {
		return nil, fmt.Errorf("failed to collect commit subjects: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}
