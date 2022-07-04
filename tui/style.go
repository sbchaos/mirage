package tui

import "github.com/charmbracelet/lipgloss"

var (
	Color   = lipgloss.AdaptiveColor{Light: "#111222", Dark: "#FAFAFA"}
	Primary = lipgloss.Color("##1976D2")
	Green   = lipgloss.Color("#1EB980")
	Teal    = lipgloss.Color("#00796B")
	Red     = lipgloss.Color("#D32F2F")
	White   = lipgloss.Color("#ffffff")
	Black   = lipgloss.Color("#000000")
	Orange  = lipgloss.Color("#FF6859")
	Feint   = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#888888"}

	TextStyle  = lipgloss.NewStyle().Foreground(Color)
	FeintStyle = TextStyle.Copy().Foreground(Feint)
	BoldStyle  = TextStyle.Copy().Bold(true)
)

// RenderError returns a formatted error string.
func RenderError(msg string) string {
	// Error applies styles to an error message
	err := lipgloss.NewStyle().Background(Red).Foreground(White).Bold(true).Padding(0, 1).Render("Error")
	content := lipgloss.NewStyle().Bold(true).Padding(0, 1).Render(msg)
	return err + content
}

// RenderWarning returns a formatted warning string.
func RenderWarning(msg string) string {
	// Error applies styles to an error message
	err := lipgloss.NewStyle().Foreground(Orange).Bold(true).Render("Warning: ")
	content := lipgloss.NewStyle().Bold(true).Foreground(Orange).Padding(0, 1).Render(msg)
	return err + content
}
