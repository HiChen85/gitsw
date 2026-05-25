package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/HiChen85/gitsw/internal/tui/styles"
)

// ConfirmYesMsg is sent when the user confirms deletion.
type ConfirmYesMsg struct {
	Nickname string
}

// ConfirmNoMsg is sent when the user cancels deletion.
type ConfirmNoMsg struct{}

// Confirm is a yes/no confirmation dialog.
type Confirm struct {
	nickname string
	cursor   int // 0 = yes, 1 = no
}

// NewConfirm creates a new confirmation dialog for deleting the given nickname.
func NewConfirm(nickname string) Confirm {
	return Confirm{
		nickname: nickname,
		cursor:   1, // default to "No"
	}
}

// Init implements tea.Model.
func (c Confirm) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			c.cursor = 0
		case "right", "l":
			c.cursor = 1
		case "y":
			return c, func() tea.Msg {
				return ConfirmYesMsg{Nickname: c.nickname}
			}
		case "n", "esc":
			return c, func() tea.Msg {
				return ConfirmNoMsg{}
			}
		case "enter":
			if c.cursor == 0 {
				return c, func() tea.Msg {
					return ConfirmYesMsg{Nickname: c.nickname}
				}
			}
			return c, func() tea.Msg {
				return ConfirmNoMsg{}
			}
		}
	}
	return c, nil
}

// View implements tea.Model.
func (c Confirm) View() string {
	var b strings.Builder

	b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("  Delete profile %q?", c.nickname)))
	b.WriteString("\n\n  ")

	yes := "  Yes, delete  "
	no := "  No, cancel  "

	if c.cursor == 0 {
		b.WriteString(styles.ErrorStyle.Render("[" + yes + "]"))
		b.WriteString("    ")
		b.WriteString(styles.DimStyle.Render(" " + no + " "))
	} else {
		b.WriteString(styles.DimStyle.Render(" " + yes + " "))
		b.WriteString("    ")
		b.WriteString(styles.ActiveStyle.Render("[" + no + "]"))
	}

	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("  [←/→] select  [enter] confirm  [y] yes  [n/esc] cancel"))

	return styles.BoxStyle.Render(b.String())
}
