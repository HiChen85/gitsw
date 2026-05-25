package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haichen-zhang/gitsw/internal/config"
	"github.com/haichen-zhang/gitsw/internal/git"
	"github.com/haichen-zhang/gitsw/internal/hook"
	"github.com/haichen-zhang/gitsw/internal/tui/styles"
	"github.com/haichen-zhang/gitsw/internal/tui/views"
)

type screen int

const (
	screenDashboard screen = iota
	screenForm
	screenConfirm
)

// appModel is the root model that routes between screens.
type appModel struct {
	cfg       *config.Config
	cfgPath   string
	repoDir   string
	screen    screen
	dashboard views.Dashboard
	form      views.Form
	confirm   views.Confirm
}

func newAppModel(cfg *config.Config, cfgPath, repoDir string) appModel {
	currentEmail := ""
	if repoDir != "" {
		identity, _ := git.GetIdentity(repoDir)
		currentEmail = identity.Email
	}

	return appModel{
		cfg:       cfg,
		cfgPath:   cfgPath,
		repoDir:   repoDir,
		screen:    screenDashboard,
		dashboard: views.NewDashboard(cfg, repoDir, currentEmail),
	}
}

// Init implements tea.Model.
func (m appModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global key handling
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case screenDashboard:
		return m.updateDashboard(msg)
	case screenForm:
		return m.updateForm(msg)
	case screenConfirm:
		return m.updateConfirm(msg)
	}

	return m, nil
}

func (m appModel) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "a":
			m.form = views.NewForm(nil)
			m.screen = screenForm
			return m, m.form.Init()
		case "e":
			p := m.dashboard.SelectedProfile()
			if p != nil {
				m.form = views.NewForm(p)
				m.screen = screenForm
				return m, m.form.Init()
			}
		case "d":
			p := m.dashboard.SelectedProfile()
			if p != nil {
				m.confirm = views.NewConfirm(p.Nickname)
				m.screen = screenConfirm
				return m, nil
			}
		case "enter":
			p := m.dashboard.SelectedProfile()
			if p != nil && m.repoDir != "" {
				if err := git.SetIdentity(m.repoDir, p.Name, p.Email); err != nil {
					m.dashboard.SetStatus(styles.ErrorStyle.Render("Error: " + err.Error()))
				} else {
					m.dashboard.SetCurrentEmail(p.Email)
					m.dashboard.SetStatus(styles.ActiveStyle.Render(fmt.Sprintf("Switched to %s <%s>", p.Name, p.Email)))
				}
				return m, nil
			} else if m.repoDir == "" {
				m.dashboard.SetStatus(styles.WarningStyle.Render("Not in a git repository"))
				return m, nil
			}
		case "i":
			if m.repoDir != "" {
				if err := hook.InstallLocal(m.repoDir); err != nil {
					m.dashboard.SetStatus(styles.ErrorStyle.Render("Hook error: " + err.Error()))
				} else {
					m.dashboard.SetStatus(styles.ActiveStyle.Render("Hook installed"))
				}
			} else {
				m.dashboard.SetStatus(styles.WarningStyle.Render("Not in a git repository"))
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	return m, cmd
}

func (m appModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case views.FormSubmitMsg:
		submitMsg := msg.(views.FormSubmitMsg)
		if submitMsg.Editing != "" {
			// Update existing profile
			if err := m.cfg.Update(submitMsg.Editing, submitMsg.Profile); err != nil {
				m.dashboard.SetStatus(styles.ErrorStyle.Render("Error: " + err.Error()))
			} else {
				_ = m.cfg.SaveTo(m.cfgPath)
				m.dashboard.SetConfig(m.cfg)
				m.dashboard.SetStatus(styles.ActiveStyle.Render(fmt.Sprintf("Updated profile %q", submitMsg.Profile.Nickname)))
			}
		} else {
			// Add new profile
			if err := m.cfg.Add(submitMsg.Profile); err != nil {
				m.dashboard.SetStatus(styles.ErrorStyle.Render("Error: " + err.Error()))
			} else {
				_ = m.cfg.SaveTo(m.cfgPath)
				m.dashboard.SetConfig(m.cfg)
				m.dashboard.SetStatus(styles.ActiveStyle.Render(fmt.Sprintf("Added profile %q", submitMsg.Profile.Nickname)))
			}
		}
		m.screen = screenDashboard
		return m, nil

	case views.FormCancelMsg:
		m.screen = screenDashboard
		return m, nil
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m appModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case views.ConfirmYesMsg:
		yesMsg := msg.(views.ConfirmYesMsg)
		if err := m.cfg.Delete(yesMsg.Nickname); err != nil {
			m.dashboard.SetStatus(styles.ErrorStyle.Render("Error: " + err.Error()))
		} else {
			_ = m.cfg.SaveTo(m.cfgPath)
			m.dashboard.SetConfig(m.cfg)
			m.dashboard.SetStatus(styles.ActiveStyle.Render(fmt.Sprintf("Deleted profile %q", yesMsg.Nickname)))
		}
		m.screen = screenDashboard
		return m, nil

	case views.ConfirmNoMsg:
		m.screen = screenDashboard
		return m, nil
	}

	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m appModel) View() string {
	switch m.screen {
	case screenForm:
		return m.form.View()
	case screenConfirm:
		return m.confirm.View()
	default:
		return m.dashboard.View()
	}
}

// Run starts the TUI application.
func Run(cfg *config.Config, cfgPath, repoDir string) error {
	model := newAppModel(cfg, cfgPath, repoDir)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
