package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/netdisco-tui/netdisco-tui/internal/api"
)

const (
	TabDevices = iota
	TabNodes
	TabVlans
	TabSubnets
	TabReports
)

var tabNames = []string{"Devices", "Nodes", "VLANs", "Subnets", "Reports"}

type RootModel struct {
	width, height int
	client        *api.Client
	activeTab     int
	initialized   []bool
	showThemeMenu bool
	themeZones    []themeClickZone
	showingSplash bool
	splash        SplashModel

	devices DevicesModel
	nodes   NodesModel
	vlans   VlansModel
	subnets SubnetsModel
	reports *ReportsModel
}

type themeClickZone struct {
	minY, maxY, minX, maxX int
	themeIdx               int
}

func NewRootModel(width, height int, client *api.Client) RootModel {
	// Load saved theme on startup
	LoadTheme()
	
	reports := NewReportsModel(width, height-4, client)
	return RootModel{
		width:         width,
		height:        height,
		client:        client,
		activeTab:     TabDevices,
		initialized:   make([]bool, 5),
		showingSplash: true,
		splash:        NewSplashModel(width, height),
		devices:       NewDevicesModel(width, height-4, client),
		nodes:         NewNodesModel(width, height-4, client),
		vlans:         NewVlansModel(width, height-4, client),
		subnets:       NewSubnetsModel(width, height-4, client),
		reports:       &reports,
	}
}

