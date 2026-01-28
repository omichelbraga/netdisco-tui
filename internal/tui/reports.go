package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/netdisco-tui/netdisco-tui/internal/api"
)

type recentDevicesMsg struct {
	devices []map[string]interface{}
	err     error
}

type deviceInventoryMsg struct {
	devices []map[string]interface{}
	err     error
}

type vlanInventoryMsg struct {
	vlans []map[string]interface{}
	err   error
}

func loadRecentDevicesCmd(client *api.Client, days int) tea.Cmd {
	return func() tea.Msg {
		devices, err := client.GetRecentDevices(days)
		return recentDevicesMsg{devices: devices, err: err}
	}
}

func loadDeviceInventoryCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		devices, err := client.GetDeviceInventory()
		return deviceInventoryMsg{devices: devices, err: err}
	}
}

func loadVlanInventoryCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		vlans, err := client.GetVlanInventory()
		return vlanInventoryMsg{vlans: vlans, err: err}
	}
}

const (
	ReportRecent = iota
	ReportAllDevices
	ReportByVendor
	ReportVLANs
)

type ReportsModel struct {
	width, height int
	client        *api.Client
	reportType    int
	devices       []map[string]interface{}
	vlans         []map[string]interface{}
	vendorStats   []map[string]interface{}
	cursor        int
	loading       bool
	err           error
	spinner       spinner.Model
	days          int

	// Resizable tables (one for each report type)
	tableRecent     ResizableTable
	tableAllDevices ResizableTable
	tableVendor     ResizableTable
	tableVLANs      ResizableTable
}

func NewReportsModel(width, height int, client *api.Client) ReportsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	tableRecent := NewResizableTable([]int{24, 16, 14, 16, 18, 18}) // Name, IP, Vendor, Model, Location, Discovered
	tableRecent.SortColumn = 5                                       // Default sort by Discovered
	tableRecent.SortAscending = false                                // Newest first

	tableAllDevices := NewResizableTable([]int{22, 15, 14, 24, 18}) // Name, IP, Vendor, Model, Location
	tableAllDevices.SortColumn = 0                                   // Default sort by Name
	tableAllDevices.SortAscending = true

	tableVendor := NewResizableTable([]int{30, 10, 12}) // Vendor, Count, Percentage
	tableVendor.SortColumn = 1                           // Default sort by Count
	tableVendor.SortAscending = false                    // Highest first

	tableVLANs := NewResizableTable([]int{8, 24, 40}) // VLAN, Name, Description
	tableVLANs.SortColumn = 0                          // Default sort by VLAN ID
	tableVLANs.SortAscending = true

	return ReportsModel{
		width:           width,
		height:          height,
		client:          client,
		loading:         true,
		spinner:         s,
		days:            7,
		reportType:      ReportAllDevices,
		tableRecent:     tableRecent,
		tableAllDevices: tableAllDevices,
		tableVendor:     tableVendor,
		tableVLANs:      tableVLANs,
	}
}

func (m ReportsModel) Init() tea.Cmd {
	return loadDeviceInventoryCmd(m.client)
}

