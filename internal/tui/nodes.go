package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/netdisco-tui/netdisco-tui/internal/api"
)

type nodeSearchMsg struct {
	results []map[string]interface{}
	err     error
}

func searchNodeCmd(client *api.Client, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.SearchNode(query)
		return nodeSearchMsg{results: results, err: err}
	}
}

type NodesModel struct {
	width, height int
	client        *api.Client
	searchInput   textinput.Model
	results       []map[string]interface{}
	flatResults   []map[string]interface{} // flattened: one row per IP+sighting combo
	cursor        int
	loading       bool
	err           error
	searched      bool
	spinner       spinner.Model

	// Detail
	inDetail     bool
	selectedNode map[string]interface{}

	// Resizable table
	table ResizableTable
}

func NewNodesModel(width, height int, client *api.Client) NodesModel {
	ti := textinput.New()
	ti.Placeholder = "Enter IP or MAC address..."
	ti.CharLimit = 64
	ti.Width = width - 6
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	table := NewResizableTable([]int{18, 15, 25, 18, 16, 16, 5, 19}) // MAC, IP, DNS, Vendor, Switch, Port, VLAN, Last Seen
	table.SortColumn = 0                                              // Default sort by MAC
	table.SortAscending = true

	return NodesModel{
		width:       width,
		height:      height,
		client:      client,
		searchInput: ti,
		spinner:     s,
		table:       table,
	}
}

func (m NodesModel) Init() tea.Cmd {
	return nil
}