func (m RootModel) Init() tea.Cmd {
	if m.showingSplash {
		return m.splash.Init()
	}
	m.initialized[TabDevices] = true
	return m.devices.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle splash screen
	if m.showingSplash {
		var cmd tea.Cmd
		m.splash, cmd = m.splash.Update(msg)
		if m.splash.done {
			m.showingSplash = false
			m.initialized[TabDevices] = true
			return m, m.devices.Init()
		}
		return m, cmd
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := m.height - 4
		m.devices.width = m.width
		m.devices.height = contentHeight
		m.nodes.width = m.width
		m.nodes.height = contentHeight
		m.vlans.width = m.width
		m.vlans.height = contentHeight
		m.subnets.width = m.width
		m.subnets.height = contentHeight
		m.reports.width = m.width
		m.reports.height = contentHeight
		return m, nil

	case tea.MouseMsg:
		// Handle clicks on theme menu if it's open
		if m.showThemeMenu && msg.Type == tea.MouseLeft {
			y := msg.Y
			x := msg.X
			
			// Use the calculated zones from renderThemeMenu
			for _, zone := range m.themeZones {
				if y >= zone.minY && y <= zone.maxY && x >= zone.minX && x <= zone.maxX {
					themes := []Theme{CyberpunkTheme, NordTheme, DraculaTheme, TokyoNightTheme, GruvboxTheme, MonokaiTheme}
					CurrentTheme = themes[zone.themeIdx]
					UpdateStylesForTheme()
					SaveTheme(CurrentTheme.Name) // Save theme selection
					m.showThemeMenu = false
					return m, nil
				}
			}
			return m, nil
		}
		
		// Handle clicks on tab bar (only if theme menu is closed)
		if !m.showThemeMenu && msg.Type == tea.MouseLeft && (msg.Y >= 1 && msg.Y <= 4) {
			// Calculate actual tab positions by measuring rendered widths
			x := msg.X
			tabEmojis := []string{"📱", "🔍", "🌐", "📊", "📋"}
			currentX := 2 // Starting X position (accounting for padding)
			
			for i, name := range tabNames {
				label := fmt.Sprintf("%s %s", tabEmojis[i], name)
				var tabWidth int
				if i == m.activeTab {
					tabWidth = lipgloss.Width(TabActiveStyle.Render(label))
				} else {
					tabWidth = lipgloss.Width(TabInactiveStyle.Render(label))
				}
				
				// Check if click is within this tab's bounds
				if x >= currentX && x < currentX+tabWidth {
					return m.switchTab(i)
				}
				
				currentX += tabWidth
			}
		}
		// If not clicking on tabs or theme menu, pass to active tab

	case tea.KeyMsg:
		// Check if any input is focused before handling number keys
		inputFocused := false
		switch m.activeTab {
		case TabDevices:
			inputFocused = m.devices.searchInput.Focused()
		case TabNodes:
			inputFocused = m.nodes.searchInput.Focused()
		case TabSubnets:
			inputFocused = m.subnets.searchInput.Focused()
		}

		// Only handle tab switching if no input is focused and theme menu is closed
		if !inputFocused && !m.showThemeMenu {
			switch msg.String() {
			case "1":
				return m.switchTab(TabDevices)
			case "2":
				return m.switchTab(TabNodes)
			case "3":
				return m.switchTab(TabVlans)
			case "4":
				return m.switchTab(TabSubnets)
			case "5":
				return m.switchTab(TabReports)
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "t", "T":
			// Toggle theme menu
			m.showThemeMenu = !m.showThemeMenu
			if m.showThemeMenu {
				// Pre-calculate theme zones when opening menu
				themes := []Theme{CyberpunkTheme, NordTheme, DraculaTheme, TokyoNightTheme, GruvboxTheme, MonokaiTheme}
				m.themeZones = []themeClickZone{}
				
				// Calculate actual header and tab bar heights
				// Header bar
				headerLeft := "⚡ NETDISCO TUI"
				headerRight := fmt.Sprintf("Theme: %s [T]", CurrentTheme.Name)
				header := headerLeft + strings.Repeat(" ", m.width-len(headerLeft)-len(headerRight)) + headerRight
				headerBar := lipgloss.NewStyle().
					Background(CurrentTheme.BackgroundAlt).
					Foreground(CurrentTheme.Text).
					Padding(0, 2).
					Render(header)
				headerHeight := lipgloss.Height(headerBar)
				
				// Tab bar
				tabEmojis := []string{"📱", "🔍", "🌐", "📊", "📋"}
				var tabs []string
				for i, name := range tabNames {
					label := fmt.Sprintf("%s %s", tabEmojis[i], name)
					if i == m.activeTab {
						tabs = append(tabs, TabActiveStyle.Render(label))
					} else {
						tabs = append(tabs, TabInactiveStyle.Render(label))
					}
				}
				tabBar := lipgloss.NewStyle().
					Padding(1, 2).
					Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
				tabBarHeight := lipgloss.Height(tabBar)
				
				// Simulate rendering to get heights
				title := lipgloss.NewStyle().
					Bold(true).
					Foreground(CurrentTheme.Primary).
					Padding(1, 0).
					Render("🎨  SELECT THEME")
				
				// Start after header/tabs, then add: outer padding + title + blank line
				currentY := headerHeight + tabBarHeight + 2 + lipgloss.Height(title) + 1
				
				var itemHeights []int
				var itemWidth int
				
				for i, theme := range themes {
					previewBox := lipgloss.NewStyle().
						Width(40).
						Height(3).
						Background(theme.Background).
						BorderStyle(lipgloss.RoundedBorder()).
						BorderForeground(theme.Border).
						Padding(0, 1).
						Render(
							lipgloss.JoinVertical(lipgloss.Left,
								lipgloss.NewStyle().Foreground(theme.Primary).Render(fmt.Sprintf("■ Primary  ")+
									lipgloss.NewStyle().Foreground(theme.Secondary).Render("■ Secondary  ")+
									lipgloss.NewStyle().Foreground(theme.Accent).Render("■ Accent")),
								lipgloss.NewStyle().Foreground(theme.Success).Render("● Success  ")+
									lipgloss.NewStyle().Foreground(theme.Warning).Render("● Warning  ")+
									lipgloss.NewStyle().Foreground(theme.Danger).Render("● Danger"),
							),
						)

					label := fmt.Sprintf("[%d] %s", i+1, theme.Name)
					if theme.Name == CurrentTheme.Name {
						label += " ✓"
						label = lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render(label)
					} else {
						label = lipgloss.NewStyle().Foreground(CurrentTheme.TextMuted).Render(label)
					}

					item := lipgloss.JoinVertical(lipgloss.Left, label, previewBox)
					itemHeights = append(itemHeights, lipgloss.Height(item))
					if i == 0 {
						itemWidth = lipgloss.Width(item)
					}
				}
				
				// Left column zones (themes 0, 1, 2)
				leftColX := 2
				yPos := currentY
				for i := 0; i < 3; i++ {
					m.themeZones = append(m.themeZones, themeClickZone{
						minY:     yPos,
						maxY:     yPos + itemHeights[i] - 1,
						minX:     leftColX,
						maxX:     leftColX + itemWidth,
						themeIdx: i,
					})
					yPos += itemHeights[i] + 1 // +1 for MarginTop
				}
				
				// Right column zones (themes 3, 4, 5)
				rightColX := leftColX + itemWidth + 2
				yPos = currentY
				for i := 3; i < 6; i++ {
					m.themeZones = append(m.themeZones, themeClickZone{
						minY:     yPos,
						maxY:     yPos + itemHeights[i] - 1,
						minX:     rightColX,
						maxX:     rightColX + itemWidth,
						themeIdx: i,
					})
					yPos += itemHeights[i] + 1
				}
			}
			return m, nil
		case "esc":
			// Close theme menu if open
			if m.showThemeMenu {
				m.showThemeMenu = false
				return m, nil
			}
		case "1", "2", "3", "4", "5", "6":
			// Theme selection when menu is open
			if m.showThemeMenu {
				themes := []Theme{CyberpunkTheme, NordTheme, DraculaTheme, TokyoNightTheme, GruvboxTheme, MonokaiTheme}
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(themes) {
					CurrentTheme = themes[idx]
					UpdateStylesForTheme()
					SaveTheme(CurrentTheme.Name) // Save theme selection
					m.showThemeMenu = false
					return m, nil
				}
			}
		}

		// Only pass left/right to root if not in a detail view or Reports tab
		if msg.String() == "left" || msg.String() == "right" {
			passToChild := false
			switch m.activeTab {
			case TabDevices:
				passToChild = m.devices.inDetail
			case TabReports:
				// Always pass to Reports so it can handle report type switching
				passToChild = true
			}
			if !passToChild {
				if msg.String() == "left" && m.activeTab > 0 {
					return m.switchTab(m.activeTab - 1)
				}
				if msg.String() == "right" && m.activeTab < len(tabNames)-1 {
					return m.switchTab(m.activeTab + 1)
				}
				return m, nil
			}
		}
	}

	// Delegate to active tab
	var cmd tea.Cmd
	switch m.activeTab {
	case TabDevices:
		m.devices, cmd = m.devices.Update(msg)
	case TabNodes:
		m.nodes, cmd = m.nodes.Update(msg)
	case TabVlans:
		m.vlans, cmd = m.vlans.Update(msg)
	case TabSubnets:
		m.subnets, cmd = m.subnets.Update(msg)
	case TabReports:
		m.reports, cmd = m.reports.Update(msg)
	}

	return m, cmd
}

func (m RootModel) switchTab(tab int) (RootModel, tea.Cmd) {
	m.activeTab = tab
	if !m.initialized[tab] {
		m.initialized[tab] = true
		switch tab {
		case TabDevices:
			return m, m.devices.Init()
		case TabNodes:
			return m, m.nodes.Init()
		case TabVlans:
			return m, m.vlans.Init()
		case TabSubnets:
			return m, m.subnets.Init()
		case TabReports:
			return m, m.reports.Init()
		}
	}
	return m, nil
}

func (m RootModel) View() string {
	// Show splash screen if active
	if m.showingSplash {
		return m.splash.View()
	}
	
	// Header with app title and theme
	headerLeft := lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary).
		Render("⚡ NETDISCO TUI")

	headerRight := lipgloss.NewStyle().
		Foreground(CurrentTheme.TextMuted).
		Render(fmt.Sprintf("Theme: %s [T]", CurrentTheme.Name))

	// Calculate spacer to fill the width
	leftWidth := lipgloss.Width(headerLeft)
	rightWidth := lipgloss.Width(headerRight)
	// Account for 4 chars of padding (2 on each side)
	availableWidth := m.width - 4
	spacer := availableWidth - leftWidth - rightWidth
	if spacer < 1 {
		spacer = 1
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		headerLeft,
		strings.Repeat(" ", spacer),
		headerRight,
	)

	headerBar := lipgloss.NewStyle().
		Background(CurrentTheme.BackgroundAlt).
		Foreground(CurrentTheme.Text).
		Padding(0, 2).
		Render(header)

	// Tab bar
	var tabs []string
	tabEmojis := []string{"📱", "🔍", "🌐", "📊", "📋"}
	for i, name := range tabNames {
		// Add emoji for visual appeal and easier identification
		label := fmt.Sprintf("%s %s", tabEmojis[i], name)
		if i == m.activeTab {
			tabs = append(tabs, TabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, TabInactiveStyle.Render(label))
		}
	}
	tabBar := lipgloss.NewStyle().
		Padding(1, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))

	// Content
	var content string
	if m.showThemeMenu {
		content = m.renderThemeMenu()
	} else {
		switch m.activeTab {
		case TabDevices:
			content = m.devices.View()
		case TabNodes:
			content = m.nodes.View()
		case TabVlans:
			content = m.vlans.View()
		case TabSubnets:
			content = m.subnets.View()
		case TabReports:
			content = m.reports.View()
		}
	}

	// Ensure content is not empty
	if content == "" {
		content = lipgloss.NewStyle().
			Foreground(CurrentTheme.TextMuted).
			Render("Loading...")
	}

	// Calculate and enforce height constraints to prevent scrolling
	// Header bar = 1 row, Tab bar = 3 rows, Footer = 1 row, margins = 2
	usedHeight := 7
	maxContentHeight := m.height - usedHeight
	if maxContentHeight < 10 {
		maxContentHeight = 10
	}

	// Truncate content if it exceeds available height
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > maxContentHeight {
		contentLines = contentLines[:maxContentHeight]
		content = strings.Join(contentLines, "\n")
	}

	// Footer with controls
	footerStyle := lipgloss.NewStyle().
		Background(CurrentTheme.BackgroundAlt).
		Foreground(CurrentTheme.TextMuted).
		Padding(0, 2)

	var footerText string
	if m.showThemeMenu {
		footerText = "Press 1-6 to select theme  ·  ESC or T to close"
	} else {
		footerText = "q quit  ·  T theme  ·  ←→ tabs  ·  Click on tabs to switch"
	}
	footer := footerStyle.Render(footerText)

	return lipgloss.JoinVertical(lipgloss.Left, headerBar, tabBar, content, footer)
}

