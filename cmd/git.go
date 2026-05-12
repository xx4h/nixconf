package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/output"
	"github.com/xx4h/nixconf/internal/runner"
)

// gitCmd handles every non-built-in subcommand by forwarding to
// `git -C <repo> <args...>` in each selected repo. It is invoked by
// rewriteArgsForGitPassthrough — users do not type `nixconf git ...` directly,
// although they could.
var gitCmd = &cobra.Command{
	Use:                "git [args...]",
	Short:              "Forward arbitrary git commands to every selected repo",
	Hidden:             true,
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
		_ = runner.Run(runner.Git(full, gitArgs...))
	}
	return nil
}