func (m *ReportsModel) Update(msg tea.Msg) (*ReportsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case recentDevicesMsg:
		m.loading = false
		m.err = msg.err
		m.devices = msg.devices
		m.cursor = 0
		return m, nil

	case deviceInventoryMsg:
		m.loading = false
		m.err = msg.err
		m.devices = msg.devices
		m.cursor = 0

		// Build vendor stats
		if m.reportType == ReportByVendor {
			m.vendorStats = m.buildVendorStats()
		}
		return m, nil

	case vlanInventoryMsg:
		m.loading = false
		m.err = msg.err
		m.vlans = msg.vlans
		m.cursor = 0
		return m, nil

	case tea.MouseMsg:
		// Calculate actual Y position of sub-tabs and table header dynamically
		// Build the sub-tab bar to measure its height
		tabEmojis := []string{"⏱", "📊", "📦", "🌐"}
		reportLabels := []string{"Recent", "All Devices", "By Vendor", "VLANs"}
		var tabBar []string
		for i, label := range reportLabels {
			labelWithEmoji := tabEmojis[i] + "  " + label
			if i == m.reportType {
				tabBar = append(tabBar, TabActiveStyle.Render(labelWithEmoji))
			} else {
				tabBar = append(tabBar, TabInactiveStyle.Render(labelWithEmoji))
			}
		}
		tabs := lipgloss.NewStyle().
			Padding(1, 2).
			Render(lipgloss.JoinHorizontal(lipgloss.Top, tabBar...))
		
		// Calculate Y positions:
		// Root UI: 4 lines (header bar + tab bar)
		// tabs: measured height
		// blank line after tabs: 1
		// header (title) line: 1
		// controls line: 1
		// blank line before table: 2 (seems to be 2 lines in actual rendering)
		rootUIHeight := 4
		tabsHeight := lipgloss.Height(tabs)
		blankLineAfterTabs := 1
		headerLine := 1
		controlsLine := 1
		blankLineBeforeTable := 2  // Increased from 1 to 2
		
		// Sub-tabs Y range
		subTabStartY := rootUIHeight
		subTabEndY := rootUIHeight + tabsHeight - 1
		
		// Table header Y (add all components above it)
		tableHeaderY := rootUIHeight + tabsHeight + blankLineAfterTabs + headerLine + controlsLine + blankLineBeforeTable
		
		// Check for sub-tab clicks FIRST
		if msg.Type == tea.MouseLeft && (msg.Y >= subTabStartY && msg.Y <= subTabEndY) {
			x := msg.X
			
			var newReportType int
			var shouldSwitch bool
			
			// Detect which sub-tab was clicked
			if x >= 0 && x < 18 {
				newReportType = ReportRecent
				shouldSwitch = true
			} else if x >= 18 && x < 36 {
				newReportType = ReportAllDevices
				shouldSwitch = true
			} else if x >= 36 && x < 54 {
				newReportType = ReportByVendor
				shouldSwitch = true
			} else if x >= 54 && x < 120 {
				newReportType = ReportVLANs
				shouldSwitch = true
			}
			
			if shouldSwitch && m.reportType != newReportType {
				m.reportType = newReportType
				m.cursor = 0
				m.loading = true
				return m, m.loadCurrentReport()
			}
			// Return without sorting if we clicked on sub-tabs
			return m, nil
		}
		
		// Set HeaderY for all tables based on calculated position
		m.tableRecent.HeaderY = tableHeaderY
		m.tableAllDevices.HeaderY = tableHeaderY
		m.tableVendor.HeaderY = tableHeaderY
		m.tableVLANs.HeaderY = tableHeaderY
		
		// Check for header click (sorting)
		// Use rowOffset=2 for some tolerance
		switch m.reportType {
		case ReportRecent:
			if col := m.tableRecent.HandleHeaderClick(msg, 0, 2); col >= 0 {
				if m.tableRecent.SortColumn == col {
					m.tableRecent.SortAscending = !m.tableRecent.SortAscending
				} else {
					m.tableRecent.SortColumn = col
					m.tableRecent.SortAscending = true
				}
				m.sortRecentDevices()
				return m, nil
			}
		case ReportAllDevices:
			if col := m.tableAllDevices.HandleHeaderClick(msg, 0, 2); col >= 0 {
				if m.tableAllDevices.SortColumn == col {
					m.tableAllDevices.SortAscending = !m.tableAllDevices.SortAscending
				} else {
					m.tableAllDevices.SortColumn = col
					m.tableAllDevices.SortAscending = true
				}
				m.sortAllDevices()
				return m, nil
			}
		case ReportByVendor:
			if col := m.tableVendor.HandleHeaderClick(msg, 0, 2); col >= 0 {
				if m.tableVendor.SortColumn == col {
					m.tableVendor.SortAscending = !m.tableVendor.SortAscending
				} else {
					m.tableVendor.SortColumn = col
					m.tableVendor.SortAscending = true
				}
				m.sortVendorStats()
				return m, nil
			}
		case ReportVLANs:
			if col := m.tableVLANs.HandleHeaderClick(msg, 0, 2); col >= 0 {
				if m.tableVLANs.SortColumn == col {
					m.tableVLANs.SortAscending = !m.tableVLANs.SortAscending
				} else {
					m.tableVLANs.SortColumn = col
					m.tableVLANs.SortAscending = true
				}
				m.sortVLANs()
				return m, nil
			}
		}
		
		// Handle table column resizing
		switch m.reportType {
		case ReportRecent:
			m.tableRecent.HandleMouse(msg, 0)
		case ReportAllDevices:
			m.tableAllDevices.HandleMouse(msg, 0)
		case ReportByVendor:
			m.tableVendor.HandleMouse(msg, 0)
		case ReportVLANs:
			m.tableVLANs.HandleMouse(msg, 0)
		}
		return m, nil

	case tea.KeyMsg:
		maxItems := len(m.devices)
		if m.reportType == ReportByVendor {
			maxItems = len(m.vendorStats)
		} else if m.reportType == ReportVLANs {
			maxItems = len(m.vlans)
		}

		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < maxItems-1 {
				m.cursor++
			}
		case "left":
			if m.reportType > 0 {
				m.reportType--
				m.cursor = 0
				m.loading = true
				return m, m.loadCurrentReport()
			}
		case "right":
			if m.reportType < ReportVLANs {
				m.reportType++
				m.cursor = 0
				m.loading = true
				return m, m.loadCurrentReport()
			}
		case "+":
			if m.reportType == ReportRecent {
				m.days += 7
				m.loading = true
				m.err = nil
				return m, loadRecentDevicesCmd(m.client, m.days)
			}
		case "-":
			if m.reportType == ReportRecent && m.days > 1 {
				m.days -= 7
				if m.days < 1 {
					m.days = 1
				}
				m.loading = true
				m.err = nil
				return m, loadRecentDevicesCmd(m.client, m.days)
			}
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadCurrentReport()
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ReportsModel) loadCurrentReport() tea.Cmd {
	switch m.reportType {
	case ReportRecent:
		return loadRecentDevicesCmd(m.client, m.days)
	case ReportAllDevices:
		return loadDeviceInventoryCmd(m.client)
	case ReportByVendor:
		return loadDeviceInventoryCmd(m.client)
	case ReportVLANs:
		return loadVlanInventoryCmd(m.client)
	}
	return nil
}

