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

Any subcommand that is not one of the built-in commands is forwarded to
'git -C <repo>' in every selected repo.`,
}

// Execute runs the root command. If the first positional argument is not a
// known subcommand, it is treated as a git subcommand and forwarded.
func Execute() {
	rewriteArgsForGitPassthrough()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to nixconf.yaml (default: search upwards from cwd)")
	rootCmd.PersistentFlags().BoolVar(&flagCommon, "common", false, "only common repos")
	rootCmd.PersistentFlags().BoolVar(&flagHosts, "hosts", false, "only host config repos")
	rootCmd.PersistentFlags().BoolVar(&flagUsers, "users", false, "only user config repos")
	rootCmd.PersistentFlags().StringVarP(&flagRepo, "repo", "r", "", "only the named repo")
	rootCmd.PersistentFlags().BoolVarP(&flagDryRun, "dry-run", "n", false, "show what would be done without making changes")

	_ = rootCmd.RegisterFlagCompletionFunc("repo", completeRepoNames)
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

// rewriteArgsForGitPassthrough finds the first positional argument in os.Args.
// If it is not a known cobra subcommand, the argument list is rewritten to
// invoke the internal `git` command instead. The persistent-flag prefix is
// parsed in-place so the values are populated before cobra dispatches to the
// (DisableFlagParsing) git command.
func rewriteArgsForGitPassthrough() {
	args := os.Args[1:]

	idx := firstPositionalIndex(args)
	if idx < 0 {
		return
	}
	if isKnownCommand(args[idx]) {
		return
	}

	persistent := args[:idx]
	gitArgs := args[idx:]

	if err := rootCmd.PersistentFlags().Parse(persistent); err != nil {
		// Let cobra surface the error in its usual format.
		return
	}

	rewritten := append([]string{"git"}, gitArgs...)
	rootCmd.SetArgs(rewritten)
}

// firstPositionalIndex returns the index of the first non-flag argument,
// accounting for persistent flags that take a value when separated by a space.
func firstPositionalIndex(args []string) int {
	flagsWithValue := map[string]bool{
		"-r":       true,
		"--repo":   true,
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

func isKnownCommand(name string) bool {
	if name == "help" || name == "completion" {
		return true
	}
	// Cobra's hidden completion-machinery commands start with "__".
	if strings.HasPrefix(name, "__") {
		return true
	}
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return true
		}
		for _, alias := range c.Aliases {
			if alias == name {
				return true
			}
		}
	}
	return false
}
