package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/output"
	"github.com/xx4h/nixconf/internal/runner"
)

// gitCmd forwards its arguments to `git -C <repo>` in every selected repo.
// Flag parsing is disabled so any git flag (e.g. `--oneline`, `-p`) reaches
// git unaltered; nixconf's own flags must appear before the `git` keyword.
var gitCmd = &cobra.Command{
	Use:   "git [args...]",
	Short: "Run a git command in every selected repo",
	Long: `Run an arbitrary git command in every selected repo.

All arguments after 'git' are passed through to 'git -C <repo>' without
further interpretation by nixconf. Place nixconf's own flags before the
'git' keyword:

  nixconf --hosts git status
  nixconf -r nixos-common git log --oneline -5`,
	DisableFlagParsing: true,
	RunE:               runGit,
}

func init() {
	rootCmd.AddCommand(gitCmd)
}

func runGit(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no git command specified")
	}
	return runGitArgs(args)
}

// runGitArgs runs `git args...` in every selected repo. args[0] must be the
// git subcommand; the dry-run flag is inserted right after it.
func runGitArgs(args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	repos, err := runner.Select(cfg, currentSelector())
	if err != nil {
		return err
	}

	for _, r := range repos {
		full := cfg.FullPath(r)
		if !runner.IsCloned(cfg, r) {
			output.Warnf("%s (%s) — not cloned, skipping", r.Name, r.Path)
			continue
		}

		output.Infof("%s (%s)", r.Name, r.Path)

		gitArgs := args
		if flagDryRun {
			gitArgs = append([]string{args[0], "--dry-run"}, args[1:]...)
		}
		if err := runner.Run(runner.Git(full, gitArgs...)); err != nil {
			output.Warnf("%s — git %s failed: %v", r.Name, args[0], err)
		}
	}
	return nil
}
