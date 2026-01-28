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

	// Resizable table
	table ResizableTable
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
		table:       NewResizableTable([]int{10, 28, 24, 16, 18}), // VLAN, Name, Device, Device IP, Last Updated
	}
}

func (m VlansModel) Init() tea.Cmd {
	return loadVlansCmd(m.client)
}

func (m VlansModel) Update(msg tea.Msg) (VlansModel, tea.Cmd) {
	var cmds []tea.Cmd

	// Always update spinner
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	if spinnerCmd != nil {
		cmds = append(cmds, spinnerCmd)
	}

	switch msg := msg.(type) {
	case vlansLoadMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.vlans = msg.vlans
			if m.vlans == nil {
				m.vlans = []map[string]interface{}{}
			}
			m.applyFilter()
		}
		return m, tea.Batch(cmds...)

	case vlanSearchMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.vlans = msg.vlans
			if m.vlans == nil {
				m.vlans = []map[string]interface{}{}
			}
			m.applyFilter()
		}
		return m, tea.Batch(cmds...)

	case tea.MouseMsg:
		m.table.HandleMouse(msg, 0)
		return m, tea.Batch(cmds...)

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
					cmds = append(cmds, searchVlanCmd(m.client, m.searchQuery))
					return m, tea.Batch(cmds...)
				}
				m.applyFilter()
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.searchQuery = m.searchInput.Value()
				m.applyFilter()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			}
			return m, tea.Batch(cmds...)
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
			cmds = append(cmds, loadVlansCmd(m.client))
			return m, tea.Batch(cmds...)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *VlansModel) applyFilter() {
	// Ensure vlans is never nil
	if m.vlans == nil {
		m.vlans = []map[string]interface{}{}
	}

	if m.searchQuery == "" {
		m.filtered = m.vlans
	} else {
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
	}

	// Ensure filtered is never nil
	if m.filtered == nil {
		m.filtered = []map[string]interface{}{}
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = 0
		m.scrollOffset = 0
	}
}

func (m VlansModel) View() string {
	// Ensure filtered is initialized
	if m.filtered == nil {
		m.filtered = []map[string]interface{}{}
	}

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
		errMsg := "Unknown error"
		if m.err != nil {
			errMsg = m.err.Error()
		}
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", errMsg)))
	}
	if len(m.filtered) == 0 {
		msg := "No VLANs found."
		if !m.loading && m.err == nil && m.vlans == nil {
			msg = "Initializing... Press 'r' to load VLANs."
		}
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			WarningStyle.Render(msg))
	}

	table := m.renderVlansTable()
	
	// Ensure we have valid dimensions for footer
	visibleEnd := minInt(m.scrollOffset+m.height-8, len(m.filtered))
	if visibleEnd < 1 {
		visibleEnd = len(m.filtered)
	}
	
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate  ·  / search  ·  r refresh",
			m.scrollOffset+1, visibleEnd, len(m.filtered)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m VlansModel) renderVlansTable() string {
	headers := []string{"VLAN", "Name", "Device", "Device IP", "Last Updated"}
	headerRow := m.table.RenderHeader(headers)
	sep := m.table.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(m.scrollOffset+maxVisible, len(m.filtered))

	for i := m.scrollOffset; i < end; i++ {
		v := m.filtered[i]
		device := orNA(shortName(getNestedString(v, "device", "name")))
		if device == "N/A" {
			device = orNA(getStringField(v, "ip"))
		}

		values := []string{
			orNA(getStringField(v, "vlan")),
			orNA(getStringField(v, "description")),
			device,
			orNA(getStringField(v, "ip")),
			formatTime(getStringField(v, "last_discover")),
		}
		colors := []lipgloss.Color{ColorText, ColorTextDim, ColorTextMuted, ColorTextMuted, ColorTextMuted}

		row := m.table.RenderRow(values, colors)
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
