package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/netdisco-tui/netdisco-tui/internal/api"
)

type recentDevicesMsg struct {
	devices []map[string]interface{}
	err     error
}

func loadRecentDevicesCmd(client *api.Client, days int) tea.Cmd {
	return func() tea.Msg {
		devices, err := client.GetRecentDevices(days)
		return recentDevicesMsg{devices: devices, err: err}
	}
}

type ReportsModel struct {
	width, height int
	client        *api.Client
	devices       []map[string]interface{}
	cursor        int
	loading       bool
	err           error
	spinner       spinner.Model
	days          int
}

func NewReportsModel(width, height int, client *api.Client) ReportsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return ReportsModel{
		width:   width,
		height:  height,
		client:  client,
		loading: true,
		spinner: s,
		days:    7,
	}
}

func (m ReportsModel) Init() tea.Cmd {
	return loadRecentDevicesCmd(m.client, m.days)
}

func (m ReportsModel) Update(msg tea.Msg) (ReportsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case recentDevicesMsg:
		m.loading = false
		m.err = msg.err
		m.devices = msg.devices
		m.cursor = 0
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.devices)-1 {
				m.cursor++
			}
		case "+":
			m.days += 7
			m.loading = true
			m.err = nil
			return m, loadRecentDevicesCmd(m.client, m.days)
		case "-":
			if m.days > 1 {
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
			return m, loadRecentDevicesCmd(m.client, m.days)
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ReportsModel) View() string {
	header := TitleStyle.Render("📋  Recently Added Devices") +
		SubtitleStyle.Render(fmt.Sprintf(" (last %d days)", m.days))

	controls := lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  +/- change range  ·  r refresh") + "\n"

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, header, controls,
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Loading...")
	}
	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, controls,
			ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error())))
	}
	if len(m.devices) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, controls,
			WarningStyle.Render(fmt.Sprintf("No devices added in the last %d days. Press '+' to expand range.", m.days)))
	}

	table := m.renderReportsTable()
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d devices  ·  ↑↓ navigate", len(m.devices)))

	return lipgloss.JoinVertical(lipgloss.Left, header, controls, table, "", footer)
}

func (m ReportsModel) renderReportsTable() string {
	colName := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(24).Render("Name")
	colIP := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("IP")
	colVendor := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(14).Render("Vendor")
	colModel := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("Model")
	colOS := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(10).Render("OS")
	colLocation := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Location")
	colDiscovered := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(18).Render("Discovered")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colName, colIP, colVendor, colModel, colOS, colLocation, colDiscovered)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 24) + "┼" + strings.Repeat("─", 16) + "┼" +
			strings.Repeat("─", 14) + "┼" + strings.Repeat("─", 16) + "┼" +
			strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 18) + "┼" + strings.Repeat("─", 18))

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.devices))

	for i := 0; i < end; i++ {
		d := m.devices[i]
		name := truncate(orNA(shortName(getStringField(d, "name"))), 22)
		ip := orNA(getStringField(d, "ip"))
		vendor := truncate(orNA(getStringField(d, "vendor")), 12)
		model := truncate(orNA(getStringField(d, "model")), 14)
		os := truncate(orNA(getStringField(d, "os")), 8)
		location := truncate(orNA(getStringField(d, "location")), 16)
		discovered := formatTime(getStringField(d, "creation"))
		if discovered == "N/A" {
			discovered = formatTime(getStringField(d, "first_seen"))
		}
		if discovered == "N/A" {
			discovered = formatTime(getStringField(d, "last_discover"))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(24).Foreground(ColorText).Render(name),
			lipgloss.NewStyle().Width(16).Foreground(ColorTextDim).Render(ip),
			lipgloss.NewStyle().Width(14).Foreground(ColorTextMuted).Render(vendor),
			lipgloss.NewStyle().Width(16).Foreground(ColorTextMuted).Render(model),
			lipgloss.NewStyle().Width(10).Foreground(ColorTextMuted).Render(os),
			lipgloss.NewStyle().Width(18).Foreground(ColorTextMuted).Render(location),
			lipgloss.NewStyle().Width(18).Foreground(ColorSuccess).Render(discovered))

		if i == m.cursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
