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

type vlansLoadMsg struct {
	vlans []map[string]interface{}
	err   error
}

type vlanSearchMsg struct {
	vlans []map[string]interface{}
	err   error
}

func loadVlansCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		vlans, err := client.GetVlanInventory()
		return vlansLoadMsg{vlans: vlans, err: err}
	}
}

func searchVlanCmd(client *api.Client, query string) tea.Cmd {
	return func() tea.Msg {
		vlans, err := client.SearchVlan(query)
		return vlanSearchMsg{vlans: vlans, err: err}
	}
}

type VlansModel struct {
	width, height int
	client        *api.Client
	vlans         []map[string]interface{}
	filtered      []map[string]interface{}
	cursor        int
	scrollOffset  int
	loading       bool
	err           error
	searching     bool
	searchInput   textinput.Model
	searchQuery   string
	spinner       spinner.Model
}

func NewVlansModel(width, height int, client *api.Client) VlansModel {
	ti := textinput.New()
	ti.Placeholder = "Search VLANs..."
	ti.CharLimit = 32
	ti.Width = width - 6

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return VlansModel{
		width:       width,
		height:      height,
		client:      client,
		loading:     true,
		spinner:     s,
		searchInput: ti,
	}
}

func (m VlansModel) Init() tea.Cmd {
	return loadVlansCmd(m.client)
}

func (m VlansModel) Update(msg tea.Msg) (VlansModel, tea.Cmd) {
	switch msg := msg.(type) {
	case vlansLoadMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.vlans = msg.vlans
			m.applyFilter()
		}
		return m, nil

	case vlanSearchMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.vlans = msg.vlans
			m.applyFilter()
		}
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.applyFilter()
			case "enter":
				m.searchQuery = m.searchInput.Value()
				m.searching = false
				if m.searchQuery != "" {
					m.loading = true
					m.err = nil
					return m, searchVlanCmd(m.client, m.searchQuery)
				}
				m.applyFilter()
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.searchQuery = m.searchInput.Value()
				m.applyFilter()
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "/":
			m.searching = true
			m.searchInput.Focus()
		case "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}
		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				maxVisible := m.height - 8
				if m.cursor >= m.scrollOffset+maxVisible {
					m.scrollOffset = m.cursor - maxVisible + 1
				}
			}
		case "r":
			m.loading = true
			m.err = nil
			return m, loadVlansCmd(m.client)
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *VlansModel) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.vlans
		return
	}
	query := strings.ToLower(m.searchQuery)
	var filtered []map[string]interface{}
	for _, v := range m.vlans {
		vlan := strings.ToLower(getStringField(v, "vlan"))
		desc := strings.ToLower(getStringField(v, "description"))
		if strings.Contains(vlan, query) || strings.Contains(desc, query) {
			filtered = append(filtered, v)
		}
	}
	m.filtered = filtered
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
		m.scrollOffset = 0
	}
}

func (m VlansModel) View() string {
	header := TitleStyle.Render("🌐  VLANs") +
		SubtitleStyle.Render(fmt.Sprintf(" (%d)", len(m.filtered)))

	var searchBar string
	if m.searching {
		searchBar = "\n" + SearchStyle.Render("🔍  "+m.searchInput.View()) + "\n"
	} else if m.searchQuery != "" {
		searchBar = "\n" + lipgloss.NewStyle().Foreground(ColorSecondary).Render(fmt.Sprintf("🔍  \"%s\"", m.searchQuery)) +
			lipgloss.NewStyle().Foreground(ColorTextMuted).Render(" (/ edit, Esc clear)") + "\n"
	} else {
		searchBar = "\n" + lipgloss.NewStyle().Foreground(ColorTextMuted).Render("Press / to search...") + "\n"
	}

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Loading VLANs...")
	}
	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error())))
	}
	if len(m.filtered) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			WarningStyle.Render("No VLANs found."))
	}

	table := m.renderVlansTable()
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate  ·  / search  ·  r refresh",
			m.scrollOffset+1, minInt(m.scrollOffset+m.height-8, len(m.filtered)), len(m.filtered)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m VlansModel) renderVlansTable() string {
	colVlan := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(10).Render("VLAN")
	colName := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(28).Render("Name")
	colDevice := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(24).Render("Device")
	colDeviceIP := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("Device IP")
	colUpdated := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Last Updated")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colVlan, colName, colDevice, colDeviceIP, colUpdated)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 28) + "┼" +
			strings.Repeat("─", 24) + "┼" + strings.Repeat("─", 16) + "┼" + strings.Repeat("─", 18))

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(m.scrollOffset+maxVisible, len(m.filtered))

	for i := m.scrollOffset; i < end; i++ {
		v := m.filtered[i]
		vlan := orNA(getStringField(v, "vlan"))
		name := truncate(orNA(getStringField(v, "description")), 26)
		device := truncate(orNA(shortName(getNestedString(v, "device", "name"))), 22)
		if device == "N/A" {
			device = orNA(getStringField(v, "ip"))
		}
		deviceIP := orNA(getStringField(v, "ip"))
		updated := formatTime(getStringField(v, "last_discover"))

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(10).Foreground(ColorText).Render(vlan),
			lipgloss.NewStyle().Width(28).Foreground(ColorTextDim).Render(name),
			lipgloss.NewStyle().Width(24).Foreground(ColorTextMuted).Render(device),
			lipgloss.NewStyle().Width(16).Foreground(ColorTextMuted).Render(deviceIP),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextMuted).Render(updated))

		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	if m.scrollOffset > 0 {
		rows = append([]string{lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ▲ more above")}, rows...)
	}
	if end < len(m.filtered) {
		rows = append(rows, lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ▼ more below"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
