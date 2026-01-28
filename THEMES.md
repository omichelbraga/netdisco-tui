# 🎨 Netdisco TUI - Themes

Netdisco TUI now supports beautiful, modern color themes!

## Available Themes

### 1. Cyberpunk (Default)
Neon cyan and magenta on dark background - futuristic and vibrant
- Primary: Cyan (#00ffff)
- Secondary: Magenta (#ff00ff)
- Accent: Yellow (#ffff00)

### 2. Nord
Cool Nordic color palette - calm and professional
- Primary: Frost Blue (#88c0d0)
- Secondary: Polar Blue (#81a1c1)
- Accent: Aurora Green (#a3be8c)

### 3. Dracula
Popular dark theme with purple accents
- Primary: Purple (#bd93f9)
- Secondary: Pink (#ff79c6)
- Accent: Cyan (#8be9fd)

### 4. Tokyo Night
Modern dark theme inspired by Tokyo at night
- Primary: Blue (#7aa2f7)
- Secondary: Purple (#bb9af7)
- Accent: Cyan (#7dcfff)

### 5. Gruvbox
Warm retro colors with high contrast
- Primary: Blue (#83a598)
- Secondary: Purple (#d3869b)
- Accent: Aqua (#8ec07c)

### 6. Monokai
Classic developer theme
- Primary: Cyan (#66d9ef)
- Secondary: Purple (#ae81ff)
- Accent: Aqua (#a1efe4)

## How to Switch Themes

1. Press `T` to open the theme menu
2. You'll see all 6 themes with live color previews
3. Press `1-6` to select a theme
4. Press `T` again to close the menu

The current theme is marked with a ✓ checkmark.

## Features

- **Live Preview**: See each theme's colors before selecting
- **Instant Switch**: Themes change immediately, no restart needed
- **Session Persistent**: Theme stays active for your entire session
- **Complete Coverage**: Themes apply to all UI elements:
  - Headers and titles
  - Tables and borders
  - Status indicators
  - Error/warning messages
  - Active/inactive elements
  - Search bars
  - Footers

## Theme Indicator

The current theme name is always shown in the top-right corner of the interface:
```
⚡ NETDISCO TUI                                    Theme: Cyberpunk [T]
```

Press `T` anytime to change it!

## For Developers

Themes are defined in `internal/tui/themes.go`. Each theme is a struct with 15+ color properties.

To add a new theme:
1. Create a new `Theme` struct with your colors
2. Add it to the themes list in `GetThemeByName()`
3. Add the name to `ListThemes()`
4. It will automatically appear in the theme menu!
