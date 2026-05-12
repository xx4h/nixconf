package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	URL      string `yaml:"url,omitempty"`
	Disabled bool   `yaml:"disabled,omitempty"`
}

type Repos struct {
	Common []Repo `yaml:"common,omitempty"`
	Hosts  []Repo `yaml:"hosts,omitempty"`
	Users  []Repo `yaml:"users,omitempty"`
}

type Config struct {
	GitBase string `yaml:"git_base,omitempty"`
	// DataDir is the directory under which repos are cloned. If empty, the
	// resolved value falls back to $XDG_DATA_HOME/nixconf (or
	// ~/.local/share/nixconf). Relative values are interpreted against the
	// directory holding nixconf.yaml.
	DataDir string `yaml:"data_dir,omitempty"`
	Repos   Repos  `yaml:"repos"`

	Root            string `yaml:"-"`
	Path            string `yaml:"-"`
	ResolvedDataDir string `yaml:"-"`
}

// Load reads nixconf.yaml from path. The directory containing the file is
// stored as Root so callers can resolve repo paths relative to it.
func Load(path string) (Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", abs, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", abs, err)
	}
	cfg.Root = filepath.Dir(abs)
	cfg.Path = abs

	resolved, err := cfg.resolveDataDir()
	if err != nil {
		return Config{}, err
	}
	cfg.ResolvedDataDir = resolved
	return cfg, nil
}

// DefaultDataDir returns $XDG_DATA_HOME/nixconf, falling back to
// ~/.local/share/nixconf when $XDG_DATA_HOME is unset.
func DefaultDataDir() (string, error) {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return filepath.Join(v, "nixconf"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "nixconf"), nil
}

// resolveDataDir applies the data_dir resolution rules: empty → default,
// absolute → as-is, relative → joined against Root.
func (c Config) resolveDataDir() (string, error) {
	if c.DataDir == "" {
		return DefaultDataDir()
	}
	if filepath.IsAbs(c.DataDir) {
		return c.DataDir, nil
	}
	return filepath.Join(c.Root, c.DataDir), nil
}

// Find walks up from startDir looking for nixconf.yaml. Returns "" if none.
func Find(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "nixconf.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// UserConfigPath returns $XDG_CONFIG_HOME/nixconf.yaml, falling back to
// ~/.config/nixconf.yaml when $XDG_CONFIG_HOME is unset.
func UserConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nixconf.yaml"), nil
}

// Resolve picks a config path without reading it. The resolution order:
//   - override, if non-empty
//   - first nixconf.yaml found walking up from cwd
//   - $XDG_CONFIG_HOME/nixconf.yaml (~/.config/nixconf.yaml fallback), if it exists
//
// Returns "" if nothing matched.
func Resolve(override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if p := Find(cwd); p != "" {
		return p, nil
	}
	user, err := UserConfigPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(user); err == nil {
		return user, nil
	}
	return "", nil
}

// LoadFromCwd locates nixconf.yaml starting at the working directory and
// loads it. If override is non-empty, it is used directly. If no config is
// found in the cwd hierarchy, $XDG_CONFIG_HOME/nixconf.yaml is tried.
func LoadFromCwd(override string) (Config, error) {
	path, err := Resolve(override)
	if err != nil {
		return Config{}, err
	}
	if path == "" {
		cwd, _ := os.Getwd()
		user, _ := UserConfigPath()
		return Config{}, fmt.Errorf("nixconf.yaml not found in %s or any parent directory, and %s does not exist", cwd, user)
	}
	return Load(path)
}

// Save writes the config back to its Path with stable indentation.
func (c Config) Save() error {
	if c.Path == "" {
		return fmt.Errorf("config has no path set; use SaveTo")
	}
	return c.SaveTo(c.Path)
}

// SaveTo writes the config to path with stable indentation.
func (c Config) SaveTo(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		_ = enc.Close()
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("closing encoder: %w", err)
	}

	if err := os.WriteFile(abs, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", abs, err)
	}
	return nil
}

// CloneURL returns the clone URL for r, falling back to ${git_base}/${name}.git.
func (c Config) CloneURL(r Repo) (string, error) {
	if r.URL != "" {
		return r.URL, nil
	}
	if c.GitBase != "" {
		return fmt.Sprintf("%s/%s.git", c.GitBase, r.Name), nil
	}
	return "", fmt.Errorf("repo %q: no 'url' set and no top-level 'git_base'", r.Name)
}

// FullPath resolves r.Path against the resolved data_dir.
func (c Config) FullPath(r Repo) string {
	return filepath.Join(c.ResolvedDataDir, r.Path)
}

// AllRepos returns every configured repo in declaration order.
func (c Config) AllRepos() []Repo {
	out := make([]Repo, 0, len(c.Repos.Common)+len(c.Repos.Hosts)+len(c.Repos.Users))
	out = append(out, c.Repos.Common...)
	out = append(out, c.Repos.Hosts...)
	out = append(out, c.Repos.Users...)
	return out
}