func (m ReportsModel) buildVendorStats() []map[string]interface{} {
	vendorMap := make(map[string]int)
	for _, dev := range m.devices {
		vendor := getStringField(dev, "vendor")
		if vendor == "" {
			vendor = "unknown"
		}
		vendorMap[vendor]++
	}

	var stats []map[string]interface{}
	for vendor, count := range vendorMap {
		stats = append(stats, map[string]interface{}{
			"vendor": vendor,
			"count":  count,
		})
	}

	// Sort by count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i]["count"].(int) > stats[j]["count"].(int)
	})

	return stats
}

func (m *ReportsModel) sortRecentDevices() {
	if len(m.devices) == 0 {
		return
	}

	for i := 0; i < len(m.devices)-1; i++ {
		for j := i + 1; j < len(m.devices); j++ {
			swap := false
			d1 := m.devices[i]
			d2 := m.devices[j]

			switch m.tableRecent.SortColumn {
			case 0: // Name
				val1 := strings.ToLower(getStringField(d1, "device_name"))
				val2 := strings.ToLower(getStringField(d2, "device_name"))
				if m.tableRecent.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // IP
				val1 := strings.ToLower(getStringField(d1, "ip"))
				val2 := strings.ToLower(getStringField(d2, "ip"))
				if m.tableRecent.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // Vendor
				val1 := strings.ToLower(getStringField(d1, "vendor"))
				val2 := strings.ToLower(getStringField(d2, "vendor"))
				if m.tableRecent.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 3: // Model
				val1 := strings.ToLower(getStringField(d1, "model"))
				val2 := strings.ToLower(getStringField(d2, "model"))
				if m.tableRecent.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 4: // Location
				val1 := strings.ToLower(getStringField(d1, "location"))
				val2 := strings.ToLower(getStringField(d2, "location"))
				if m.tableRecent.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 5: // Discovered
				val1 := strings.ToLower(getStringField(d1, "creation"))
				val2 := strings.ToLower(getStringField(d2, "creation"))
				if m.tableRecent.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.devices[i], m.devices[j] = m.devices[j], m.devices[i]
			}
		}
	}
}

func (m *ReportsModel) sortAllDevices() {
	if len(m.devices) == 0 {
		return
	}

	for i := 0; i < len(m.devices)-1; i++ {
		for j := i + 1; j < len(m.devices); j++ {
			swap := false
			d1 := m.devices[i]
			d2 := m.devices[j]

			switch m.tableAllDevices.SortColumn {
			case 0: // Name
				val1 := strings.ToLower(getStringField(d1, "device_name"))
				val2 := strings.ToLower(getStringField(d2, "device_name"))
				if m.tableAllDevices.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // IP
				val1 := strings.ToLower(getStringField(d1, "ip"))
				val2 := strings.ToLower(getStringField(d2, "ip"))
				if m.tableAllDevices.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // Vendor
				val1 := strings.ToLower(getStringField(d1, "vendor"))
				val2 := strings.ToLower(getStringField(d2, "vendor"))
				if m.tableAllDevices.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 3: // Model
				val1 := strings.ToLower(getStringField(d1, "model"))
				val2 := strings.ToLower(getStringField(d2, "model"))
				if m.tableAllDevices.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 4: // Location
				val1 := strings.ToLower(getStringField(d1, "location"))
				val2 := strings.ToLower(getStringField(d2, "location"))
				if m.tableAllDevices.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.devices[i], m.devices[j] = m.devices[j], m.devices[i]
			}
		}
	}
}

