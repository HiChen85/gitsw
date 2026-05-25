# gitsw

一个用于管理多个 Git 用户身份的 TUI 工具，支持 GitHub/GitLab 账号切换，并通过 pre-push hook 在每次推送前确认身份，防止使用错误账号推送代码。

[English](./README.md)

## 功能特性

- **推送前身份确认** — 通过 git hook 在每次 push 前确认当前的 user/email 配置
- **快速切换** — 在已保存的配置之间即时切换本地 git 身份
- **配置管理** — 通过交互式 TUI 界面添加、编辑、删除身份配置
- **本地隔离** — 使用 `git config --local`，确保各仓库之间互不影响
- **一键安装 Hook** — 一条命令为当前仓库或全局安装 pre-push hook

## 安装

### 通过 `go install`

```bash
go install github.com/HiChen85/gitsw/cmd/gitsw@latest
```

确保 `$GOPATH/bin`（通常是 `~/go/bin`）在你的 `PATH` 中。

### 从源码构建

```bash
git clone https://github.com/HiChen85/gitsw.git
cd gitsw
go build -o gitsw ./cmd/gitsw
```

## 快速开始

```bash
# 启动 TUI 管理配置
gitsw

# 为当前仓库安装 pre-push hook
gitsw install

# 全局安装 pre-push hook（所有仓库生效）
gitsw install -g
```

## 命令说明

```
gitsw              启动交互式 TUI 界面
gitsw hook         Pre-push hook 模式（由 git hook 调用）
gitsw install      为当前仓库安装 pre-push hook
gitsw install -g   全局安装 pre-push hook
gitsw uninstall    移除当前仓库的 pre-push hook
gitsw uninstall -g 移除全局 pre-push hook
gitsw list         列出所有已配置的身份
gitsw help         显示帮助信息
```

### TUI 快捷键

| 按键 | 操作 |
|------|------|
| `j/k` 或 `↑/↓` | 上下移动选择 |
| `Enter` | 切换到选中的身份 |
| `a` | 添加新身份 |
| `e` | 编辑选中的身份 |
| `d` | 删除选中的身份 |
| `i` | 安装 pre-push hook |
| `q` / `Esc` | 退出 |

### Pre-push Hook 效果

当你执行 `git push` 时，hook 会显示：

```
╭─ gitsw: Identity Check ──────────────────────────────────╮
│ Repo:    ~/code/project → git@github.com:user/repo.git
│ User:    Your Name <your@email.com> (local)
│ Profile: work (gitlab)
╰──────────────────────────────────────────────────────────╯
Push as this identity? [Y/n]
```

- 按 `Y` 或 `Enter` 继续推送
- 按 `n` 中止推送，然后运行 `gitsw` 切换身份

## 配置文件

身份配置存储在 `~/.gitswitch/profiles.yaml`：

```yaml
profiles:
  - nickname: "work"
    name: "你的名字"
    email: "you@company.com"
    platform: "gitlab"
  - nickname: "personal"
    name: "你的名字"
    email: "you@gmail.com"
    platform: "github"
```

## 许可证

MIT
