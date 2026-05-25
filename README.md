# gitsw

A TUI tool for managing multiple Git user identities across GitHub/GitLab accounts, with pre-push hook confirmation to prevent accidental pushes with the wrong identity.

[中文文档](./README_zh.md)

## Features

- **Pre-push identity check** — Confirms your git user/email before every push via a git hook
- **Quick switch** — Switch local git config between saved profiles instantly
- **Profile management** — Add, edit, and delete identity profiles through an interactive TUI
- **Local isolation** — Uses `git config --local` so each repo has its own identity
- **Hook install** — One command to install the pre-push hook (per-repo or global)

## Installation

### Via `go install`

```bash
go install github.com/HiChen85/gitsw/cmd/gitsw@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`.

### Build from source

```bash
git clone https://github.com/HiChen85/gitsw.git
cd gitsw
go build -o gitsw ./cmd/gitsw
```

## Quick Start

```bash
# Launch TUI to manage profiles
gitsw

# Install pre-push hook for current repo
gitsw install

# Install pre-push hook globally (all repos)
gitsw install -g
```

## Usage

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

### TUI Keybindings

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate profiles |
| `Enter` | Switch to selected profile |
| `a` | Add new profile |
| `e` | Edit selected profile |
| `d` | Delete selected profile |
| `i` | Install pre-push hook |
| `q` / `Esc` | Quit |

### Pre-push Hook

When you run `git push`, the hook shows:

```
╭─ gitsw: Identity Check ──────────────────────────────────╮
│ Repo:    ~/code/project → git@github.com:user/repo.git
│ User:    Your Name <your@email.com> (local)
│ Profile: work (gitlab)
╰──────────────────────────────────────────────────────────╯
Push as this identity? [Y/n]
```

- Press `Y` or `Enter` to proceed
- Press `n` to abort and switch identity with `gitsw`

## Configuration

Profiles are stored at `~/.gitswitch/profiles.yaml`:

```yaml
profiles:
  - nickname: "work"
    name: "Your Name"
    email: "you@company.com"
    platform: "gitlab"
  - nickname: "personal"
    name: "Your Name"
    email: "you@gmail.com"
    platform: "github"
```

## License

MIT