func (m *ReportsModel) sortVendorStats() {
	if len(m.vendorStats) == 0 {
		return
	}

	totalDevices := len(m.devices)
	if totalDevices == 0 {
		totalDevices = 1 // Avoid division by zero
	}

	for i := 0; i < len(m.vendorStats)-1; i++ {
		for j := i + 1; j < len(m.vendorStats); j++ {
			swap := false
			v1 := m.vendorStats[i]
			v2 := m.vendorStats[j]

			switch m.tableVendor.SortColumn {
			case 0: // Vendor
				val1 := strings.ToLower(getStringField(v1, "vendor"))
				val2 := strings.ToLower(getStringField(v2, "vendor"))
				if m.tableVendor.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // Count (stored as int, not float64)
				val1, ok1 := v1["count"].(int)
				val2, ok2 := v2["count"].(int)
				if !ok1 || !ok2 {
					continue
				}
				if m.tableVendor.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // Percentage (calculated, not stored)
				count1, ok1 := v1["count"].(int)
				count2, ok2 := v2["count"].(int)
				if !ok1 || !ok2 {
					continue
				}
				pct1 := float64(count1) / float64(totalDevices) * 100
				pct2 := float64(count2) / float64(totalDevices) * 100
				if m.tableVendor.SortAscending {
					swap = pct1 > pct2
				} else {
					swap = pct1 < pct2
				}
			}

			if swap {
				m.vendorStats[i], m.vendorStats[j] = m.vendorStats[j], m.vendorStats[i]
			}
		}
	}
}

func (m *ReportsModel) sortVLANs() {
	if len(m.vlans) == 0 {
		return
	}

	for i := 0; i < len(m.vlans)-1; i++ {
		for j := i + 1; j < len(m.vlans); j++ {
			swap := false
			v1 := m.vlans[i]
			v2 := m.vlans[j]

			switch m.tableVLANs.SortColumn {
			case 0: // VLAN ID
				val1, _ := v1["vlan"].(float64)
				val2, _ := v2["vlan"].(float64)
				if m.tableVLANs.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // Name (description)
				val1 := strings.ToLower(getStringField(v1, "description"))
				val2 := strings.ToLower(getStringField(v2, "description"))
				if m.tableVLANs.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // Description (same as name - might be redundant)
				val1 := strings.ToLower(getStringField(v1, "description"))
				val2 := strings.ToLower(getStringField(v2, "description"))
				if m.tableVLANs.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.vlans[i], m.vlans[j] = m.vlans[j], m.vlans[i]
			}
		}
	}
}

