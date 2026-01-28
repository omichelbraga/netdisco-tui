package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplashModel struct {
	width, height int
	frame         int
	done          bool
}

type splashTickMsg time.Time
type splashDoneMsg struct{}

func NewSplashModel(width, height int) SplashModel {
	return SplashModel{
		width:  width,
		height: height,
		frame:  0,
		done:   false,
	}
}

func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(
		tickSplash(),
		func() tea.Msg {
			time.Sleep(3 * time.Second)
			return splashDoneMsg{}
		},
	)
}

func tickSplash() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return splashTickMsg(t)
	})
}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg.(type) {
	case splashTickMsg:
		m.frame++
		if m.frame > 30 {
			m.done = true
			return m, nil
		}
		return m, tickSplash()
	case splashDoneMsg:
		m.done = true
		return m, nil
	}
	return m, nil
}

func (m SplashModel) View() string {
	if m.done {
		return ""
	}

	// ASCII art logo
	logo := []string{
		"███╗   ██╗███████╗████████╗██████╗ ██╗███████╗ ██████╗ ██████╗ ",
		"████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝██╔═══██╗",
		"██╔██╗ ██║█████╗     ██║   ██║  ██║██║███████╗██║     ██║   ██║",
		"██║╚██╗██║██╔══╝     ██║   ██║  ██║██║╚════██║██║     ██║   ██║",
		"██║ ╚████║███████╗   ██║   ██████╔╝██║███████║╚██████╗╚██████╔╝",
		"╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═════╝ ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ",
	}

	subtitle := "Network Discovery Terminal Interface"
	author := "by Mike Guimaraes"

	// Calculate animation based on frame
	var lines []string
	
	// Fade in effect (first 10 frames)
	if m.frame < 10 {
		opacity := m.frame
		for _, line := range logo {
			if opacity > len(line)/10 {
				lines = append(lines, lipgloss.NewStyle().Foreground(CurrentTheme.Primary).Render(line))
			} else {
				truncated := line[:opacity*len(line)/10]
				lines = append(lines, lipgloss.NewStyle().Foreground(CurrentTheme.Primary).Render(truncated))
			}
		}
	} else {
		// Full logo with color cycling
		color := CurrentTheme.Primary
		if m.frame%6 == 0 {
			color = CurrentTheme.Primary
		} else if m.frame%6 == 1 {
			color = CurrentTheme.Secondary
		} else if m.frame%6 == 2 {
			color = CurrentTheme.Accent
		} else if m.frame%6 == 3 {
			color = CurrentTheme.Primary
		} else if m.frame%6 == 4 {
			color = CurrentTheme.Secondary
		} else {
			color = CurrentTheme.Accent
		}
		
		for _, line := range logo {
			lines = append(lines, lipgloss.NewStyle().Foreground(color).Bold(true).Render(line))
		}
	}

	// Add subtitle and author
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().
		Foreground(CurrentTheme.Secondary).
		Italic(true).
		Render(centerText(subtitle, 68)))
	
	lines = append(lines, "")
	
	// Animated dots after author name
	dots := strings.Repeat(".", m.frame%4)
	authorWithDots := author + dots
	lines = append(lines, lipgloss.NewStyle().
		Foreground(CurrentTheme.TextMuted).
		Render(centerText(authorWithDots, 68)))

	// Progress indicator
	if m.frame > 10 {
		progress := (m.frame - 10) * 100 / 20
		if progress > 100 {
			progress = 100
		}
		barWidth := 50
		filled := progress * barWidth / 100
		bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled) + "]"
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			Foreground(CurrentTheme.Accent).
			Render(centerText(bar, 68)))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(CurrentTheme.TextMuted).
			Render(centerText(fmt.Sprintf("Loading... %d%%", progress), 68)))
	}

	content := strings.Join(lines, "\n")

	// Center vertically
	topPadding := (m.height - lipgloss.Height(content)) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	return strings.Repeat("\n", topPadding) + centerHorizontal(content, m.width)
}

func centerText(text string, width int) string {
	textWidth := len(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func centerHorizontal(content string, width int) string {
	lines := strings.Split(content, "\n")
	var centered []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth >= width {
			centered = append(centered, line)
		} else {
			padding := (width - lineWidth) / 2
			centered = append(centered, strings.Repeat(" ", padding)+line)
		}
	}
	return strings.Join(centered, "\n")
}
