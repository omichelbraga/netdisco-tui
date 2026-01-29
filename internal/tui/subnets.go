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

type subnetUtilMsg struct {
	subnets []map[string]interface{}
	err     error
}

type ipInventoryMsg struct {
	ips []map[string]interface{}
	err error
}

func loadSubnetUtilCmd(client *api.Client, subnet string) tea.Cmd {
	return func() tea.Msg {
		subnets, err := client.GetSubnetUtilization(subnet)
		return subnetUtilMsg{subnets: subnets, err: err}
	}
}

func loadIpInventoryCmd(client *api.Client, subnet string) tea.Cmd {
	return func() tea.Msg {
		ips, err := client.GetIpInventory(subnet, 256)
		return ipInventoryMsg{ips: ips, err: err}
	}
}

type SubnetsModel struct {
	width, height int
	client        *api.Client
	searchInput   textinput.Model
	subnets       []map[string]interface{}
	cursor        int
	scrollOffset  int
	loading       bool
	err           error
	searched      bool
	spinner       spinner.Model

	// IP inventory drill-down
	inIPInventory  bool
	selectedSubnet string
	ipInventory    []map[string]interface{}
	ipLoading      bool
	ipErr          error
	ipCursor       int

	// Resizable tables
	table   ResizableTable
	ipTable ResizableTable
}

func NewSubnetsModel(width, height int, client *api.Client) SubnetsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter CIDR (e.g. 10.0.0.0/24)..."
	ti.CharLimit = 20
	ti.Width = width - 6
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	table := NewResizableTable([]int{20, 26, 8, 8, 8, 10}) // Subnet, Description, Total, Used, Free, Util %
	table.SortColumn = 0                                   // Default sort by Subnet
	table.SortAscending = true

	ipTable := NewResizableTable([]int{16, 20, 24, 18, 18, 18}) // IP, MAC, DNS, Vendor, First Seen, Last Seen
	ipTable.SortColumn = 0                                       // Default sort by IP
	ipTable.SortAscending = true

	return SubnetsModel{
		width:       width,
		height:      height,
		client:      client,
		searchInput: ti,
		spinner:     s,
		table:       table,
		ipTable:     ipTable,
	}
}

func (m SubnetsModel) Init() tea.Cmd {
	return nil
}

