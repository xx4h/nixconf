package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	URL  string `yaml:"url,omitempty"`
}

type Repos struct {
	Common []Repo `yaml:"common"`
	Hosts  []Repo `yaml:"hosts"`
	Users  []Repo `yaml:"users"`
}

type Config struct {
	GitBase string `yaml:"git_base"`
	Repos   Repos  `yaml:"repos"`

	Root string `yaml:"-"`
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
	return cfg, nil
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

// LoadFromCwd locates nixconf.yaml starting at the working directory and
// loads it. If override is non-empty, it is used directly.
func LoadFromCwd(override string) (Config, error) {
	if override != "" {
		return Load(override)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, err
	}
	path := Find(cwd)
	if path == "" {
		return Config{}, fmt.Errorf("nixconf.yaml not found in %s or any parent directory", cwd)
	}
	return Load(path)
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

// FullPath resolves r.Path against the directory containing nixconf.yaml.
func (c Config) FullPath(r Repo) string {
	return filepath.Join(c.Root, r.Path)
}

// AllRepos returns every configured repo in declaration order.
func (c Config) AllRepos() []Repo {
	out := make([]Repo, 0, len(c.Repos.Common)+len(c.Repos.Hosts)+len(c.Repos.Users))
	out = append(out, c.Repos.Common...)
	out = append(out, c.Repos.Hosts...)
	out = append(out, c.Repos.Users...)
	return out
}
