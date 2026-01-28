package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name string

	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Danger  lipgloss.Color
	Info    lipgloss.Color

	// Text colors
	Text       lipgloss.Color
	TextDim    lipgloss.Color
	TextMuted  lipgloss.Color
	TextInvert lipgloss.Color

	// Background colors
	Background       lipgloss.Color
	BackgroundAlt    lipgloss.Color
	BackgroundActive lipgloss.Color

	// Border colors
	Border       lipgloss.Color
	BorderActive lipgloss.Color
	BorderMuted  lipgloss.Color
}

var (
	// Current active theme
	CurrentTheme = CyberpunkTheme

	// Available themes
	CyberpunkTheme = Theme{
		Name:             "Cyberpunk",
		Primary:          lipgloss.Color("#00ffff"),
		Secondary:        lipgloss.Color("#ff00ff"),
		Accent:           lipgloss.Color("#ffff00"),
		Success:          lipgloss.Color("#00ff00"),
		Warning:          lipgloss.Color("#ff9500"),
		Danger:           lipgloss.Color("#ff0055"),
		Info:             lipgloss.Color("#00aaff"),
		Text:             lipgloss.Color("#ffffff"),
		TextDim:          lipgloss.Color("#b0b0b0"),
		TextMuted:        lipgloss.Color("#707070"),
		TextInvert:       lipgloss.Color("#000000"),
		Background:       lipgloss.Color("#0a0a0f"),
		BackgroundAlt:    lipgloss.Color("#141420"),
		BackgroundActive: lipgloss.Color("#1a1a30"),
		Border:           lipgloss.Color("#00ffff"),
		BorderActive:     lipgloss.Color("#ff00ff"),
		BorderMuted:      lipgloss.Color("#2a2a3a"),
	}

	NordTheme = Theme{
		Name:             "Nord",
		Primary:          lipgloss.Color("#88c0d0"),
		Secondary:        lipgloss.Color("#81a1c1"),
		Accent:           lipgloss.Color("#a3be8c"),
		Success:          lipgloss.Color("#a3be8c"),
		Warning:          lipgloss.Color("#ebcb8b"),
		Danger:           lipgloss.Color("#bf616a"),
		Info:             lipgloss.Color("#5e81ac"),
		Text:             lipgloss.Color("#eceff4"),
		TextDim:          lipgloss.Color("#d8dee9"),
		TextMuted:        lipgloss.Color("#4c566a"),
		TextInvert:       lipgloss.Color("#2e3440"),
		Background:       lipgloss.Color("#2e3440"),
		BackgroundAlt:    lipgloss.Color("#3b4252"),
		BackgroundActive: lipgloss.Color("#434c5e"),
		Border:           lipgloss.Color("#4c566a"),
		BorderActive:     lipgloss.Color("#88c0d0"),
		BorderMuted:      lipgloss.Color("#3b4252"),
	}

	DraculaTheme = Theme{
		Name:             "Dracula",
		Primary:          lipgloss.Color("#bd93f9"),
		Secondary:        lipgloss.Color("#ff79c6"),
		Accent:           lipgloss.Color("#8be9fd"),
		Success:          lipgloss.Color("#50fa7b"),
		Warning:          lipgloss.Color("#ffb86c"),
		Danger:           lipgloss.Color("#ff5555"),
		Info:             lipgloss.Color("#8be9fd"),
		Text:             lipgloss.Color("#f8f8f2"),
		TextDim:          lipgloss.Color("#f8f8f2"),
		TextMuted:        lipgloss.Color("#6272a4"),
		TextInvert:       lipgloss.Color("#282a36"),
		Background:       lipgloss.Color("#282a36"),
		BackgroundAlt:    lipgloss.Color("#44475a"),
		BackgroundActive: lipgloss.Color("#44475a"),
		Border:           lipgloss.Color("#6272a4"),
		BorderActive:     lipgloss.Color("#bd93f9"),
		BorderMuted:      lipgloss.Color("#44475a"),
	}

	TokyoNightTheme = Theme{
		Name:             "Tokyo Night",
		Primary:          lipgloss.Color("#7aa2f7"),
		Secondary:        lipgloss.Color("#bb9af7"),
		Accent:           lipgloss.Color("#7dcfff"),
		Success:          lipgloss.Color("#9ece6a"),
		Warning:          lipgloss.Color("#e0af68"),
		Danger:           lipgloss.Color("#f7768e"),
		Info:             lipgloss.Color("#7dcfff"),
		Text:             lipgloss.Color("#c0caf5"),
		TextDim:          lipgloss.Color("#a9b1d6"),
		TextMuted:        lipgloss.Color("#565f89"),
		TextInvert:       lipgloss.Color("#1a1b26"),
		Background:       lipgloss.Color("#1a1b26"),
		BackgroundAlt:    lipgloss.Color("#24283b"),
		BackgroundActive: lipgloss.Color("#414868"),
		Border:           lipgloss.Color("#414868"),
		BorderActive:     lipgloss.Color("#7aa2f7"),
		BorderMuted:      lipgloss.Color("#24283b"),
	}

	GruvboxTheme = Theme{
		Name:             "Gruvbox",
		Primary:          lipgloss.Color("#83a598"),
		Secondary:        lipgloss.Color("#d3869b"),
		Accent:           lipgloss.Color("#8ec07c"),
		Success:          lipgloss.Color("#b8bb26"),
		Warning:          lipgloss.Color("#fabd2f"),
		Danger:           lipgloss.Color("#fb4934"),
		Info:             lipgloss.Color("#83a598"),
		Text:             lipgloss.Color("#ebdbb2"),
		TextDim:          lipgloss.Color("#d5c4a1"),
		TextMuted:        lipgloss.Color("#928374"),
		TextInvert:       lipgloss.Color("#282828"),
		Background:       lipgloss.Color("#282828"),
		BackgroundAlt:    lipgloss.Color("#3c3836"),
		BackgroundActive: lipgloss.Color("#504945"),
		Border:           lipgloss.Color("#504945"),
		BorderActive:     lipgloss.Color("#83a598"),
		BorderMuted:      lipgloss.Color("#3c3836"),
	}

	MonokaiTheme = Theme{
		Name:             "Monokai",
		Primary:          lipgloss.Color("#66d9ef"),
		Secondary:        lipgloss.Color("#ae81ff"),
		Accent:           lipgloss.Color("#a1efe4"),
		Success:          lipgloss.Color("#a6e22e"),
		Warning:          lipgloss.Color("#e6db74"),
		Danger:           lipgloss.Color("#f92672"),
		Info:             lipgloss.Color("#66d9ef"),
		Text:             lipgloss.Color("#f8f8f2"),
		TextDim:          lipgloss.Color("#f8f8f2"),
		TextMuted:        lipgloss.Color("#75715e"),
		TextInvert:       lipgloss.Color("#272822"),
		Background:       lipgloss.Color("#272822"),
		BackgroundAlt:    lipgloss.Color("#3e3d32"),
		BackgroundActive: lipgloss.Color("#49483e"),
		Border:           lipgloss.Color("#49483e"),
		BorderActive:     lipgloss.Color("#66d9ef"),
		BorderMuted:      lipgloss.Color("#3e3d32"),
	}
)

