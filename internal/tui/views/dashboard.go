package views

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haichen-zhang/gitsw/internal/config"
	"github.com/haichen-zhang/gitsw/internal/tui/styles"
)

// Dashboard is the main view showing the profile list.
type Dashboard struct {
	cfg          *config.Config
	repoDir      string
	cursor       int
	currentEmail string
	status       string
}

// NewDashboard creates a new dashboard view.
func NewDashboard(cfg *config.Config, repoDir, currentEmail string) Dashboard {
	return Dashboard{
		cfg:          cfg,
		repoDir:      repoDir,
		cursor:       0,
		currentEmail: currentEmail,
		status:       "",
	}
}

// Init implements tea.Model.
func (d Dashboard) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (d Dashboard) Update(msg tea.Msg) (Dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
		case "down", "j":
			if d.cursor < len(d.cfg.Profiles)-1 {
				d.cursor++
			}
		}
	}
	return d, nil
}

// View implements tea.Model.
func (d Dashboard) View() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("gitsw — Git Identity Switcher"))
	b.WriteString("\n\n")

	// Repo info
	if d.repoDir != "" {
		b.WriteString(styles.DimStyle.Render("Repo: " + shortenHome(d.repoDir)))
		b.WriteString("\n")
		if d.currentEmail != "" {
			b.WriteString(styles.DimStyle.Render("Active: " + d.currentEmail))
		} else {
			b.WriteString(styles.DimStyle.Render("Active: (none)"))
		}
		b.WriteString("\n")
	} else {
		b.WriteString(styles.DimStyle.Render("Not in a git repository"))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Profile list
	if len(d.cfg.Profiles) == 0 {
		b.WriteString(styles.DimStyle.Render("  No profiles configured. Press [a] to add one."))
		b.WriteString("\n")
	} else {
		for i, p := range d.cfg.Profiles {
			// Build the line
			marker := "  "
			if p.Email == d.currentEmail {
				marker = styles.ActiveStyle.Render("● ")
			}

			// Platform coloring
			platform := p.Platform
			if platform == "gitlab" {
				platform = lipgloss.NewStyle().Foreground(styles.ColorYellow).Render(platform)
			}

			line := fmt.Sprintf("%s%-12s %s <%s>  (%s)", marker, p.Nickname, p.Name, p.Email, platform)

			// Active marker
			if p.Email == d.currentEmail {
				line += styles.ActiveStyle.Render("  active")
			}

			// Cursor highlight
			if i == d.cursor {
				line = styles.SelectedStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Status message
	if d.status != "" {
		b.WriteString(d.status)
		b.WriteString("\n\n")
	}

	// Help bar
	help := "[enter] switch  [a] add  [e] edit  [d] delete  [i] install hook  [q] quit"
	b.WriteString(styles.HelpStyle.Render(help))

	return b.String()
}

// SelectedProfile returns the currently selected profile, or nil if none.
func (d *Dashboard) SelectedProfile() *config.Profile {
	if len(d.cfg.Profiles) == 0 {
		return nil
	}
	if d.cursor < 0 || d.cursor >= len(d.cfg.Profiles) {
		return nil
	}
	return &d.cfg.Profiles[d.cursor]
}

// SetStatus sets the status message.
func (d *Dashboard) SetStatus(msg string) {
	d.status = msg
}

// SetCurrentEmail updates the current email shown in the dashboard.
func (d *Dashboard) SetCurrentEmail(email string) {
	d.currentEmail = email
}

// SetConfig updates the config reference and adjusts cursor if needed.
func (d *Dashboard) SetConfig(cfg *config.Config) {
	d.cfg = cfg
	if d.cursor >= len(cfg.Profiles) {
		d.cursor = len(cfg.Profiles) - 1
	}
	if d.cursor < 0 {
		d.cursor = 0
	}
}

// shortenHome replaces the home directory prefix with ~.
func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
