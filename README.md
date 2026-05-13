# nixconf

<!-- markdownlint-disable no-empty-links -->

[![Lint Code Base](https://github.com/xx4h/nixconf/actions/workflows/linter-full.yml/badge.svg)](https://github.com/xx4h/nixconf/actions/workflows/linter-full.yml)
[![Test Code Base](https://github.com/xx4h/nixconf/actions/workflows/test-full.yml/badge.svg)](https://github.com/xx4h/nixconf/actions/workflows/test-full.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/xx4h/nixconf?)](https://goreportcard.com/report/github.com/xx4h/nixconf)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)
[![Latest tag](https://img.shields.io/github/v/tag/xx4h/nixconf)](https://github.com/xx4h/nixconf/tags)

<!-- markdownlint-enable no-empty-links -->

Repository manager for a multi-repo NixOS configuration: one shared
`nixos-common`, one `nixos-<host>` per machine, and one `nixos-<user>` per user.
`nixconf` keeps the sub-repos in sync — cloning them, running `nix flake update`
across all of them, verifying that consumers track the latest `nixos-common`,
and forwarding arbitrary `git` commands into each working tree.

## Install

### Pre-built binaries

Grab a tarball for your platform from the
[releases page](https://github.com/xx4h/nixconf/releases) and place the
`nixconf` binary on your `$PATH`.

### Nix flake

```bash
nix run github:xx4h/nixconf
# or, in a flake:
inputs.nixconf.url = "github:xx4h/nixconf";
```

The package installs Bash/Zsh/Fish completion files into the standard
`share/` paths, so shells that auto-load vendor completions pick them up
without further configuration.

### Home Manager

A Home Manager module is exposed at `homeManagerModules.default`:

```nix
{
  inputs.nixconf.url = "github:xx4h/nixconf";
}
```

```nix
# in your Home Manager config
{ inputs, ... }: {
  imports = [ inputs.nixconf.homeManagerModules.default ];

  programs.nixconf = {
    enable = true;
    # All three default to true; toggle off to skip the shell init hook.
    # enableBashIntegration = true;
    # enableZshIntegration  = true;
    # enableFishIntegration = true;
  };
}
```

### Homebrew

```bash
brew install xx4h/tap/nixconf
```

### From source

```bash
git clone https://github.com/xx4h/nixconf.git
cd nixconf
task build               # → ./bin/nixconf
task local-install       # → ~/.local/bin/nixconf
```

## Configuration

`nixconf` reads `nixconf.yaml` from the first match in this order:

1. The path passed to `-c, --config <path>`
2. The first `nixconf.yaml` found walking up from the current directory
3. `$XDG_CONFIG_HOME/nixconf.yaml` (or `~/.config/nixconf.yaml`)

Run `nixconf init` to create a starter file at
`$XDG_CONFIG_HOME/nixconf.yaml`, or `nixconf init <path>` to write it
elsewhere.

The file lists the repos to manage, grouped into `common`, `hosts`, and
`users`:

```yaml
git_base: "git@github.com:<user>"

# Optional. Repos are cloned under this directory. Defaults to
# $XDG_DATA_HOME/nixconf (or ~/.local/share/nixconf). Relative values
# are resolved against the directory holding this file.
# data_dir: ""

repos:
  common:
    - name: nixos-common
      path: nixos-common
  hosts:
    - name: nixos-desktop
      path: hosts/nixos-desktop
    - name: nixos-laptop
      path: hosts/nixos-laptop
    # Per-repo `url` overrides git_base entirely.
    - name: nixos-other
      path: hosts/nixos-other
      url: "https://git.example.com/team/nixos-other.git"
    # Disabled entries are skipped by clone/update/verify/git.
    - name: nixos-archive
      path: hosts/nixos-archive
      disabled: true
  users:
    - name: nixos-<user>
      path: users/nixos-<user>
```

`git_base` is optional if every repository provides its own `url`.

## Commands

| Command                                | What it does                                                              |
| -------------------------------------- | ------------------------------------------------------------------------- |
| `nixconf init [path]`                  | Create a starter `nixconf.yaml` (default: `$XDG_CONFIG_HOME`)             |
| `nixconf clone`                        | Clone every repository in `nixconf.yaml` into its `path` under `data_dir` |
| `nixconf update [INPUT ...]`           | `nix flake update` + commit `flake.lock` + push, in every repository      |
| `nixconf verify`                       | Check host/user `flake.lock`s point at the latest `nixos-common`          |
| `nixconf git <args>`                   | Run `git -C <repo> <args>` in every selected repository                   |
| `nixconf config add <kind> <name>`     | Add a `host` or `user` entry (flags: `--path`, `--url`)                   |
| `nixconf config edit <kind> <name>`    | Edit an entry (flags: `--name`, `--path`, `--url`)                        |
| `nixconf config delete <kind> <name>`  | Remove an entry (aliases: `rm`, `remove`)                                 |
| `nixconf config disable <kind> <name>` | Mark an entry as disabled — skipped by other commands                     |
| `nixconf config enable <kind> <name>`  | Re-enable a previously disabled entry                                     |
| `nixconf version`                      | Print version / commit / build date                                       |
| `nixconf completion <shell>`           | Print shell completion (Bash/Zsh/Fish/PowerShell)                         |

Arbitrary Git commands are run only via the explicit `nixconf git`
subcommand; unknown subcommands are no longer auto-forwarded.

**Selectors** (combine with any command):

| Flag                  | Effect                              |
| --------------------- | ----------------------------------- |
| `--common`            | only common repos                   |
| `--hosts`             | only host config repos              |
| `--users`             | only user config repos              |
| `-r, --repo <name>`   | only the named repository           |
| `-n, --dry-run`       | show what would be done; no changes |
| `-c, --config <path>` | override the `nixconf.yaml` lookup  |

### Examples

```bash
# create a starter config in $XDG_CONFIG_HOME/nixconf.yaml
nixconf init

# add a host, then clone everything
nixconf config add host nixos-desktop
nixconf clone

# bump nixpkgs across all repos, commit + push the lockfile in each
nixconf update nixpkgs

# bump everything, only in host repos
nixconf --hosts update

# check that every host/user flake.lock pins the latest nixos-common
nixconf verify

# git in every repo (use the explicit `git` subcommand)
nixconf git status
nixconf --hosts git pull
nixconf -r nixos-common git log --oneline -5
```

## Shell completion

```bash
# zsh — one-time setup
nixconf completion zsh > "${fpath[1]}/_nixconf"

# bash
nixconf completion bash > /etc/bash_completion.d/nixconf
```

## License

Apache-2.0. See [LICENSE](./LICENSE) and [NOTICE](./NOTICE).
