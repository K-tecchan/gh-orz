# gh-orz

A [GitHub CLI](https://cli.github.com/) extension to bulk clone, update, and manage repositories under an org or user.

Repositories are organized in a [ghq](https://github.com/x-motemen/ghq)-compatible directory structure (`<root>/<host>/<owner>/<repo>`), so gh-orz works alongside ghq seamlessly.

> **Why "orz"?**
> "orz" is a Japanese internet emoticon representing a person kneeling on the ground -- often used to express frustration or exhaustion.
> Managing dozens of repositories across multiple orgs can feel exactly like that.
> gh-orz takes the pain away so you don't have to _orz_ anymore.

## Installation

```sh
gh extension install https://github.com/K-tecchan/gh-orz
```

## Commands

### `gh orz clone [owner]`

Clone repositories under an org or user with interactive multi-select.

```sh
gh orz clone                                 # fuzzy-select an org or user, then interactive selection
gh orz clone my-org                          # interactive selection
gh orz clone my-org --repo=api,web,docs      # non-interactive
gh orz clone my-org --include-archived       # include archived repos
```

Features:
- If `owner` is omitted, a fuzzy finder lists the orgs you belong to ( org) and your own account ( user), plus any owner you've already cloned repos from, tagged with `[N cloned]`
- Interactive multi-select UI with filter search (`/` key)
- Private and fork repos are tagged with [Nerd Font](https://www.nerdfonts.com/) icons ( private,  fork); set `gh-orz.iconDisabled` to `true` to fall back to `[private]`/`[fork]`/`[org]`/`[user]` text tags
- Already cloned repos are dimmed and non-selectable
- SSH/HTTPS based on `gh config get git_protocol`

### `gh orz pull <owner>`

Bulk update all cloned repositories under an owner.

```sh
gh orz pull my-org                  # pull default branch for all repos
gh orz pull my-org --current-branch # pull the current branch as-is
```

Features:
- Checks out and pulls the default branch by default
- Uses `--ff-only` to avoid merge conflicts
- Skips repos with uncommitted changes (no data loss)
- Parallel execution for speed

### `gh orz list [owner]`

List cloned repositories.

```sh
gh orz list                    # all repos
gh orz list my-org             # repos under a specific owner
gh orz list --full-path        # output full paths (useful for shell integration)
```

Shell integration example:

```sh
cd "$(gh orz list --full-path | fzf)"
```

### `gh orz rm <owner>`

Remove cloned repositories with interactive selection.

```sh
gh orz rm my-org                       # interactive selection
gh orz rm my-org --repo=old,unused     # non-interactive
```

Automatically cleans up empty owner directories after removal.

### `gh orz exec <owner> -- <command> [args...]`

Run a command in all cloned repositories under an owner.

```sh
gh orz exec my-org -- git gc                # run git gc in all repos
gh orz exec my-org -- gh pr list            # list PRs across all repos
gh orz exec my-org -- git status --short    # check status of all repos
```

Features:
- Parallel execution for speed
- Output is grouped by repository

### `gh orz status [owner]`

Show a health summary of all cloned repositories.

```sh
gh orz status              # all repos
gh orz status my-org       # specific owner
```

Reports:
- Uncommitted changes
- Unpushed commits
- Commits ahead of the default branch

## GitHub Enterprise

gh-orz automatically detects authenticated hosts from `gh auth`. You can also specify a host explicitly:

```sh
gh orz clone my-org --host=ghe.example.com
```

## Root directory

gh-orz resolves the root directory in this order:

1. `gh-orz.root` in `.gitconfig`
2. `ghq.root` in `.gitconfig`
3. `~/gh-orz` (default)

## Icons

Repo tags (private/fork) and owner tags (org/user) use [Nerd Font](https://www.nerdfonts.com/) icons by default. If your terminal font isn't a Nerd Font, disable them to fall back to plain text tags:

```sh
git config --global gh-orz.iconDisabled true
```

## Acknowledgments

gh-orz is heavily inspired by [ghq](https://github.com/x-motemen/ghq) by [motemen](https://github.com/motemen). The ghq-compatible directory structure and root directory resolution are direct tributes to ghq's excellent design. Thank you for creating such a great tool.

## License

MIT