func (m RootModel) renderThemeMenu() string {
	themes := []Theme{CyberpunkTheme, NordTheme, DraculaTheme, TokyoNightTheme, GruvboxTheme, MonokaiTheme}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary).
		Padding(1, 0).
		Render("🎨  SELECT THEME")

	var themeItems []string
	for i, theme := range themes {
		// Create a preview box with theme colors
		previewBox := lipgloss.NewStyle().
			Width(40).
			Height(3).
			Background(theme.Background).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(0, 1).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					lipgloss.NewStyle().Foreground(theme.Primary).Render(fmt.Sprintf("■ Primary  ")+
						lipgloss.NewStyle().Foreground(theme.Secondary).Render("■ Secondary  ")+
						lipgloss.NewStyle().Foreground(theme.Accent).Render("■ Accent")),
					lipgloss.NewStyle().Foreground(theme.Success).Render("● Success  ")+
						lipgloss.NewStyle().Foreground(theme.Warning).Render("● Warning  ")+
						lipgloss.NewStyle().Foreground(theme.Danger).Render("● Danger"),
				),
			)

		label := fmt.Sprintf("[%d] %s", i+1, theme.Name)
		if theme.Name == CurrentTheme.Name {
			label += " ✓"
			label = lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render(label)
		} else {
			label = lipgloss.NewStyle().Foreground(CurrentTheme.TextMuted).Render(label)
		}

		item := lipgloss.JoinVertical(lipgloss.Left, label, previewBox)
		themeItems = append(themeItems, item)
	}

	// Arrange in 2 columns
	col1 := lipgloss.JoinVertical(lipgloss.Left,
		themeItems[0],
		lipgloss.NewStyle().MarginTop(1).Render(themeItems[1]),
		lipgloss.NewStyle().MarginTop(1).Render(themeItems[2]),
	)

	col2 := lipgloss.JoinVertical(lipgloss.Left,
		themeItems[3],
		lipgloss.NewStyle().MarginTop(1).Render(themeItems[4]),
		lipgloss.NewStyle().MarginTop(1).Render(themeItems[5]),
	)

	grid := lipgloss.JoinHorizontal(lipgloss.Top, col1, "  ", col2)

	menu := lipgloss.NewStyle().
		Padding(2).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", grid))

	return menu
}
