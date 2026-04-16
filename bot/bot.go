// Package bot provides the core GitHub bot functionality for lerobot.
package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

// Bot represents the lerobot instance with its configuration and GitHub client.
type Bot struct {
	client *github.Client
	owner  string
	repo   string
	log    *log.Logger
}

// Config holds the configuration for creating a new Bot.
type Config struct {
	Token  string
	Owner  string
	Repo   string
	Logger *log.Logger
}

// New creates a new Bot instance using the provided configuration.
func New(cfg Config) (*Bot, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("github token must not be empty")
	}
	if cfg.Owner == "" || cfg.Repo == "" {
		return nil, fmt.Errorf("owner and repo must not be empty")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}

	return &Bot{
		client: client,
		owner:  cfg.Owner,
		repo:   cfg.Repo,
		log:    logger,
	}, nil
}

// HandleWebhook processes an incoming GitHub webhook request.
func (b *Bot) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, nil)
	if err != nil {
		b.log.Printf("error validating webhook payload: %v", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		b.log.Printf("error parsing webhook: %v", err)
		http.Error(w, "could not parse webhook", http.StatusBadRequest)
		return
	}

	switch e := event.(type) {
	case *github.IssueCommentEvent:
		if err := b.handleIssueComment(r.Context(), e); err != nil {
			b.log.Printf("error handling issue comment: %v", err)
		}
	case *github.PullRequestEvent:
		if err := b.handlePullRequest(r.Context(), e); err != nil {
			b.log.Printf("error handling pull request event: %v", err)
		}
	default:
		b.log.Printf("unhandled event type: %T", e)
	}

	w.WriteHeader(http.StatusOK)
}

// handleIssueComment processes issue comment events.
func (b *Bot) handleIssueComment(ctx context.Context, event *github.IssueCommentEvent) error {
	if event.GetAction() != "created" {
		return nil
	}
	b.log.Printf("issue comment on #%d by %s", event.GetIssue().GetNumber(), event.GetComment().GetUser().GetLogin())
	return nil
}

// handlePullRequest processes pull request events.
func (b *Bot) handlePullRequest(ctx context.Context, event *github.PullRequestEvent) error {
	b.log.Printf("pull request #%d action=%s", event.GetPullRequest().GetNumber(), event.GetAction())
	return nil
}
