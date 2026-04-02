package main

import (
	"fmt"
	"os"

	"github.com/simonmittag/gh-pr-summarizer/internal/config"
	"github.com/simonmittag/gh-pr-summarizer/internal/git"
	"github.com/simonmittag/gh-pr-summarizer/internal/render"
	"github.com/simonmittag/gh-pr-summarizer/internal/tracker"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx, err := git.GetContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	mergeBase, err := ctx.GetMergeBase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing merge-base: %v\n", err)
		os.Exit(1)
	}

	subjects, err := git.GetCommitSubjects(mergeBase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting commit subjects: %v\n", err)
		os.Exit(1)
	}

	var ticket *tracker.Ticket
	if cfg.Tracker == "linear" {
		tr := tracker.NewLinearTracker(cfg.IssueUrlStem)
		t, err := tr.FetchIssue(ctx.CurrentBranch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch issue from linear: %v\n", err)
		} else {
			ticket = t
		}
	}

	markdown := render.PRBody(subjects, ticket)
	fmt.Print(markdown)
}