func (m NodesModel) Update(msg tea.Msg) (NodesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case nodeSearchMsg:
		m.loading = false
		m.searched = true
		m.err = msg.err
		m.results = msg.results
		m.flatResults = flattenNodeResults(msg.results)
		m.cursor = 0
		return m, nil

	case tea.MouseMsg:
		if !m.inDetail {
			// Check for header click (sorting)
			if col := m.table.HandleHeaderClick(msg, 0, 6); col >= 0 {
				if m.table.SortColumn == col {
					m.table.SortAscending = !m.table.SortAscending
				} else {
					m.table.SortColumn = col
					m.table.SortAscending = true
				}
				m.sortNodes()
				m.cursor = 0
				return m, nil
			}
			m.table.HandleMouse(msg, 0)
		}
		return m, nil

	case tea.KeyMsg:
		if m.inDetail {
			switch msg.String() {
			case "esc", "b":
				m.inDetail = false
			}
			return m, nil
		}

		switch msg.String() {
		case "enter":
			if !m.searchInput.Focused() {
				m.searchInput.Focus()
				return m, nil
			}
			query := m.searchInput.Value()
			if query != "" {
				m.loading = true
				m.err = nil
				return m, searchNodeCmd(m.client, query)
			}
		case "esc":
			m.searchInput.Blur()
			return m, nil
		case "up":
			if !m.searchInput.Focused() && m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if !m.searchInput.Focused() && m.cursor < len(m.flatResults)-1 {
				m.cursor++
			}
		case "tab":
			if m.searchInput.Focused() && len(m.flatResults) > 0 {
				m.searchInput.Blur()
			}
		}

		// If not focused on search, allow selection
		if !m.searchInput.Focused() && msg.String() == "enter" {
			if len(m.results) > 0 && m.cursor < len(m.results) {
				m.inDetail = true
				m.selectedNode = m.results[m.cursor]
			}
			return m, nil
		}

		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func flattenNodeResults(results []map[string]interface{}) []map[string]interface{} {
	var flat []map[string]interface{}
	for _, item := range results {
		row := map[string]interface{}{}

		// Check if this is a full MAC search result (has "sightings", "ips" sections)
		if sightings, hasSightings := item["sightings"].([]interface{}); hasSightings {
			// Extract from sightings (port/VLAN location)
			if len(sightings) > 0 {
				if s, ok := sightings[0].(map[string]interface{}); ok {
					row["mac"] = getStringField(s, "mac")
					row["port"] = getStringField(s, "port")
					row["vlan"] = getStringField(s, "vlan")
					row["switch_ip"] = getStringField(s, "switch")
					row["time_last"] = getStringField(s, "time_last")
					if dev, ok := s["device"].(map[string]interface{}); ok {
						row["switch_name"] = shortName(getStringField(dev, "name"))
					}
				}
			}

			// Extract IP/DNS from ips section
			if ips, ok := item["ips"].([]interface{}); ok && len(ips) > 0 {
				for _, ipRaw := range ips {
					if ip, ok := ipRaw.(map[string]interface{}); ok {
						// Prefer non-IPv6 addresses
						ipAddr := getStringField(ip, "ip")
						if ipAddr != "" && !strings.Contains(ipAddr, ":") {
							row["ip"] = ipAddr
							row["dns"] = getStringField(ip, "dns")
							if mfr, ok := ip["manufacturer"].(map[string]interface{}); ok {
								if company, ok := mfr["company"].(string); ok && company != "" {
									row["vendor"] = company
								} else if abbrev, ok := mfr["abbrev"].(string); ok {
									row["vendor"] = abbrev
								}
							}
							break
						}
					}
				}
			}
		} else {
			// Simple IP search result (no sightings)
			row["mac"] = getStringField(item, "mac")
			row["ip"] = getStringField(item, "ip")
			row["dns"] = getStringField(item, "dns")
			row["time_last"] = getStringField(item, "time_last")

			if mfr, ok := item["manufacturer"].(map[string]interface{}); ok {
				if company, ok := mfr["company"].(string); ok && company != "" {
					row["vendor"] = company
				} else if abbrev, ok := mfr["abbrev"].(string); ok {
					row["vendor"] = abbrev
				}
			}

			row["switch_name"] = shortName(getStringField(item, "router_name"))
			row["switch_ip"] = getStringField(item, "router_ip")
			row["port"] = "N/A"
			row["vlan"] = "N/A"
		}

		flat = append(flat, row)
	}
	return flat
}

func (m *NodesModel) sortNodes() {
	if len(m.flatResults) == 0 {
		return
	}

	for i := 0; i < len(m.flatResults)-1; i++ {
		for j := i + 1; j < len(m.flatResults); j++ {
			swap := false
			n1 := m.flatResults[i]
			n2 := m.flatResults[j]

			switch m.table.SortColumn {
			case 0: // MAC
				val1 := strings.ToLower(getStringField(n1, "mac"))
				val2 := strings.ToLower(getStringField(n2, "mac"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // IP
				val1 := strings.ToLower(getStringField(n1, "ip"))
				val2 := strings.ToLower(getStringField(n2, "ip"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // DNS
				val1 := strings.ToLower(getStringField(n1, "dns"))
				val2 := strings.ToLower(getStringField(n2, "dns"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 3: // Vendor
				val1 := strings.ToLower(getStringField(n1, "vendor"))
				val2 := strings.ToLower(getStringField(n2, "vendor"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 4: // Switch
				val1 := strings.ToLower(getStringField(n1, "switch_name"))
				val2 := strings.ToLower(getStringField(n2, "switch_name"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 5: // Port
				val1 := strings.ToLower(getStringField(n1, "port"))
				val2 := strings.ToLower(getStringField(n2, "port"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 6: // VLAN
				val1 := strings.ToLower(getStringField(n1, "vlan"))
				val2 := strings.ToLower(getStringField(n2, "vlan"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 7: // Last Seen
				val1 := strings.ToLower(getStringField(n1, "time_last"))
				val2 := strings.ToLower(getStringField(n2, "time_last"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.flatResults[i], m.flatResults[j] = m.flatResults[j], m.flatResults[i]
			}
		}
	}
}

func (m NodesModel) View() string {
	if m.inDetail {
		return m.viewDetail()
	}

	header := TitleStyle.Render("🔍  Node Lookup")
	searchBar := "\n" + SearchStyle.Render("🔍  "+m.searchInput.View()) + "\n"

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Searching...")
	}

	if !m.searched {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			lipgloss.NewStyle().Foreground(ColorTextMuted).Render("Enter an IP or MAC address and press Enter to search."))
	}

	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s", m.err.Error())))
	}

	if len(m.flatResults) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			WarningStyle.Render("No results found."))
	}

	// Set HeaderY for click detection
	m.table.HeaderY = 7

	table := m.renderNodesTable()
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d results  ·  ↑↓ navigate  ·  click header to sort  ·  Enter detail", len(m.flatResults)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m NodesModel) renderNodesTable() string {
	headers := []string{"MAC", "IP", "DNS", "Vendor", "Switch", "Port", "VLAN", "Last Seen"}
	headerRow := m.table.RenderHeader(headers)
	sep := m.table.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.flatResults))

	for i := 0; i < end; i++ {
		n := m.flatResults[i]
		values := []string{
			orNA(getStringField(n, "mac")),
			orNA(getStringField(n, "ip")),
			orNA(getStringField(n, "dns")),
			orNA(getStringField(n, "vendor")),
			orNA(getStringField(n, "switch_name")),
			orNA(getStringField(n, "port")),
			orNA(getStringField(n, "vlan")),
			formatTime(getStringField(n, "time_last")),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextMuted}

		row := m.table.RenderRow(values, colors)
		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m NodesModel) viewDetail() string {
	header := TitleStyle.Render("🔍  Node Detail") +
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ← Esc/b")

	if m.selectedNode == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", WarningStyle.Render("No node selected"))
	}

	var lines []string
	lines = append(lines, header, "")

	// Extract basic info from IPs section (if present)
	mac := getStringField(m.selectedNode, "mac")
	ip := ""
	dns := ""
	vendor := "N/A"
	timeFirst := ""
	timeLast := ""

	if ips, ok := m.selectedNode["ips"].([]interface{}); ok && len(ips) > 0 {
		for _, ipRaw := range ips {
			if ipData, ok := ipRaw.(map[string]interface{}); ok {
				ipAddr := getStringField(ipData, "ip")
				if ipAddr != "" && !strings.Contains(ipAddr, ":") {
					ip = ipAddr
					dns = getStringField(ipData, "dns")
					timeFirst = getStringField(ipData, "time_first")
					timeLast = getStringField(ipData, "time_last")
					if mac == "" {
						mac = getStringField(ipData, "mac")
					}
					if mfr, ok := ipData["manufacturer"].(map[string]interface{}); ok {
						if company, ok := mfr["company"].(string); ok && company != "" {
							vendor = company
						} else if abbrev, ok := mfr["abbrev"].(string); ok {
							vendor = abbrev
						}
					}
					break
				}
			}
		}
	}

	// Basic Info section
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("  Node Information"))
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", 110)))

	lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
		"  MAC: %-20s  IP: %-16s  DNS: %s",
		orNA(mac),
		orNA(ip),
		orNA(dns))))

	lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
		"  Vendor: %s",
		vendor)))

	if timeFirst != "" || timeLast != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
			"  First Seen: %-20s  Last Seen: %s",
			formatTime(timeFirst),
			formatTime(timeLast))))
	}

	lines = append(lines, "")

	// Sightings/Connections section
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("  Connections"))
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", 110)))

	if sightings, ok := m.selectedNode["sightings"].([]interface{}); ok && len(sightings) > 0 {
		for _, sRaw := range sightings {
			if s, ok := sRaw.(map[string]interface{}); ok {
				switchName := "N/A"
				if dev, ok := s["device"].(map[string]interface{}); ok {
					switchName = shortName(getStringField(dev, "name"))
				}
				lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
					"  Switch: %-20s  IP: %-16s  Port: %-18s  VLAN: %-6s  Last: %s",
					orNA(switchName),
					orNA(getStringField(s, "switch")),
					orNA(getStringField(s, "port")),
					orNA(getStringField(s, "vlan")),
					formatTime(getStringField(s, "time_last")))))
			}
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  No connection information"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
