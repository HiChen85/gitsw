# gitsw — Git Identity Switcher

A TUI tool for managing multiple Git user identities across GitHub/GitLab accounts, with pre-push hook confirmation to prevent accidental pushes with the wrong identity.

## Problem

Developers working across multiple organizations (personal projects, client work, open source) frequently forget to switch their local git identity before pushing. This leads to commits attributed to the wrong email/account, which is difficult to fix after the fact.

## Solution

A single Go binary (`gitsw`) that:
1. Installs as a git pre-push hook to confirm identity before every push
2. Provides a full TUI for managing and switching between saved profiles
3. Stores profiles in a dedicated config file, switches via `git config --local`

## Architecture

### Single binary, dual mode

- **Hook mode** (`gitsw hook`) — lightweight Y/n prompt, no Bubble Tea dependency in this path
- **TUI mode** (`gitsw`) — full Bubble Tea interactive interface

### Commands

```
gitsw              Launch interactive TUI
gitsw hook         Pre-push hook mode (used by git hooks)
gitsw install      Install pre-push hook to current repo
gitsw install -g   Install pre-push hook globally
gitsw uninstall    Remove pre-push hook from current repo
gitsw uninstall -g Remove global pre-push hook
gitsw list         List all configured profiles
gitsw help         Show help message
```

Flags: `-h/--help`, `-v/--version`

### Project structure

```
cmd/gitsw/main.go          — entrypoint, command routing
internal/config/            — profile CRUD, YAML I/O
internal/git/               — git config read/write operations
internal/hook/              — hook mode logic (compact prompt)
internal/tui/               — Bubble Tea TUI (models, update, view)
internal/tui/views/         — individual TUI screens
```

## Data Model

### Profile storage

Location: `~/.gitswitch/profiles.yaml`

```yaml
profiles:
  - nickname: "work"
    name: "Zhang Haichen"
    email: "haichen@company.com"
    platform: "gitlab"
  - nickname: "personal"
    name: "Zhang Haichen"
    email: "haichen@gmail.com"
    platform: "github"
  - nickname: "oss"
    name: "HaiChen Z"
    email: "hc@opensource.dev"
    platform: "github"
```

Fields:
- `nickname` (required, unique) — short identifier for display and selection
- `name` (required) — maps to `user.name`
- `email` (required) — maps to `user.email`
- `platform` (required) — one of `github`, `gitlab`, `other`

## Hook Mode

### Display (compact, 3-line info box)

```
╭─ gitsw: Identity Check ──────────────────────────────────╮
│ Repo:    ~/code/client-project → gitlab.company.com:team/project.git
│ User:    Zhang Haichen <haichen@company.com> (local)
│ Profile: work (gitlab)
╰──────────────────────────────────────────────────────────╯
Push as this identity? [Y/n]
```

### Behavior

- **Y or Enter** — exit 0, push proceeds
- **n** — exit 1, push aborted, prints: `✗ Push aborted. Run gitsw to switch identity, then push again.`

### Warning states

- No local config (global fallback): yellow/red border, shows `(global ⚠)` and message "No local config — using global fallback"
- Identity matches no saved profile: shows "unrecognized" instead of profile name

### Special cases

- Not a TTY (CI/CD): skip confirmation, exit 0
- No profiles configured: skip confirmation, exit 0
- Binary not in PATH: hook script prints helpful error

## TUI Mode

### Main dashboard

Shows:
- Current repo path and active identity (highlighted)
- List of all profiles with active one marked

Keybindings:
- `↑/↓` or `j/k` — navigate profiles
- `Enter` — switch current repo to selected profile
- `a` — add new profile
- `e` — edit selected profile
- `d` — delete (with confirmation)
- `i` — install/uninstall hook
- `q/Esc` — quit

### Add/Edit form

Tab-navigable input fields:
- Nickname (text input)
- Name (text input)
- Email (text input)
- Platform (horizontal selector: github / gitlab / other)

Keybindings: `Tab/Shift+Tab` navigate, `Enter` save, `Esc` cancel.

### Switching mechanism

Executes:
```
git config --local user.name "<name>"
git config --local user.email "<email>"
```

Only affects the current repo's `.git/config`. Other repos are untouched.

## Hook Installation

### Per-repo (`gitsw install`)

Writes to `.git/hooks/pre-push`:
```bash
#!/bin/sh
exec gitsw hook "$@"
```

### Global (`gitsw install -g`)

- Sets `core.hooksPath` to `~/.gitswitch/hooks/` (in `~/.gitconfig`)
- Creates `~/.gitswitch/hooks/pre-push` with the same script
- Note: when `core.hooksPath` is set, git ignores repo-level `.git/hooks/`. If a repo needs additional hooks, they must be placed in `~/.gitswitch/hooks/` or chained manually.
- If `core.hooksPath` is already set to a non-gitsw path, abort and warn the user rather than overwriting.

### Uninstall

- Per-repo: removes `.git/hooks/pre-push` (only if it contains `gitsw`)
- Global: unsets `core.hooksPath` only if it points to `~/.gitswitch/hooks/`, then removes that directory

## Error Handling

- Config file missing → create fresh with empty profiles list
- Config file corrupt → back up to `.yaml.bak`, create fresh, warn user
- Not in a git repo (for hook/install commands) → error message, exit 1
- Permission denied writing hook → suggest fix, exit 1
- Hook outside TTY → skip confirmation, allow push

## Dependencies

- Go 1.22+
- `charmbracelet/bubbletea` — TUI framework
- `charmbracelet/lipgloss` — TUI styling
- `charmbracelet/bubbles` — input and list components
- `gopkg.in/yaml.v3` — config serialization

## Testing Strategy

- Unit tests: `internal/config/` (profile CRUD, YAML round-trip), `internal/git/` (command generation)
- Integration tests: hook mode (simulate input, verify exit codes)
- Manual testing: TUI interactions
