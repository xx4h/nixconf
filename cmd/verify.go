package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/config"
	"github.com/xx4h/nixconf/internal/output"
	"github.com/xx4h/nixconf/internal/runner"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Check that host/user repos follow the latest nixos-common",
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(_ *cobra.Command, _ []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Repos.Common) == 0 {
		return fmt.Errorf("no common repo configured in nixconf.yaml")
	}
	common := cfg.Repos.Common[0]
	commonPath := cfg.FullPath(common)

	if !runner.IsCloned(cfg, common) {
		return fmt.Errorf("%s not cloned at %s", common.Name, commonPath)
	}

	latest, err := gitRevParseHead(commonPath)
	if err != nil {
		return fmt.Errorf("reading %s HEAD: %w", common.Name, err)
	}
	latestShort := shortRev(latest)

	output.Infof("%s latest: %s (%s)", common.Name, latestShort, latest)
	fmt.Println()

	allOK := true
	for _, r := range append([]config.Repo{}, append(cfg.Repos.Hosts, cfg.Repos.Users...)...) {
		if flagRepo != "" && r.Name != flagRepo {
			continue
		}

		if !runner.IsCloned(cfg, r) {
			output.Warnf("%s (%s) — not cloned, skipping", r.Name, r.Path)
			continue
		}

		lockPath := filepath.Join(cfg.FullPath(r), "flake.lock")
		data, err := os.ReadFile(lockPath)
		if err != nil {
			output.Warnf("%s (%s) — no flake.lock found", r.Name, r.Path)
			continue
		}

		locked, err := extractCommonRev(data, common.Name)
		if err != nil || locked == "" {
			fmt.Printf("  %s  %s — no %s input in flake.lock\n",
				output.Skip.Render("SKIP"), r.Name, common.Name)
			continue
		}

		lockedShort := shortRev(locked)
		switch locked {
		case latest:
			fmt.Printf("  %s    %s — %s\n", output.Success.Render("OK"), r.Name, lockedShort)
		default:
			fmt.Printf("  %s %s — %s (expected %s)\n",
				output.Stale.Render("STALE"), r.Name, lockedShort, latestShort)
			allOK = false
		}
	}

	fmt.Println()
	if allOK {
		fmt.Println(output.Success.Render(fmt.Sprintf("All repos follow the latest %s.", common.Name)))
		return nil
	}
	fmt.Println(output.Skip.Render("Some repos are behind. Run 'nixconf update' to update them."))
	os.Exit(1)
	return nil
}

func gitRevParseHead(repoPath string) (string, error) {
	out, err := runner.Git(repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func shortRev(rev string) string {
	if len(rev) < 7 {
		return rev
	}
	return rev[:7]
}

// extractCommonRev parses flake.lock and returns the locked rev of the input
// whose node name matches commonName.
func extractCommonRev(data []byte, commonName string) (string, error) {
	var lock struct {
		Nodes map[string]struct {
			Locked struct {
				Rev string `json:"rev"`
			} `json:"locked"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		return "", err
	}
	node, ok := lock.Nodes[commonName]
	if !ok {
		return "", nil
	}
	return node.Locked.Rev, nil
}
