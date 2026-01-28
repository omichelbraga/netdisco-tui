package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#00d4ff")
	ColorSecondary = lipgloss.Color("#7c3aed")
	ColorSuccess   = lipgloss.Color("#22c55e")
	ColorDanger    = lipgloss.Color("#ef4444")
	ColorWarning   = lipgloss.Color("#f59e0b")
	ColorText      = lipgloss.Color("#e2e8f0")
	ColorTextDim   = lipgloss.Color("#94a3b8")
	ColorTextMuted = lipgloss.Color("#64748b")
	ColorBorder    = lipgloss.Color("#334155")
	ColorBg        = lipgloss.Color("#0f172a")
	ColorBgActive  = lipgloss.Color("#1e293b")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Padding(0, 2)

	SearchStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	ActiveRowStyle = lipgloss.NewStyle().
			Background(ColorBgActive).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	NormalRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	DetailKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Width(14).
			Bold(true)

	DetailValStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	StatusUp   = lipgloss.NewStyle().Foreground(ColorSuccess).Render("● Up")
	StatusDown = lipgloss.NewStyle().Foreground(ColorDanger).Render("● Down")
)
