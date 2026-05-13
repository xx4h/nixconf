package cmd

import (
	"github.com/spf13/cobra"
)

// gitShortcuts are top-level aliases for common `nixconf git <sub>` invocations.
// Each entry registers `nixconf <sub>` as equivalent to `nixconf git <sub>`,
// forwarding remaining args to git unaltered.
var gitShortcuts = []string{"push", "pull", "status"}

func init() {
	for _, sub := range gitShortcuts {
		sub := sub
		c := &cobra.Command{
			Use:                sub + " [args...]",
			Short:              "Shortcut for 'nixconf git " + sub + "'",
			DisableFlagParsing: true,
			RunE: func(_ *cobra.Command, args []string) error {
				return runGitArgs(append([]string{sub}, args...))
			},
		}
		rootCmd.AddCommand(c)
	}
}
