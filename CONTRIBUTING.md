# Contributing to nixconf

## Development Setup

```bash
git clone https://github.com/xx4h/nixconf.git
cd nixconf

# Using Nix (recommended — pulls in go, golangci-lint, goreleaser, task, etc.)
nix develop
# Or with direnv: `direnv allow` once for automatic shell activation.

task build               # → ./bin/nixconf
task --list              # see all targets
```

## Making Changes

1. Create a feature branch from `main`.
2. Make your changes.
3. Add or update tests as needed.
4. Run the test suite: `task test-unit`.
5. Run the linter: `task test-style`.
6. Commit using [Conventional Commits](https://www.conventionalcommits.org/).
7. Open a pull request against `main`.

## Releases

Releases are cut by pushing a `vX.Y.Z` tag; `.github/workflows/release.yml`
runs `goreleaser` and publishes archives. The `VERSION` file should be bumped
in the same commit as the tag.
