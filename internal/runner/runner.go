package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xx4h/nixconf/internal/config"
)

// Selector describes which repos a command should act on.
type Selector struct {
	Common bool
	Hosts  bool
	Users  bool
	Repo   string
	// IncludeDisabled keeps repos with disabled=true in the result.
	// Default behavior filters them out so they aren't acted upon by
	// clone/update/verify/git-passthrough.
	IncludeDisabled bool
}

// Select returns the repos matching s. If no group is set and Repo is empty,
// every repo is returned. Disabled repos are filtered out unless an explicit
// --repo selector names one or s.IncludeDisabled is set.
func Select(cfg config.Config, s Selector) ([]config.Repo, error) {
	all := !s.Common && !s.Hosts && !s.Users

	var repos []config.Repo
	if all || s.Common {
		repos = append(repos, cfg.Repos.Common...)
	}
	if all || s.Hosts {
		repos = append(repos, cfg.Repos.Hosts...)
	}
	if all || s.Users {
		repos = append(repos, cfg.Repos.Users...)
	}

	if s.Repo != "" {
		for _, r := range repos {
			if r.Name == s.Repo {
				return []config.Repo{r}, nil
			}
		}
		return nil, fmt.Errorf("repo %q not found in nixconf.yaml", s.Repo)
	}

	if s.IncludeDisabled {
		return repos, nil
	}
	kept := repos[:0]
	for _, r := range repos {
		if !r.Disabled {
			kept = append(kept, r)
		}
	}
	return kept, nil
}

// IsCloned reports whether the repo's working tree contains a .git directory.
func IsCloned(cfg config.Config, r config.Repo) bool {
	info, err := os.Stat(filepath.Join(cfg.FullPath(r), ".git"))
	return err == nil && info.IsDir()
}

// Run executes cmd inheriting stdio and returns its exit error.
func Run(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Git builds `git -C <repoPath> args...` ready to run.
func Git(repoPath string, args ...string) *exec.Cmd {
	full := append([]string{"-C", repoPath}, args...)
	return exec.Command("git", full...)
}

// Exec is a thin wrapper around exec.Command for symmetry with Git.
func Exec(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
