package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/output"
	"github.com/xx4h/nixconf/internal/runner"
)

var updateCmd = &cobra.Command{
	Use:   "update [INPUT ...]",
	Short: "Run nix flake update, then commit and push flake.lock in every repo",
	Long: `Run 'nix flake update' in every selected repo, commit the resulting
flake.lock change, and push.

If one or more INPUT names are given they are passed through to
'nix flake update INPUT ...' so only those inputs are bumped.`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(_ *cobra.Command, inputs []string) error {
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

		if flagDryRun {
			if len(inputs) > 0 {
				output.Infof("%s (%s) — would run: nix flake update %s, commit flake.lock, push",
					r.Name, r.Path, strings.Join(inputs, " "))
			} else {
				output.Infof("%s (%s) — would run: nix flake update, commit flake.lock, push", r.Name, r.Path)
			}
			continue
		}

		output.Infof("Updating %s (%s)", r.Name, r.Path)

		nixArgs := append([]string{"flake", "update"}, inputs...)
		nixArgs = append(nixArgs, "--flake", full)
		if err := runner.Run(runner.Exec("nix", nixArgs...)); err != nil {
			output.Warnf("%s — flake update failed: %v", r.Name, err)
			continue
		}

		if !hasFlakeLockChanges(full) {
			output.Fprintf(output.Stdout(), "  No changes after update\n")
			continue
		}

		commitMsg := "chore(deps): flake update"
		if len(inputs) > 0 {
			commitMsg = "chore(deps): flake update " + strings.Join(inputs, " ")
		}
		if err := runner.Run(runner.Git(full, "commit", "-v", "flake.lock", "-m", commitMsg, "-s", "-S")); err != nil {
			output.Warnf("%s — commit failed: %v", r.Name, err)
			continue
		}
		if err := runner.Run(runner.Git(full, "push")); err != nil {
			output.Warnf("%s — push failed: %v", r.Name, err)
			continue
		}
		output.Fprintf(output.Stdout(), "  %s\n", output.Success.Render("Updated and pushed"))
	}
	return nil
}

// hasFlakeLockChanges reports whether the repo has staged or unstaged diffs.
func hasFlakeLockChanges(repoPath string) bool {
	staged := runner.Git(repoPath, "diff", "--cached", "--quiet")
	unstaged := runner.Git(repoPath, "diff", "--quiet")
	return staged.Run() != nil || unstaged.Run() != nil
}