func (m SubnetsModel) Update(msg tea.Msg) (SubnetsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case subnetUtilMsg:
		m.loading = false
		m.searched = true
		m.err = msg.err
		m.subnets = msg.subnets
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil

	case ipInventoryMsg:
		m.ipLoading = false
		m.ipErr = msg.err
		m.ipInventory = msg.ips
		m.ipCursor = 0
		return m, nil

	case tea.MouseMsg:
		// Calculate HeaderY dynamically based on rendered components
		// Root UI: 4 lines (header bar + tab bar)
		// Title: 1 line
		// Search bar: "\n" + input + "\n" = 3 lines
		// Table header is right after
		rootUIHeight := 4
		titleHeight := 1
		searchBarHeight := 3
		m.table.HeaderY = rootUIHeight + titleHeight + searchBarHeight
		
		// IP inventory has different layout
		// Root UI: 4 lines
		// Title with back info: 1 line
		// Blank line: 1
		// Table header
		m.ipTable.HeaderY = 4 + 1 + 1
		
		if m.inIPInventory {
			// Check for IP table header click
			if col := m.ipTable.HandleHeaderClick(msg, 0, 3); col >= 0 {
				if m.ipTable.SortColumn == col {
					m.ipTable.SortAscending = !m.ipTable.SortAscending
				} else {
					m.ipTable.SortColumn = col
					m.ipTable.SortAscending = true
				}
				m.sortIPInventory()
				m.ipCursor = 0
				return m, nil
			}
			m.ipTable.HandleMouse(msg, 0)
		} else {
			// Check for subnet table header click (wider tolerance)
			if col := m.table.HandleHeaderClick(msg, 0, 6); col >= 0 {
				if m.table.SortColumn == col {
					m.table.SortAscending = !m.table.SortAscending
				} else {
					m.table.SortColumn = col
					m.table.SortAscending = true
				}
				m.sortSubnets()
				m.cursor = 0
				m.scrollOffset = 0
				return m, nil
			}
			m.table.HandleMouse(msg, 0)
		}
		return m, nil

	case tea.KeyMsg:
		if m.inIPInventory {
			switch msg.String() {
			case "esc", "b":
				m.inIPInventory = false
			case "up":
				if m.ipCursor > 0 {
					m.ipCursor--
				}
			case "down":
				if m.ipCursor < len(m.ipInventory)-1 {
					m.ipCursor++
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "enter":
			if m.searchInput.Focused() {
				subnet := m.searchInput.Value()
				if subnet != "" {
					m.loading = true
					m.err = nil
					m.searchInput.Blur()
					return m, loadSubnetUtilCmd(m.client, subnet)
				}
			} else if len(m.subnets) > 0 && m.cursor < len(m.subnets) {
				subnet := getStringField(m.subnets[m.cursor], "subnet")
				if subnet != "" {
					m.inIPInventory = true
					m.selectedSubnet = subnet
					m.ipLoading = true
					return m, loadIpInventoryCmd(m.client, subnet)
				}
			}
		case "esc":
			m.searchInput.Blur()
		case "tab":
			if m.searchInput.Focused() && len(m.subnets) > 0 {
				m.searchInput.Blur()
			}
		case "up":
			if !m.searchInput.Focused() && m.cursor > 0 {
				m.cursor--
				// Adjust scroll to keep cursor visible
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}
		case "down":
			if !m.searchInput.Focused() && m.cursor < len(m.subnets)-1 {
				m.cursor++
				// Adjust scroll to keep cursor visible
				maxVisible := m.height - 10
				if maxVisible < 5 {
					maxVisible = 5
				}
				if m.cursor >= m.scrollOffset+maxVisible {
					m.scrollOffset = m.cursor - maxVisible + 1
				}
			}
		case "pgup", "pageup":
			if !m.searchInput.Focused() {
				maxVisible := m.height - 10
				if maxVisible < 5 {
					maxVisible = 5
				}
				m.cursor -= maxVisible
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.scrollOffset = m.cursor
			}
		case "pgdown", "pagedown":
			if !m.searchInput.Focused() {
				maxVisible := m.height - 10
				if maxVisible < 5 {
					maxVisible = 5
				}
				m.cursor += maxVisible
				if m.cursor >= len(m.subnets) {
					m.cursor = len(m.subnets) - 1
				}
				if m.cursor >= m.scrollOffset+maxVisible {
					m.scrollOffset = m.cursor - maxVisible + 1
				}
			}
		case "home":
			if !m.searchInput.Focused() && len(m.subnets) > 0 {
				m.cursor = 0
				m.scrollOffset = 0
			}
		case "end":
			if !m.searchInput.Focused() && len(m.subnets) > 0 {
				m.cursor = len(m.subnets) - 1
				maxVisible := m.height - 10
				if maxVisible < 5 {
					maxVisible = 5
				}
				m.scrollOffset = m.cursor - maxVisible + 1
				if m.scrollOffset < 0 {
					m.scrollOffset = 0
				}
			}
		case "r":
			if m.searched {
				m.loading = true
				m.err = nil
				return m, loadSubnetUtilCmd(m.client, m.searchInput.Value())
			}
		}

		if !m.inIPInventory {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *SubnetsModel) sortSubnets() {
	if len(m.subnets) == 0 {
		return
	}

	for i := 0; i < len(m.subnets)-1; i++ {
		for j := i + 1; j < len(m.subnets); j++ {
			swap := false
			s1 := m.subnets[i]
			s2 := m.subnets[j]

			switch m.table.SortColumn {
			case 0: // Subnet
				val1 := strings.ToLower(getStringField(s1, "subnet"))
				val2 := strings.ToLower(getStringField(s2, "subnet"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // Description
				val1 := strings.ToLower(getStringField(s1, "description"))
				val2 := strings.ToLower(getStringField(s2, "description"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // Total
				val1, _ := s1["subnet_size"].(float64)
				val2, _ := s2["subnet_size"].(float64)
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 3: // Used
				val1, _ := s1["active"].(float64)
				val2, _ := s2["active"].(float64)
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 4: // Free (calculated: subnet_size - active)
				total1, _ := s1["subnet_size"].(float64)
				used1, _ := s1["active"].(float64)
				val1 := total1 - used1
				total2, _ := s2["subnet_size"].(float64)
				used2, _ := s2["active"].(float64)
				val2 := total2 - used2
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 5: // Util %
				val1, _ := s1["percent"].(float64)
				val2, _ := s2["percent"].(float64)
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.subnets[i], m.subnets[j] = m.subnets[j], m.subnets[i]
			}
		}
	}
}

func (m *SubnetsModel) sortIPInventory() {
	if len(m.ipInventory) == 0 {
		return
	}

	for i := 0; i < len(m.ipInventory)-1; i++ {
		for j := i + 1; j < len(m.ipInventory); j++ {
			swap := false
			ip1 := m.ipInventory[i]
			ip2 := m.ipInventory[j]

			switch m.ipTable.SortColumn {
			case 0: // IP
				val1 := strings.ToLower(getStringField(ip1, "ip"))
				val2 := strings.ToLower(getStringField(ip2, "ip"))
				if m.ipTable.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // MAC
				val1 := strings.ToLower(getStringField(ip1, "mac"))
				val2 := strings.ToLower(getStringField(ip2, "mac"))
				if m.ipTable.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // DNS
				val1 := strings.ToLower(getStringField(ip1, "dns"))
				val2 := strings.ToLower(getStringField(ip2, "dns"))
				if m.ipTable.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 3: // Vendor
				val1 := strings.ToLower(getStringField(ip1, "vendor"))
				val2 := strings.ToLower(getStringField(ip2, "vendor"))
				if m.ipTable.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 4: // First Seen
				val1 := strings.ToLower(getStringField(ip1, "time_first"))
				val2 := strings.ToLower(getStringField(ip2, "time_first"))
				if m.ipTable.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 5: // Last Seen
				val1 := strings.ToLower(getStringField(ip1, "time_last"))
				val2 := strings.ToLower(getStringField(ip2, "time_last"))
				if m.ipTable.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.ipInventory[i], m.ipInventory[j] = m.ipInventory[j], m.ipInventory[i]
			}
		}
	}
}

func (m *SubnetsModel) View() string {
	if m.inIPInventory {
		return m.viewIPInventory()
	}

	header := TitleStyle.Render("📊  Subnet Utilization")
	searchBar := "\n" + SearchStyle.Render("🔍  "+m.searchInput.View()) + "\n"

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Loading...")
	}

	if !m.searched {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			lipgloss.NewStyle().Foreground(ColorTextMuted).Render("Enter a CIDR subnet and press Enter."))
	}

	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error())))
	}

	if len(m.subnets) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			WarningStyle.Render("No subnet data found."))
	}

	table := m.renderSubnetsTable()

	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	visibleEnd := minInt(m.scrollOffset+maxVisible, len(m.subnets))

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate  ·  click header to sort  ·  Enter → IP inventory",
			m.scrollOffset+1, visibleEnd, len(m.subnets)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m SubnetsModel) renderSubnetsTable() string {
	headers := []string{"Subnet", "Description", "Total", "Used", "Free", "Util %"}
	headerRow := m.table.RenderHeader(headers)
	sep := m.table.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	// Calculate visible range
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(m.scrollOffset+maxVisible, len(m.subnets))

	for i := m.scrollOffset; i < end; i++ {
		s := m.subnets[i]
		// API returns: subnet_size (total), active (used), percent (util %)
		totalF := getFloat(s, "subnet_size")
		usedF := getFloat(s, "active")
		freeF := totalF - usedF
		pct := getFloat(s, "percent")

		// Color util based on percentage
		utilColor := ColorSuccess
		if pct > 75 {
			utilColor = ColorDanger
		} else if pct > 50 {
			utilColor = ColorWarning
		}

		values := []string{
			orNA(getStringField(s, "subnet")),
			orNA(getStringField(s, "description")),
			fmt.Sprintf("%d", int(totalF)),
			fmt.Sprintf("%d", int(usedF)),
			fmt.Sprintf("%d", int(freeF)),
			fmt.Sprintf("%.1f%%", pct),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextMuted, utilColor}

		row := m.table.RenderRow(values, colors)
		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	// Add scroll indicators
	if m.scrollOffset > 0 {
		rows = append([]string{lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ▲ more above")}, rows...)
	}
	if end < len(m.subnets) {
		rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ▼ more below"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *SubnetsModel) viewIPInventory() string {
	header := TitleStyle.Render("📊  IP Inventory") +
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render(fmt.Sprintf("  [%s]", m.selectedSubnet)) +
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ← Esc/b")

	if m.ipLoading {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Loading IPs...")
	}
	if m.ipErr != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s", m.ipErr.Error())))
	}
	if len(m.ipInventory) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			WarningStyle.Render("No IPs found in this subnet."))
	}

	headers := []string{"IP", "MAC", "DNS", "Vendor", "First Seen", "Last Seen"}
	headerRow := m.ipTable.RenderHeader(headers)
	sep := m.ipTable.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.ipInventory))

	for i := 0; i < end; i++ {
		ip := m.ipInventory[i]
		values := []string{
			orNA(getStringField(ip, "ip")),
			orNA(getStringField(ip, "mac")),
			orNA(getStringField(ip, "dns")),
			orNA(getStringField(ip, "oui")),
			formatTime(getStringField(ip, "time_first")),
			formatTime(getStringField(ip, "time_last")),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextMuted, ColorTextMuted}

		row := m.ipTable.RenderRow(values, colors)
		if i == m.ipCursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d IPs  ·  ↑↓ navigate  ·  click header to sort", len(m.ipInventory)))
	rows = append(rows, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{header, ""}, rows...)...)
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}
