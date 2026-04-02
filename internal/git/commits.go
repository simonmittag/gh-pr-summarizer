package git

import (
	"strings"
)

// GetCommitSubjects collects commit subjects from merge-base to HEAD in chronological order.
// If git fails (e.g., not a repository), it returns a list with a single placeholder.
func GetCommitSubjects(mergeBase string) ([]string, error) {
	// 'git log' with --reverse ensures chronological order (oldest first).
	// %s is the format for commit subjects.
	output, err := runGit("log", mergeBase+"..HEAD", "--format=%s", "--reverse")
	if err != nil {
		return make([]string, 0), nil
	}

	if output == "" {
		return make([]string, 0), nil
	}

	return strings.Split(output, "\n"), nil
}
