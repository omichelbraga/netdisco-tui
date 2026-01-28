package tui

import "github.com/charmbracelet/lipgloss"

// Legacy color exports for backward compatibility
var (
	ColorPrimary   = CurrentTheme.Primary
	ColorSecondary = CurrentTheme.Secondary
	ColorSuccess   = CurrentTheme.Success
	ColorDanger    = CurrentTheme.Danger
	ColorWarning   = CurrentTheme.Warning
	ColorText      = CurrentTheme.Text
	ColorTextDim   = CurrentTheme.TextDim
	ColorTextMuted = CurrentTheme.TextMuted
	ColorBorder    = CurrentTheme.Border
	ColorBg        = CurrentTheme.Background
	ColorBgActive  = CurrentTheme.BackgroundActive
)

// UpdateStylesForTheme refreshes all styles when theme changes
func UpdateStylesForTheme() {
	ColorPrimary = CurrentTheme.Primary
	ColorSecondary = CurrentTheme.Secondary
	ColorSuccess = CurrentTheme.Success
	ColorDanger = CurrentTheme.Danger
	ColorWarning = CurrentTheme.Warning
	ColorText = CurrentTheme.Text
	ColorTextDim = CurrentTheme.TextDim
	ColorTextMuted = CurrentTheme.TextMuted
	ColorBorder = CurrentTheme.Border
	ColorBg = CurrentTheme.Background
	ColorBgActive = CurrentTheme.BackgroundActive

	// Refresh all styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary).
		Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.TextMuted).
		Italic(true)

	TabActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary).
		Background(CurrentTheme.BackgroundActive).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.BorderActive).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderTop(true).
		Padding(0, 2).
		MarginRight(1)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.TextMuted).
		Background(CurrentTheme.BackgroundAlt).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.BorderMuted).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderTop(true).
		Padding(0, 2).
		MarginRight(1)

	SearchStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Primary).
		Foreground(CurrentTheme.Text).
		Background(CurrentTheme.BackgroundAlt).
		Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.Danger).
		Bold(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Danger).
		Padding(1, 2).
		MarginTop(1)

	WarningStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.Warning).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Warning).
		Padding(1, 2).
		MarginTop(1)

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.Primary).
		Bold(true)

	ActiveRowStyle = lipgloss.NewStyle().
		Background(CurrentTheme.BackgroundActive).
		Foreground(CurrentTheme.Text).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(CurrentTheme.Primary).
		Padding(0, 1).
		Bold(true)

	NormalRowStyle = lipgloss.NewStyle().
		Padding(0, 1)

	DetailKeyStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.Secondary).
		Width(16).
		Bold(true).
		Align(lipgloss.Right).
		MarginRight(1)

	DetailValStyle = lipgloss.NewStyle().
		Foreground(CurrentTheme.Text)

	StatusUp = lipgloss.NewStyle().
		Foreground(CurrentTheme.Success).
		Bold(true).
		Render("● UP")

	StatusDown = lipgloss.NewStyle().
		Foreground(CurrentTheme.Danger).
		Bold(true).
		Render("● DOWN")
}

// Initialize styles on package load
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(CurrentTheme.Primary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.TextMuted).
			Italic(true)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(CurrentTheme.Primary).
			Background(CurrentTheme.BackgroundActive).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.BorderActive).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true).
			BorderTop(true).
			Padding(0, 2).
			MarginRight(1)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(CurrentTheme.TextMuted).
				Background(CurrentTheme.BackgroundAlt).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(CurrentTheme.BorderMuted).
				BorderBottom(true).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				Padding(0, 2).
				MarginRight(1)

	SearchStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Primary).
			Foreground(CurrentTheme.Text).
			Background(CurrentTheme.BackgroundAlt).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.Danger).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Danger).
			Padding(1, 2).
			MarginTop(1)

	WarningStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.Warning).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Warning).
			Padding(1, 2).
			MarginTop(1)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.Primary).
			Bold(true)

	ActiveRowStyle = lipgloss.NewStyle().
			Background(CurrentTheme.BackgroundActive).
			Foreground(CurrentTheme.Text).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(CurrentTheme.Primary).
			Padding(0, 1).
			Bold(true)

	NormalRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	DetailKeyStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.Secondary).
			Width(16).
			Bold(true).
			Align(lipgloss.Right).
			MarginRight(1)

	DetailValStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.Text)

	StatusUp   = lipgloss.NewStyle().Foreground(CurrentTheme.Success).Bold(true).Render("● UP")
	StatusDown = lipgloss.NewStyle().Foreground(CurrentTheme.Danger).Bold(true).Render("● DOWN")
)
