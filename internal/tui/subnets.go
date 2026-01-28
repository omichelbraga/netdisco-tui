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
	loading       bool
	err           error
	searched      bool
	spinner       spinner.Model

	// IP inventory drill-down
	inIPInventory   bool
	selectedSubnet  string
	ipInventory     []map[string]interface{}
	ipLoading       bool
	ipErr           error
	ipCursor        int
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

	return SubnetsModel{
		width:       width,
		height:      height,
		client:      client,
		searchInput: ti,
		spinner:     s,
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
		return m, nil

	case ipInventoryMsg:
		m.ipLoading = false
		m.ipErr = msg.err
		m.ipInventory = msg.ips
		m.ipCursor = 0
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
				subnet := getStringField(m.subnets[m.cursor], "net")
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
			}
		case "down":
			if !m.searchInput.Focused() && m.cursor < len(m.subnets)-1 {
				m.cursor++
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

func (m SubnetsModel) View() string {
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
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d subnets  ·  Tab to list  ·  ↑↓ navigate  ·  Enter → IP inventory", len(m.subnets)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m SubnetsModel) renderSubnetsTable() string {
	colSubnet := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(20).Render("Subnet")
	colDesc := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(26).Render("Description")
	colTotal := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(8).Render("Total")
	colUsed := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(8).Render("Used")
	colFree := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(8).Render("Free")
	colUtil := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(10).Render("Util %")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colSubnet, colDesc, colTotal, colUsed, colFree, colUtil)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 20) + "┼" + strings.Repeat("─", 26) + "┼" +
			strings.Repeat("─", 8) + "┼" + strings.Repeat("─", 8) + "┼" +
			strings.Repeat("─", 8) + "┼" + strings.Repeat("─", 10))

	var rows []string
	rows = append(rows, headerRow, sep)

	for i, s := range m.subnets {
		subnet := orNA(getStringField(s, "net"))
		desc := truncate(orNA(getStringField(s, "description")), 24)
		total := orNA(getStringField(s, "total"))
		used := orNA(getStringField(s, "used"))
		free := orNA(getStringField(s, "free"))

		// Calculate utilization
		util := "N/A"
		totalF := getFloat(s, "total")
		usedF := getFloat(s, "used")
		if totalF > 0 {
			pct := (usedF / totalF) * 100
			util = fmt.Sprintf("%.1f%%", pct)
		}

		// Color util based on percentage
		utilColor := ColorSuccess
		if totalF > 0 {
			pct := (usedF / totalF) * 100
			if pct > 75 {
				utilColor = ColorDanger
			} else if pct > 50 {
				utilColor = ColorWarning
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(20).Foreground(ColorText).Render(subnet),
			lipgloss.NewStyle().Width(26).Foreground(ColorTextDim).Render(desc),
			lipgloss.NewStyle().Width(8).Foreground(ColorTextMuted).Render(total),
			lipgloss.NewStyle().Width(8).Foreground(ColorTextMuted).Render(used),
			lipgloss.NewStyle().Width(8).Foreground(ColorTextMuted).Render(free),
			lipgloss.NewStyle().Width(10).Foreground(utilColor).Render(util))

		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m SubnetsModel) viewIPInventory() string {
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

	colIP := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("IP")
	colMAC := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(20).Render("MAC")
	colDNS := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(24).Render("DNS")
	colVendor := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Vendor")
	colFirst := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("First Seen")
	colLast := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Last Seen")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colIP, colMAC, colDNS, colVendor, colFirst, colLast)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 16) + "┼" + strings.Repeat("─", 20) + "┼" +
			strings.Repeat("─", 24) + "┼" + strings.Repeat("─", 18) + "┼" +
			strings.Repeat("─", 18) + "┼" + strings.Repeat("─", 18))

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.ipInventory))

	for i := 0; i < end; i++ {
		ip := m.ipInventory[i]
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(16).Foreground(ColorText).Render(orNA(getStringField(ip, "ip"))),
			lipgloss.NewStyle().Width(20).Foreground(ColorTextDim).Render(truncate(orNA(getStringField(ip, "mac")), 18)),
			lipgloss.NewStyle().Width(24).Foreground(ColorTextMuted).Render(truncate(orNA(getStringField(ip, "dns")), 22)),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextMuted).Render(truncate(orNA(getStringField(ip, "oui")), 16)),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextMuted).Render(formatTime(getStringField(ip, "time_first"))),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextMuted).Render(formatTime(getStringField(ip, "time_last"))))

		if i == m.ipCursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d IPs  ·  ↑↓ navigate", len(m.ipInventory)))
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
