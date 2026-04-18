package bot

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the bot.
type Config struct {
	// GitHub configuration
	GitHub GitHubConfig `yaml:"github"`

	// Rules defines the set of automation rules the bot will apply.
	Rules RulesConfig `yaml:"rules"`
}

// GitHubConfig holds GitHub-specific settings.
type GitHubConfig struct {
	// Token is the GitHub API token used for authentication.
	Token string `yaml:"token"`

	// Owner is the GitHub organization or user that owns the repositories.
	Owner string `yaml:"owner"`

	// Repos is the list of repositories the bot should manage.
	// If empty, the bot will manage all repositories under Owner.
	Repos []string `yaml:"repos"`

	// APIURL allows overriding the GitHub API base URL (e.g. for GitHub Enterprise).
	APIURL string `yaml:"api_url"`
}

// RulesConfig defines the automation rules applied by the bot.
type RulesConfig struct {
	// AutoMerge enables automatic merging of PRs that meet all requirements.
	AutoMerge bool `yaml:"auto_merge"`

	// AutoMergeMethod is the merge method to use (merge, squash, rebase).
	AutoMergeMethod string `yaml:"auto_merge_method"`

	// RequiredApprovals is the minimum number of approvals before auto-merging.
	RequiredApprovals int `yaml:"required_approvals"`

	// Labels contains label-based automation rules.
	Labels LabelsConfig `yaml:"labels"`
}

// LabelsConfig defines label-based automation rules.
type LabelsConfig struct {
	// NeedsFeedback is the label applied when a PR or issue needs author feedback.
	NeedsFeedback string `yaml:"needs_feedback"`

	// WIP is the label that prevents auto-merging.
	WIP string `yaml:"wip"`

	// Approved is the label applied when a PR has been approved.
	Approved string `yaml:"approved"`
}

// LoadConfig reads and parses a YAML configuration file from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// validate checks that required configuration fields are set.
func (c *Config) validate() error {
	if c.GitHub.Token == "" {
		// Fall back to environment variable.
		c.GitHub.Token = os.Getenv("GITHUB_TOKEN")
	}
	if c.GitHub.Token == "" {
		return fmt.Errorf("github.token must be set or GITHUB_TOKEN env var must be provided")
	}
	if c.GitHub.Owner == "" {
		return fmt.Errorf("github.owner must be set")
	}
	return nil
}

// applyDefaults sets sensible default values for optional fields.
func (c *Config) applyDefaults() {
	if c.Rules.AutoMergeMethod == "" {
		c.Rules.AutoMergeMethod = "merge"
	}
	if c.Rules.RequiredApprovals == 0 {
		c.Rules.RequiredApprovals = 1
	}
	if c.Rules.Labels.NeedsFeedback == "" {
		c.Rules.Labels.NeedsFeedback = "needs-feedback"
	}
	if c.Rules.Labels.WIP == "" {
		c.Rules.Labels.WIP = "WIP"
	}
	if c.Rules.Labels.Approved == "" {
		c.Rules.Labels.Approved = "approved"
	}
}
