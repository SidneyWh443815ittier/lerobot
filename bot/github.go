package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

// GitHubClient wraps the GitHub API client with helper methods
// used by the bot to interact with pull requests and issues.
type GitHubClient struct {
	client *github.Client
	owner  string
	repo   string
}

// NewGitHubClient creates an authenticated GitHub client using the provided token.
func NewGitHubClient(ctx context.Context, token, owner, repo string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &GitHubClient{
		client: github.NewClient(tc),
		owner:  owner,
		repo:   repo,
	}
}

// AddLabel adds a label to the given issue or pull request number.
func (g *GitHubClient) AddLabel(ctx context.Context, number int, label string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(ctx, g.owner, g.repo, number, []string{label})
	if err != nil {
		return fmt.Errorf("adding label %q to #%d: %w", label, number, err)
	}
	log.Printf("Added label %q to #%d", label, number)
	return nil
}

// RemoveLabel removes a label from the given issue or pull request number.
func (g *GitHubClient) RemoveLabel(ctx context.Context, number int, label string) error {
	_, err := g.client.Issues.RemoveLabelForIssue(ctx, g.owner, g.repo, number, label)
	if err != nil {
		return fmt.Errorf("removing label %q from #%d: %w", label, number, err)
	}
	log.Printf("Removed label %q from #%d", label, number)
	return nil
}

// PostComment posts a comment on the given issue or pull request.
func (g *GitHubClient) PostComment(ctx context.Context, number int, body string) error {
	comment := &github.IssueComment{Body: github.String(body)}
	_, _, err := g.client.Issues.CreateComment(ctx, g.owner, g.repo, number, comment)
	if err != nil {
		return fmt.Errorf("posting comment on #%d: %w", number, err)
	}
	log.Printf("Posted comment on #%d", number)
	return nil
}

// GetPullRequest retrieves a pull request by number.
func (g *GitHubClient) GetPullRequest(ctx context.Context, number int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, g.owner, g.repo, number)
	if err != nil {
		return nil, fmt.Errorf("getting PR #%d: %w", number, err)
	}
	return pr, nil
}

// ListOpenPullRequests returns all open pull requests for the repository.
func (g *GitHubClient) ListOpenPullRequests(ctx context.Context) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allPRs []*github.PullRequest
	for {
		prs, resp, err := g.client.PullRequests.List(ctx, g.owner, g.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("listing open PRs: %w", err)
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allPRs, nil
}

// RequestReview requests a review from the specified reviewers on a pull request.
func (g *GitHubClient) RequestReview(ctx context.Context, number int, reviewers []string) error {
	reviewReq := github.ReviewersRequest{Reviewers: reviewers}
	_, _, err := g.client.PullRequests.RequestReviewers(ctx, g.owner, g.repo, number, reviewReq)
	if err != nil {
		return fmt.Errorf("requesting reviewers on #%d: %w", number, err)
	}
	log.Printf("Requested review from %v on #%d", reviewers, number)
	return nil
}
