# AGENTS.md

This file provides context for AI coding agents working on this project.

## Project overview

gh-orz is a GitHub CLI extension written in Go. It bulk clones, updates, and manages repositories under GitHub orgs/users in a ghq-compatible directory structure.

## Architecture

```
main.go                    # Entry point, calls cmd.Execute()
cmd/
  root.go                  # Cobra root command, --host flag
  clone.go                 # gh orz clone <owner>
  pull.go                  # gh orz pull <owner>
  list.go                  # gh orz list [owner]
  rm.go                    # gh orz rm <owner>
  exec.go                  # gh orz exec <owner> -- <command>
  status.go                # gh orz status [owner]
internal/
  config/root.go           # Root directory resolution (gh-orz.root > ghq.root > ~/gh-orz), icon toggle (gh-orz.iconDisabled)
  github/repos.go          # GitHub API client (repo listing with pagination, org/user/auth-user fallback)
  git/clone.go             # Git clone operations
  git/pull.go              # Git pull with dirty check, default branch detection
  git/exec.go              # Arbitrary command execution across repos
  git/status.go            # Git status inspection (dirty, unpushed, ahead of default)
  ui/select.go             # Interactive multi-select UI (bubbletea)
  ui/color.go              # Terminal color helpers (termenv)
```

## Key dependencies

- `github.com/cli/go-gh/v2` — GitHub CLI integration (API client, auth, host detection)
- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/bubbletea` — Terminal UI
- `github.com/muesli/termenv` — Terminal color output

## Coding conventions

- Go standard project layout with `cmd/` and `internal/`
- Each subcommand is a separate file in `cmd/`
- Business logic lives in `internal/`, not in `cmd/`
- Use `go-gh` REST client for GitHub API calls (not raw HTTP)
- Use `exec.Command` for git operations (not a git library)
- Parallel operations use `sync.WaitGroup` with goroutines

## Testing

```sh
go test ./...
```

- Tests use `testing.T` with `t.TempDir()` for filesystem tests
- GitHub API tests use `httptest.Server` with a custom `RoundTripper` to redirect requests
- Git operation tests create real git repos in temp directories
- Interactive UI (`internal/ui`) is not unit tested

## Pre-commit hooks

This project uses [lefthook](https://github.com/evilmartians/lefthook) for pre-commit hooks (gofmt, go vet, go test).

```sh
go install github.com/evilmartians/lefthook@latest
lefthook install
```

## Build and run

```sh
go build -o gh-orz .
gh orz --help
```

## Root directory resolution

Priority order:
1. `git config --global gh-orz.root`
2. `git config --global ghq.root`
3. `~/gh-orz`

## Icons

Repo tags (private/fork) in the `gh orz clone` selector use Nerd Font glyphs by default (`internal/ui/select.go`). `git config --global gh-orz.iconDisabled true` falls back to plain text tags (`[private]`, `[fork]`) via `config.IconsDisabled()`.

## GHE support

The `--host` flag defaults to the authenticated host from `gh auth`. The `go-gh` library handles API routing for GitHub Enterprise automatically via `api.NewRESTClient(api.ClientOptions{Host: host})`.
