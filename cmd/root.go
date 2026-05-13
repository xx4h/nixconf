package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/config"
	"github.com/xx4h/nixconf/internal/runner"
)

var (
	cfgFile string

	flagCommon bool
	flagHosts  bool
	flagUsers  bool
	flagRepo   string
	flagDryRun bool
)

var rootCmd = &cobra.Command{
	Use:   "nixconf",
	Short: "Repository manager for NixOS multi-repo configuration",
	Long: `nixconf manages a workspace of related NixOS configuration repos
(nixos-common, host repos, user repos) declared in nixconf.yaml.

Use 'nixconf [--hosts|--users|--common|-r NAME] git <args>' to run a git
command in every selected repo.`,
}

// Execute runs the root command. If the user invoked the `git` subcommand,
// the persistent-flag prefix is parsed in-place so the selector flags work
// even though gitCmd has DisableFlagParsing (which otherwise disables flag
// parsing for the whole command path, including root's persistent flags).
func Execute() {
	rewriteArgsForGitPassthrough()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// rewriteArgsForGitPassthrough detects an invocation of the form
//
//	nixconf [persistent-flags] git    [git-args...]
//	nixconf [persistent-flags] <sub>  [git-args...]   // sub in gitShortcuts
//
// and splits the arg list so cobra sees the persistent flags as belonging
// to the root command. Without this, DisableFlagParsing on the passthrough
// commands causes `nixconf --hosts git status` (or `nixconf --hosts push`)
// to forward `--hosts` to git verbatim.
func rewriteArgsForGitPassthrough() {
	args := os.Args[1:]

	idx := firstPositionalIndex(args)
	if idx < 0 || !isGitPassthroughKeyword(args[idx]) {
		return
	}

	persistent := args[:idx]
	remaining := args[idx:]

	if err := rootCmd.PersistentFlags().Parse(persistent); err != nil {
		// Let cobra surface the error in its usual format.
		return
	}

	rootCmd.SetArgs(remaining)
}

func isGitPassthroughKeyword(s string) bool {
	if s == "git" {
		return true
	}
	for _, sub := range gitShortcuts {
		if s == sub {
			return true
		}
	}
	return false
}

// firstPositionalIndex returns the index of the first non-flag argument,
// accounting for persistent flags that take a value when separated by a
// space (e.g. `--repo foo` or `-c path`).
func firstPositionalIndex(args []string) int {
	flagsWithValue := map[string]bool{
		"-r":       true,
		"--repo":   true,
		"-c":       true,
		"--config": true,
	}
	skipNext := false
	for i, a := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "--" {
			if i+1 < len(args) {
				return i + 1
			}
			return -1
		}
		if strings.HasPrefix(a, "-") {
			if strings.Contains(a, "=") {
				continue
			}
			if flagsWithValue[a] {
				skipNext = true
			}
			continue
		}
		return i
	}
	return -1
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to nixconf.yaml (default: search upwards from cwd, then $XDG_CONFIG_HOME/nixconf.yaml)")
	rootCmd.PersistentFlags().BoolVar(&flagCommon, "common", false, "only common repos")
	rootCmd.PersistentFlags().BoolVar(&flagHosts, "hosts", false, "only host config repos")
	rootCmd.PersistentFlags().BoolVar(&flagUsers, "users", false, "only user config repos")
	rootCmd.PersistentFlags().StringVarP(&flagRepo, "repo", "r", "", "only the named repo")
	rootCmd.PersistentFlags().BoolVarP(&flagDryRun, "dry-run", "n", false, "show what would be done without making changes")

	if err := rootCmd.RegisterFlagCompletionFunc("repo", completeRepoNames); err != nil {
		panic(fmt.Errorf("registering repo flag completion: %w", err))
	}
}

func loadConfig() (config.Config, error) {
	return config.LoadFromCwd(cfgFile)
}

func currentSelector() runner.Selector {
	return runner.Selector{
		Common: flagCommon,
		Hosts:  flagHosts,
		Users:  flagUsers,
		Repo:   flagRepo,
	}
}
