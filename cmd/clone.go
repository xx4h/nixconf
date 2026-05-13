package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/output"
	"github.com/xx4h/nixconf/internal/runner"
)

var cloneCmd = &cobra.Command{
	Use:               "clone",
	Short:             "Clone all configured repos into their directories",
	Args:              cobra.NoArgs,
	ValidArgsFunction: completeFlagsOnly,
	RunE:              runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

func runClone(_ *cobra.Command, _ []string) error {
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

		if runner.IsCloned(cfg, r) {
			output.Infof("%s (%s) — already cloned, skipping", r.Name, r.Path)
			continue
		}

		url, err := cfg.CloneURL(r)
		if err != nil {
			output.Warnf("%s (%s) — %v", r.Name, r.Path, err)
			continue
		}

		if flagDryRun {
			output.Infof("%s (%s) — would clone from %s", r.Name, r.Path, url)
			continue
		}

		output.Infof("Cloning %s → %s", r.Name, r.Path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := runner.Run(exec.Command("git", "clone", url, full)); err != nil {
			output.Warnf("%s — git clone failed: %v", r.Name, err)
		}
	}
	return nil
}
