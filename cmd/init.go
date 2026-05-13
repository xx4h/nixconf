package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/config"
	"github.com/xx4h/nixconf/internal/output"
)

var flagInitForce bool

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a starter nixconf.yaml",
	Long: `Create a starter nixconf.yaml. If no path is given, write to
$XDG_CONFIG_HOME/nixconf.yaml. Refuses to overwrite an existing file unless --force.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&flagInitForce, "force", "f", false, "overwrite an existing file")
	rootCmd.AddCommand(initCmd)
}

const initTemplate = `# nixconf.yaml — managed by ` + "`nixconf`" + `
#
# git_base provides the default git remote prefix. Per-repo ` + "`url`" + ` overrides it.
git_base: ""

# data_dir is where repos are cloned. If omitted, defaults to
# $XDG_DATA_HOME/nixconf (or ~/.local/share/nixconf). Relative values are
# resolved against the directory holding this file.
# data_dir: ""

# update controls how ` + "`nixconf update`" + ` commits the bumped flake.lock.
# commit_message: ` + "`{{inputs}}`" + ` is replaced with " <input1> <input2>" when
# INPUTs are passed on the command line, and with "" otherwise.
# commit_flags: extra args appended to ` + "`git commit`" + `. An explicit empty list
# disables -v/-s/-S (the built-in defaults).
# update:
#   commit_message: "chore(deps): flake update{{inputs}}"
#   commit_flags: ["-v", "-s", "-S"]

repos:
  common: []
  hosts: []
  users: []
`

func runInit(_ *cobra.Command, args []string) error {
	path, err := resolveInitPath(args)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil && !flagInitForce {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(initTemplate), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	output.Infof("Created %s", path)
	return nil
}

func resolveInitPath(args []string) (string, error) {
	if len(args) == 1 {
		return filepath.Abs(args[0])
	}
	return config.UserConfigPath()
}
