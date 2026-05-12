package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/config"
)

// completeRepoNames offers the configured repo names for the --repo flag.
func completeRepoNames(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.LoadFromCwd(cfgFile)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, r := range cfg.AllRepos() {
		if strings.HasPrefix(r.Name, toComplete) {
			names = append(names, r.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