func (m *ReportsModel) View() string {
	// Report type tabs with emojis
	reportData := []struct{ emoji, name string }{
		{"🕐", "Recent"},
		{"📋", "All Devices"},
		{"🏢", "By Vendor"},
		{"🌐", "VLANs"},
	}
	var tabBar []string
	for i, data := range reportData {
		label := fmt.Sprintf("%s %s", data.emoji, data.name)
		if i == m.reportType {
			tabBar = append(tabBar, TabActiveStyle.Render(label))
		} else {
			tabBar = append(tabBar, TabInactiveStyle.Render(label))
		}
	}
	tabs := lipgloss.NewStyle().
		Padding(1, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, tabBar...))

	var header, controls, content string

	switch m.reportType {
	case ReportRecent:
		header = TitleStyle.Render("📋  Recently Added Devices") +
			SubtitleStyle.Render(fmt.Sprintf(" (last %d days)", m.days))
		controls = lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  +/- change range  ·  click header to sort  ·  r refresh  ·  ←→ switch report") + "\n"

		if m.loading {
			content = lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading..."
		} else if m.err != nil {
			content = ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error()))
		} else if len(m.devices) == 0 {
			content = WarningStyle.Render(fmt.Sprintf("No devices added in the last %d days. Press '+' to expand range.", m.days))
		} else {
			content = m.renderRecentDevicesTable()
		}

	case ReportAllDevices:
		header = TitleStyle.Render("📋  All Devices") +
			SubtitleStyle.Render(fmt.Sprintf(" (%d total)", len(m.devices)))
		controls = lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  click header to sort  ·  r refresh  ·  ←→ switch report") + "\n"

		if m.loading {
			content = lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading..."
		} else if m.err != nil {
			content = ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error()))
		} else {
			content = m.renderAllDevicesTable()
		}

	case ReportByVendor:
		header = TitleStyle.Render("📋  Devices by Vendor") +
			SubtitleStyle.Render(fmt.Sprintf(" (%d vendors)", len(m.vendorStats)))
		controls = lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  click header to sort  ·  r refresh  ·  ←→ switch report") + "\n"

		if m.loading {
			content = lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading..."
		} else if m.err != nil {
			content = ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error()))
		} else {
			content = m.renderVendorStatsTable()
		}

	case ReportVLANs:
		header = TitleStyle.Render("📋  VLAN Inventory") +
			SubtitleStyle.Render(fmt.Sprintf(" (%d VLANs)", len(m.vlans)))
		controls = lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  click header to sort  ·  r refresh  ·  ←→ switch report") + "\n"

		if m.loading {
			content = lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading..."
		} else if m.err != nil {
			content = ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error()))
		} else {
			content = m.renderVLANsTable()
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabs, "", header, controls, content)
}

func (m ReportsModel) renderRecentDevicesTable() string {
	headers := []string{"Name", "IP", "Vendor", "Model", "Location", "Discovered"}
	headerRow := m.tableRecent.RenderHeader(headers)
	sep := m.tableRecent.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.devices))

	for i := 0; i < end; i++ {
		d := m.devices[i]
		values := []string{
			orNA(shortName(getStringField(d, "device_name"))),
			orNA(getStringField(d, "ip")),
			orNA(getStringField(d, "vendor")),
			orNA(getStringField(d, "model")),
			orNA(getStringField(d, "location")),
			formatTime(getStringField(d, "creation")),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextMuted, ColorSuccess}

		row := m.tableRecent.RenderRow(values, colors)
		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate", 1, end, len(m.devices)))
	rows = append(rows, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m ReportsModel) renderAllDevicesTable() string {
	headers := []string{"Name", "IP", "Vendor", "Model", "Location"}
	headerRow := m.tableAllDevices.RenderHeader(headers)
	sep := m.tableAllDevices.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.devices))

	for i := 0; i < end; i++ {
		d := m.devices[i]
		values := []string{
			orNA(shortName(getStringField(d, "device_name"))),
			orNA(getStringField(d, "ip")),
			orNA(getStringField(d, "vendor")),
			orNA(getStringField(d, "model")),
			orNA(getStringField(d, "location")),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextMuted}

		row := m.tableAllDevices.RenderRow(values, colors)
		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate", 1, end, len(m.devices)))
	rows = append(rows, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m ReportsModel) renderVendorStatsTable() string {
	headers := []string{"Vendor", "Count", "Percentage"}
	headerRow := m.tableVendor.RenderHeader(headers)
	sep := m.tableVendor.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.vendorStats))

	totalDevices := len(m.devices)

	for i := 0; i < end; i++ {
		v := m.vendorStats[i]
		values := []string{
			orNA(getStringField(v, "vendor")),
			fmt.Sprintf("%d", v["count"].(int)),
			fmt.Sprintf("%.1f%%", float64(v["count"].(int))/float64(totalDevices)*100),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorSuccess}

		row := m.tableVendor.RenderRow(values, colors)
		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d vendors  ·  %d total devices  ·  ↑↓ navigate", len(m.vendorStats), totalDevices))
	rows = append(rows, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m ReportsModel) renderVLANsTable() string {
	headers := []string{"VLAN", "Name", "Description"}
	headerRow := m.tableVLANs.RenderHeader(headers)
	sep := m.tableVLANs.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.vlans))

	for i := 0; i < end; i++ {
		v := m.vlans[i]
		values := []string{
			orNA(getStringField(v, "vlan")),
			orNA(getStringField(v, "name")),
			orNA(getStringField(v, "description")),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted}

		row := m.tableVLANs.RenderRow(values, colors)
		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate", 1, end, len(m.vlans)))
	rows = append(rows, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
