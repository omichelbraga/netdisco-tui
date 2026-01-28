package tui

import (
	"fmt"

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

	devices  DevicesModel
	nodes    NodesModel
	vlans    VlansModel
	subnets  SubnetsModel
	reports  ReportsModel
}

func NewRootModel(width, height int, client *api.Client) RootModel {
	return RootModel{
		width:       width,
		height:      height,
		client:      client,
		activeTab:   TabDevices,
		initialized: make([]bool, 5),
		devices:     NewDevicesModel(width, height-4, client),
		nodes:       NewNodesModel(width, height-4, client),
		vlans:       NewVlansModel(width, height-4, client),
		subnets:     NewSubnetsModel(width, height-4, client),
		reports:     NewReportsModel(width, height-4, client),
	}
}

func (m RootModel) Init() tea.Cmd {
	m.initialized[TabDevices] = true
	return m.devices.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case tea.KeyMsg:
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
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		// Only pass left/right to root if not in a detail view
		if msg.String() == "left" || msg.String() == "right" {
			passToChild := false
			switch m.activeTab {
			case TabDevices:
				passToChild = m.devices.inDetail
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
	// Tab bar
	var tabs []string
	for i, name := range tabNames {
		label := fmt.Sprintf("%d  %s", i+1, name)
		if i == m.activeTab {
			tabs = append(tabs, TabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, TabInactiveStyle.Render(label))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Content
	var content string
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

	// Footer
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  q quit  ·  1-5 tabs  ·  ←→ navigate tabs")

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, "", content, "", footer)
}
