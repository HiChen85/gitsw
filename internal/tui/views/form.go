package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/haichen-zhang/gitsw/internal/config"
	"github.com/haichen-zhang/gitsw/internal/tui/styles"
)

// FormSubmitMsg is sent when the form is submitted.
type FormSubmitMsg struct {
	Profile config.Profile
	Editing string // empty if adding new, otherwise the original nickname
}

// FormCancelMsg is sent when the form is cancelled.
type FormCancelMsg struct{}

// Form is the add/edit profile form view.
type Form struct {
	inputs     []textinput.Model
	focusIndex int
	platforms  []string
	platIndex  int
	editing    string // original nickname when editing, empty for add
}

const (
	formFieldNickname = 0
	formFieldName     = 1
	formFieldEmail    = 2
	formFieldPlatform = 3 // virtual field for platform selector
)

// NewForm creates a new form. If existing is non-nil, the form is pre-filled for editing.
func NewForm(existing *config.Profile) Form {
	platforms := []string{"github", "gitlab", "other"}

	inputs := make([]textinput.Model, 3)

	inputs[formFieldNickname] = textinput.New()
	inputs[formFieldNickname].Placeholder = "e.g. work"
	inputs[formFieldNickname].CharLimit = 30
	inputs[formFieldNickname].Width = 30
	inputs[formFieldNickname].Prompt = ""

	inputs[formFieldName] = textinput.New()
	inputs[formFieldName].Placeholder = "e.g. John Doe"
	inputs[formFieldName].CharLimit = 60
	inputs[formFieldName].Width = 40
	inputs[formFieldName].Prompt = ""

	inputs[formFieldEmail] = textinput.New()
	inputs[formFieldEmail].Placeholder = "e.g. john@example.com"
	inputs[formFieldEmail].CharLimit = 60
	inputs[formFieldEmail].Width = 40
	inputs[formFieldEmail].Prompt = ""

	var editing string
	platIndex := 0

	if existing != nil {
		inputs[formFieldNickname].SetValue(existing.Nickname)
		inputs[formFieldName].SetValue(existing.Name)
		inputs[formFieldEmail].SetValue(existing.Email)
		editing = existing.Nickname
		for i, p := range platforms {
			if p == existing.Platform {
				platIndex = i
				break
			}
		}
	}

	// Focus first field
	inputs[formFieldNickname].Focus()

	return Form{
		inputs:     inputs,
		focusIndex: 0,
		platforms:  platforms,
		platIndex:  platIndex,
		editing:    editing,
	}
}

// Init implements tea.Model.
func (f Form) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return FormCancelMsg{} }
		case "tab", "down":
			f.focusIndex++
			if f.focusIndex > formFieldPlatform {
				f.focusIndex = 0
			}
			return f, f.updateFocus()
		case "shift+tab", "up":
			f.focusIndex--
			if f.focusIndex < 0 {
				f.focusIndex = formFieldPlatform
			}
			return f, f.updateFocus()
		case "left":
			if f.focusIndex == formFieldPlatform {
				f.platIndex--
				if f.platIndex < 0 {
					f.platIndex = len(f.platforms) - 1
				}
				return f, nil
			}
		case "right":
			if f.focusIndex == formFieldPlatform {
				f.platIndex++
				if f.platIndex >= len(f.platforms) {
					f.platIndex = 0
				}
				return f, nil
			}
		case "enter":
			// If on platform field or all fields filled, submit
			if f.focusIndex == formFieldPlatform {
				return f, f.submit()
			}
			// Move to next field on enter
			f.focusIndex++
			if f.focusIndex > formFieldPlatform {
				return f, f.submit()
			}
			return f, f.updateFocus()
		}
	}

	// Update the focused text input
	if f.focusIndex < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
		return f, cmd
	}

	return f, nil
}

func (f Form) submit() tea.Cmd {
	nickname := strings.TrimSpace(f.inputs[formFieldNickname].Value())
	name := strings.TrimSpace(f.inputs[formFieldName].Value())
	email := strings.TrimSpace(f.inputs[formFieldEmail].Value())

	if nickname == "" || name == "" || email == "" {
		return nil // Don't submit incomplete form
	}

	profile := config.Profile{
		Nickname: nickname,
		Name:     name,
		Email:    email,
		Platform: f.platforms[f.platIndex],
	}

	return func() tea.Msg {
		return FormSubmitMsg{
			Profile: profile,
			Editing: f.editing,
		}
	}
}

func (f *Form) updateFocus() tea.Cmd {
	var cmds []tea.Cmd
	for i := range f.inputs {
		if i == f.focusIndex {
			cmds = append(cmds, f.inputs[i].Focus())
		} else {
			f.inputs[i].Blur()
		}
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// View implements tea.Model.
func (f Form) View() string {
	var b strings.Builder

	title := "Add Profile"
	if f.editing != "" {
		title = "Edit Profile"
	}
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n\n")

	// Fields
	labels := []string{"Nickname", "Name", "Email"}
	for i, label := range labels {
		style := styles.DimStyle
		if i == f.focusIndex {
			style = styles.ActiveStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("  %s:", label)))
		b.WriteString("\n")
		b.WriteString("  " + f.inputs[i].View())
		b.WriteString("\n\n")
	}

	// Platform selector
	style := styles.DimStyle
	if f.focusIndex == formFieldPlatform {
		style = styles.ActiveStyle
	}
	b.WriteString(style.Render("  Platform:"))
	b.WriteString("\n  ")
	for i, p := range f.platforms {
		if i == f.platIndex {
			b.WriteString(styles.ActiveStyle.Render("[" + p + "]"))
		} else {
			b.WriteString(styles.DimStyle.Render(" " + p + " "))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Render("  [tab] next field  [shift+tab] prev  [enter] submit  [esc] cancel"))

	return styles.BoxStyle.Render(b.String())
}
