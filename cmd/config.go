package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xx4h/nixconf/internal/config"
	"github.com/xx4h/nixconf/internal/output"
)

var (
	flagCfgPath string
	flagCfgURL  string
	flagCfgName string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage entries in nixconf.yaml",
	Long: `Manage entries in nixconf.yaml.

Subcommands operate on a kind (host or user) and a repo name. Changes are
written back to the same nixconf.yaml that 'nixconf' would otherwise load.`,
}

var configAddCmd = &cobra.Command{
	Use:               "add <host|user> <name>",
	Short:             "Add a host or user entry",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeKindThenNew,
	RunE:              runConfigAdd,
}

var configEditCmd = &cobra.Command{
	Use:               "edit <host|user> <name>",
	Short:             "Edit fields on an existing host or user entry",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeKindThenExisting,
	RunE:              runConfigEdit,
}

var configDeleteCmd = &cobra.Command{
	Use:               "delete <host|user> <name>",
	Aliases:           []string{"rm", "remove"},
	Short:             "Remove a host or user entry",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeKindThenExisting,
	RunE:              runConfigDelete,
}

var configDisableCmd = &cobra.Command{
	Use:               "disable <host|user> <name>",
	Short:             "Mark a host or user entry as disabled (skipped by other commands)",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeKindThenExisting,
	RunE:              runConfigSetDisabled(true),
}

var configEnableCmd = &cobra.Command{
	Use:               "enable <host|user> <name>",
	Short:             "Re-enable a previously disabled host or user entry",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeKindThenExisting,
	RunE:              runConfigSetDisabled(false),
}

func init() {
	configAddCmd.Flags().StringVar(&flagCfgPath, "path", "", "filesystem path for the repo (default: <kind>s/<name>)")
	configAddCmd.Flags().StringVar(&flagCfgURL, "url", "", "explicit clone URL (overrides git_base)")

	configEditCmd.Flags().StringVar(&flagCfgName, "name", "", "rename the entry")
	configEditCmd.Flags().StringVar(&flagCfgPath, "path", "", "set the filesystem path")
	configEditCmd.Flags().StringVar(&flagCfgURL, "url", "", "set the clone URL (empty string clears it)")

	configCmd.AddCommand(configAddCmd, configEditCmd, configDeleteCmd, configDisableCmd, configEnableCmd)
	rootCmd.AddCommand(configCmd)
}

// kind is "host" or "user". Plural forms are accepted as aliases.
func parseKind(s string) (string, error) {
	switch strings.ToLower(s) {
	case "host", "hosts":
		return "host", nil
	case "user", "users":
		return "user", nil
	default:
		return "", fmt.Errorf("unknown kind %q (want: host, user)", s)
	}
}

// repoSlice returns a pointer to the slice in cfg.Repos that backs kind.
func repoSlice(cfg *config.Config, kind string) *[]config.Repo {
	if kind == "host" {
		return &cfg.Repos.Hosts
	}
	return &cfg.Repos.Users
}

func findRepo(slice []config.Repo, name string) int {
	for i, r := range slice {
		if r.Name == name {
			return i
		}
	}
	return -1
}

func loadForEdit() (config.Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return cfg, err
	}
	if cfg.Path == "" {
		return cfg, fmt.Errorf("loaded config has no on-disk path")
	}
	return cfg, nil
}

func runConfigAdd(cmd *cobra.Command, args []string) error {
	kind, err := parseKind(args[0])
	if err != nil {
		return err
	}
	name := args[1]

	cfg, err := loadForEdit()
	if err != nil {
		return err
	}

	slice := repoSlice(&cfg, kind)
	if findRepo(*slice, name) >= 0 {
		return fmt.Errorf("%s %q already exists in %s", kind, name, cfg.Path)
	}

	p := flagCfgPath
	if p == "" {
		p = path.Join(kind+"s", name)
	}
	entry := config.Repo{Name: name, Path: p}
	if cmd.Flags().Changed("url") {
		entry.URL = flagCfgURL
	}

	*slice = append(*slice, entry)
	if err := cfg.Save(); err != nil {
		return err
	}
	output.Infof("Added %s %q (%s)", kind, name, p)
	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	kind, err := parseKind(args[0])
	if err != nil {
		return err
	}
	name := args[1]

	cfg, err := loadForEdit()
	if err != nil {
		return err
	}

	slice := repoSlice(&cfg, kind)
	idx := findRepo(*slice, name)
	if idx < 0 {
		return fmt.Errorf("%s %q not found in %s", kind, name, cfg.Path)
	}

	changed := false
	if cmd.Flags().Changed("name") && flagCfgName != "" && flagCfgName != (*slice)[idx].Name {
		if findRepo(*slice, flagCfgName) >= 0 {
			return fmt.Errorf("%s %q already exists", kind, flagCfgName)
		}
		(*slice)[idx].Name = flagCfgName
		changed = true
	}
	if cmd.Flags().Changed("path") {
		(*slice)[idx].Path = flagCfgPath
		changed = true
	}
	if cmd.Flags().Changed("url") {
		(*slice)[idx].URL = flagCfgURL
		changed = true
	}

	if !changed {
		return fmt.Errorf("nothing to change (pass --name, --path, or --url)")
	}

	if err := cfg.Save(); err != nil {
		return err
	}
	output.Infof("Updated %s %q", kind, (*slice)[idx].Name)
	return nil
}

func runConfigDelete(_ *cobra.Command, args []string) error {
	kind, err := parseKind(args[0])
	if err != nil {
		return err
	}
	name := args[1]

	cfg, err := loadForEdit()
	if err != nil {
		return err
	}

	slice := repoSlice(&cfg, kind)
	idx := findRepo(*slice, name)
	if idx < 0 {
		return fmt.Errorf("%s %q not found in %s", kind, name, cfg.Path)
	}

	*slice = append((*slice)[:idx], (*slice)[idx+1:]...)
	if err := cfg.Save(); err != nil {
		return err
	}
	output.Infof("Removed %s %q", kind, name)
	return nil
}

func runConfigSetDisabled(disabled bool) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		kind, err := parseKind(args[0])
		if err != nil {
			return err
		}
		name := args[1]

		cfg, err := loadForEdit()
		if err != nil {
			return err
		}

		slice := repoSlice(&cfg, kind)
		idx := findRepo(*slice, name)
		if idx < 0 {
			return fmt.Errorf("%s %q not found in %s", kind, name, cfg.Path)
		}

		if (*slice)[idx].Disabled == disabled {
			state := "enabled"
			if disabled {
				state = "disabled"
			}
			output.Infof("%s %q is already %s", kind, name, state)
			return nil
		}

		(*slice)[idx].Disabled = disabled
		if err := cfg.Save(); err != nil {
			return err
		}
		verb := "Enabled"
		if disabled {
			verb = "Disabled"
		}
		output.Infof("%s %s %q", verb, kind, name)
		return nil
	}
}

// completeKindThenNew offers "host"/"user" for the first arg; nothing for the second.
func completeKindThenNew(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return []string{"host", "user"}, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeKindThenExisting offers "host"/"user" for the first arg, and the
// existing names of that kind for the second.
func completeKindThenExisting(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return []string{"host", "user"}, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		kind, err := parseKind(args[0])
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		cfg, err := config.LoadFromCwd(cfgFile)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var names []string
		slice := *repoSlice(&cfg, kind)
		for _, r := range slice {
			if strings.HasPrefix(r.Name, toComplete) {
				names = append(names, r.Name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
