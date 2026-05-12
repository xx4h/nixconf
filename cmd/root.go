package cmd

import (
	"fmt"
	"os"

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

Use 'nixconf git -- <args>' to run a git command in every selected repo.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to nixconf.yaml (default: search upwards from cwd, then $XDG_CONFIG_HOME/nixconf.yaml)")
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
