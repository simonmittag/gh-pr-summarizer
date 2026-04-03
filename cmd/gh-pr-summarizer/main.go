package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
	"github.com/simonmittag/gh-pr-summarizer/internal/config"
	"github.com/simonmittag/gh-pr-summarizer/internal/git"
	"github.com/simonmittag/gh-pr-summarizer/internal/render"
	"github.com/simonmittag/gh-pr-summarizer/internal/tracker"
)

var (
	version = "dev"
)

func main() {
	_ = godotenv.Load()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	if os.Getenv("ZEROLOG_LEVEL") != "" {
		level, err := zerolog.ParseLevel(os.Getenv("ZEROLOG_LEVEL"))
		if err == nil {
			zerolog.SetGlobalLevel(level)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Str("version", version).Msg("starting gh-pr-summarizer")

	currentBranchFlag := flag.String("current", "", "override detected current branch")
	baseBranchFlag := flag.String("base", "", "override detected base branch")
	draftFlag := flag.Bool("draft", false, "mark PR as draft (adds 🚧 emoji to title)")
	versionFlag := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("gh-pr-summarizer version: %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to parse configuration file, check ./ghpr.toml or ~/.config/gh-pr-summarizer/config")
		os.Exit(-1)
	}

	gitCtx, err := git.GetContext()
	if err != nil {
		log.Debug().Err(err).Msg("error getting git context")
		os.Exit(1)
	}

	if *currentBranchFlag != "" {
		gitCtx.CurrentBranch = *currentBranchFlag
	}
	if *baseBranchFlag != "" {
		gitCtx.BaseBranch = *baseBranchFlag
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	var aiClient *openai.Client
	if apiKey != "" {
		aiClient = openai.NewClient(apiKey)
		log.Debug().Err(err).Msg("proceeding with openai integration configured")
	}
	renderer := render.NewRenderer(aiClient)
	renderer.Draft = *draftFlag

	mergeBase, err := gitCtx.GetMergeBase()
	if err != nil {
		log.Debug().Err(err).Msg("error computing merge-base")
		os.Exit(1)
	}

	subjects, err := git.GetCommitSubjects(mergeBase)
	if err != nil {
		log.Debug().Err(err).Msg("error collecting commit subjects")
		os.Exit(1)
	}

	var ticket *tracker.Ticket
	switch cfg.Tracker {
	case "linear":
		tr := tracker.NewLinearTracker(cfg.Linear.TicketUrlStem, os.Getenv(cfg.Linear.TokenEnv))
		t, err := tr.FetchTicket(gitCtx.CurrentBranch)
		if err != nil {
			log.Debug().Err(err).Msg("unable to fetch ticket from linear, proceeding without ticket")
		} else {
			ticket = t
		}
	case "github":
		owner, repo, err := gitCtx.GetRemoteOwnerRepo()
		if err != nil {
			log.Debug().Err(err).Msg("unable to detect github repo, proceeding without repo access")
		} else {
			tr := tracker.NewGitHubTracker(owner, repo, os.Getenv(cfg.GitHub.TokenEnv))
			t, err := tr.FetchTicket(gitCtx.CurrentBranch)
			if err != nil {
				log.Debug().Err(err).Msg("unable to fetch ticket from github, proceeding without ticket")
			} else {
				ticket = t
			}
		}
	case "jira":
		tr := tracker.NewJiraTracker(cfg.Jira.TicketUrlStem, os.Getenv(cfg.Jira.TokenEnv), os.Getenv(cfg.Jira.EmailEnv))
		t, err := tr.FetchTicket(gitCtx.CurrentBranch)
		if err != nil {
			log.Debug().Err(err).Msg("unable to fetch ticket from jira, proceeding without ticket")
		} else {
			ticket = t
		}
	}

	markdown := renderer.PRBody(subjects, ticket, gitCtx.CurrentBranch)
	log.Debug().Msg("successfully generated PR markdown")
	fmt.Print(markdown)
}
