package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/xx4h/nixconf/internal/config"
	"github.com/xx4h/nixconf/internal/runner"
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

// completeFlagsOnly returns the command's long-form flag names as candidates
// and suppresses file completion. Use as ValidArgsFunction on commands whose
// positional slot has nothing useful to offer, so <TAB><TAB> surfaces flags
// instead of silently adding a space.
func completeFlagsOnly(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return flagNames(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// flagNames lists non-hidden long flag names (`--foo`) of cmd, including
// inherited persistent flags, filtered by the toComplete prefix.
func flagNames(cmd *cobra.Command, toComplete string) []string {
	var out []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		name := "--" + f.Name
		if strings.HasPrefix(name, toComplete) {
			out = append(out, name)
		}
	})
	sort.Strings(out)
	return out
}

// completeUpdateInputs offers the flake input names of the repo selected with
// --repo. Without --repo we fall back to flag suggestions — completing across
// multiple repos would suggest inputs that don't exist in some of them and
// silently misapply the update.
func completeUpdateInputs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if flagRepo == "" {
		return flagNames(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	cfg, err := config.LoadFromCwd(cfgFile)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	repos, err := runner.Select(cfg, currentSelector())
	if err != nil || len(repos) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	r := repos[0]
	if !runner.IsCloned(cfg, r) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	already := make(map[string]struct{}, len(args))
	for _, a := range args {
		already[a] = struct{}{}
	}
	var out []string
	for _, name := range flakeRootInputs(cfg.FullPath(r)) {
		if _, dup := already[name]; dup {
			continue
		}
		if strings.HasPrefix(name, toComplete) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, cobra.ShellCompDirectiveNoFileComp
}

// flakeRootInputs returns the root-level input names declared in
// <repoPath>/flake.lock, or nil if the file is missing or unreadable.
func flakeRootInputs(repoPath string) []string {
	data, err := os.ReadFile(filepath.Join(repoPath, "flake.lock"))
	if err != nil {
		return nil
	}
	var lock struct {
		Root  string `json:"root"`
		Nodes map[string]struct {
			Inputs map[string]json.RawMessage `json:"inputs"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil
	}
	root := lock.Root
	if root == "" {
		root = "root"
	}
	node, ok := lock.Nodes[root]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(node.Inputs))
	for name := range node.Inputs {
		out = append(out, name)
	}
	return out
}
