package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	inDetail      bool
	selectedNode  map[string]interface{}
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

	return NodesModel{
		width:       width,
		height:      height,
		client:      client,
		searchInput: ti,
		spinner:     s,
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

		// Get first IP
		if ips, ok := item["ips"].([]interface{}); ok && len(ips) > 0 {
			if ip, ok := ips[0].(map[string]interface{}); ok {
				row["mac"] = ip["mac"]
				row["ip"] = ip["ip"]
				row["dns"] = ip["dns"]
				row["time_last"] = ip["time_last"]
				if mfr, ok := ip["manufacturer"].(map[string]interface{}); ok {
					if company, ok := mfr["company"].(string); ok && company != "" {
						row["vendor"] = company
					} else if abbrev, ok := mfr["abbrev"].(string); ok {
						row["vendor"] = abbrev
					}
				}
			}
		}

		// Get first sighting
		if sightings, ok := item["sightings"].([]interface{}); ok && len(sightings) > 0 {
			if s, ok := sightings[0].(map[string]interface{}); ok {
				if dev, ok := s["device"].(map[string]interface{}); ok {
					row["switch_name"] = shortName(getStringField(dev, "name"))
				}
				row["switch_ip"] = s["switch"]
				row["port"] = s["port"]
				row["vlan"] = s["vlan"]
				if row["time_last"] == nil {
					row["time_last"] = s["time_last"]
				}
			}
		}

		flat = append(flat, row)
	}
	return flat
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

	table := m.renderNodesTable()
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d results  ·  Tab to list  ·  ↑↓ navigate  ·  Enter detail", len(m.flatResults)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m NodesModel) renderNodesTable() string {
	colMAC := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(20).Render("MAC")
	colIP := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("IP")
	colDNS := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(22).Render("DNS")
	colVendor := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("Vendor")
	colSwitch := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Switch")
	colPort := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(12).Render("Port")
	colVlan := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(8).Render("VLAN")
	colSeen := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Last Seen")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colMAC, colIP, colDNS, colVendor, colSwitch, colPort, colVlan, colSeen)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 20) + "┼" + strings.Repeat("─", 16) + "┼" +
			strings.Repeat("─", 22) + "┼" + strings.Repeat("─", 16) + "┼" +
			strings.Repeat("─", 18) + "┼" + strings.Repeat("─", 12) + "┼" +
			strings.Repeat("─", 8) + "┼" + strings.Repeat("─", 18))

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.flatResults))

	for i := 0; i < end; i++ {
		n := m.flatResults[i]
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(20).Foreground(ColorText).Render(truncate(orNA(getStringField(n, "mac")), 18)),
			lipgloss.NewStyle().Width(16).Foreground(ColorTextDim).Render(truncate(orNA(getStringField(n, "ip")), 14)),
			lipgloss.NewStyle().Width(22).Foreground(ColorTextMuted).Render(truncate(orNA(getStringField(n, "dns")), 20)),
			lipgloss.NewStyle().Width(16).Foreground(ColorTextMuted).Render(truncate(orNA(getStringField(n, "vendor")), 14)),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextDim).Render(truncate(orNA(getStringField(n, "switch_name")), 16)),
			lipgloss.NewStyle().Width(12).Foreground(ColorTextMuted).Render(truncate(orNA(getStringField(n, "port")), 10)),
			lipgloss.NewStyle().Width(8).Foreground(ColorTextMuted).Render(orNA(getStringField(n, "vlan"))),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextMuted).Render(formatTime(getStringField(n, "time_last"))))

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

	// IPs section
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("  IPs"))
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", 60)))

	if ips, ok := m.selectedNode["ips"].([]interface{}); ok {
		for _, ipRaw := range ips {
			if ip, ok := ipRaw.(map[string]interface{}); ok {
				lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
					"  IP: %-16s  DNS: %-24s  First: %-16s  Last: %s",
					orNA(getStringField(ip, "ip")),
					orNA(getStringField(ip, "dns")),
					formatTime(getStringField(ip, "time_first")),
					formatTime(getStringField(ip, "time_last")))))
			}
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  No IPs"))
	}

	lines = append(lines, "")

	// Sightings section
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("  Connections"))
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", 60)))

	if sightings, ok := m.selectedNode["sightings"].([]interface{}); ok {
		for _, sRaw := range sightings {
			if s, ok := sRaw.(map[string]interface{}); ok {
				switchName := ""
				if dev, ok := s["device"].(map[string]interface{}); ok {
					switchName = shortName(getStringField(dev, "name"))
				}
				lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
					"  Switch: %-16s  IP: %-16s  Port: %-12s  VLAN: %-8s  Last: %s",
					orNA(switchName),
					orNA(getStringField(s, "switch")),
					orNA(getStringField(s, "port")),
					orNA(getStringField(s, "vlan")),
					formatTime(getStringField(s, "time_last")))))
			}
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  No connections"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