// GetThemeByName returns a theme by name, defaults to Cyberpunk
func GetThemeByName(name string) Theme {
	switch name {
	case "nord", "Nord":
		return NordTheme
	case "dracula", "Dracula":
		return DraculaTheme
	case "tokyo", "tokyonight", "Tokyo Night":
		return TokyoNightTheme
	case "gruvbox", "Gruvbox":
		return GruvboxTheme
	case "monokai", "Monokai":
		return MonokaiTheme
	case "cyberpunk", "Cyberpunk":
		return CyberpunkTheme
	default:
		return CyberpunkTheme
	}
}

// ListThemes returns all available theme names
func ListThemes() []string {
	return []string{
		"Cyberpunk",
		"Nord",
		"Dracula",
		"Tokyo Night",
		"Gruvbox",
		"Monokai",
	}
}

// SaveTheme saves the current theme name to config file
func SaveTheme(themeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configPath := filepath.Join(homeDir, ".netdisco-tui-theme")
	return os.WriteFile(configPath, []byte(themeName), 0644)
}

// LoadTheme loads the saved theme from config file
func LoadTheme() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // Use default theme
	}
	
	configPath := filepath.Join(homeDir, ".netdisco-tui-theme")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return // Use default theme
	}
	
	themeName := strings.TrimSpace(string(data))
	CurrentTheme = GetThemeByName(themeName)
	UpdateStylesForTheme()
}
