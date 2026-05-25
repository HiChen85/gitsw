package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorCyan   = lipgloss.Color("#64ffda")
	ColorPurple = lipgloss.Color("#bb86fc")
	ColorYellow = lipgloss.Color("#ffd93d")
	ColorRed    = lipgloss.Color("#ff6b6b")
	ColorDim    = lipgloss.Color("#888888")
	ColorWhite  = lipgloss.Color("#ffffff")

	TitleStyle    = lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
	ActiveStyle   = lipgloss.NewStyle().Foreground(ColorCyan)
	DimStyle      = lipgloss.NewStyle().Foreground(ColorDim)
	WarningStyle  = lipgloss.NewStyle().Foreground(ColorYellow)
	ErrorStyle    = lipgloss.NewStyle().Foreground(ColorRed)
	SelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#2a2a4a")).Padding(0, 1)
	BoxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorCyan).Padding(1, 2)
	HelpStyle     = lipgloss.NewStyle().Foreground(ColorDim)
)
