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

// --- Messages ---

type devicesLoadMsg struct {
	devices []map[string]interface{}
	err     error
}

type deviceDetailLoadMsg struct {
	device map[string]interface{}
	err    error
}

type devicePortsLoadMsg struct {
	ports []map[string]interface{}
	err   error
}

type deviceNeighborsLoadMsg struct {
	neighbors map[string]interface{}
	err       error
}

type deviceVlansLoadMsg struct {
	vlans []map[string]interface{}
	err   error
}

type portNodesLoadMsg struct {
	nodes []map[string]interface{}
	err   error
}

// --- Commands ---

func loadDevicesCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		devices, err := client.GetDeviceInventory()
		return devicesLoadMsg{devices: devices, err: err}
	}
}

func searchDevicesCmd(client *api.Client, query string) tea.Cmd {
	return func() tea.Msg {
		devices, err := client.SearchDevice(query, "", "", "")
		return devicesLoadMsg{devices: devices, err: err}
	}
}

func loadDeviceDetailCmd(client *api.Client, ip string) tea.Cmd {
	return func() tea.Msg {
		device, err := client.GetDevice(ip)
		return deviceDetailLoadMsg{device: device, err: err}
	}
}

func loadDevicePortsCmd(client *api.Client, ip string) tea.Cmd {
	return func() tea.Msg {
		ports, err := client.GetDevicePorts(ip)
		return devicePortsLoadMsg{ports: ports, err: err}
	}
}

func loadDeviceNeighborsCmd(client *api.Client, ip string) tea.Cmd {
	return func() tea.Msg {
		neighbors, err := client.GetDeviceNeighbors(ip, 1)
		return deviceNeighborsLoadMsg{neighbors: neighbors, err: err}
	}
}

func loadDeviceVlansCmd(client *api.Client, ip string) tea.Cmd {
	return func() tea.Msg {
		vlans, err := client.GetDeviceVlans(ip)
		return deviceVlansLoadMsg{vlans: vlans, err: err}
	}
}

func loadPortNodesCmd(client *api.Client, ip, port string) tea.Cmd {
	return func() tea.Msg {
		nodes, err := client.GetPortActiveNodes(ip, port)
		return portNodesLoadMsg{nodes: nodes, err: err}
	}
}

// --- DevicesModel ---

type DevicesModel struct {
	width, height int
	client        *api.Client
	devices       []map[string]interface{}
	filtered      []map[string]interface{}
	cursor        int
	scrollOffset  int
	loading       bool
	err           error
	searching     bool
	searchInput   textinput.Model
	searchQuery   string
	spinner       spinner.Model

	// Detail state
	inDetail         bool
	selectedIP       string
	detailTab        int // 0=Info 1=Ports 2=Neighbors 3=VLANs
	detailDevice     map[string]interface{}
	detailLoading    bool
	detailErr        error
	ports            []map[string]interface{}
	portsLoading     bool
	portsErr         error
	neighbors        map[string]interface{}
	neighborsLoading bool
	neighborsErr     error
	vlans            []map[string]interface{}
	vlansLoading     bool
	vlansErr         error

	// Port nodes
	portsCursor      int
	showPortNodes    bool
	portNodesIP      string
	portNodesPort    string
	portNodes        []map[string]interface{}
	portNodesLoading bool
	portNodesErr     error

	// Resizable table
	table ResizableTable
}

func NewDevicesModel(width, height int, client *api.Client) DevicesModel {
	ti := textinput.New()
	ti.Placeholder = "Search devices..."
	ti.CharLimit = 64
	ti.Width = width - 6

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	table := NewResizableTable([]int{22, 15, 14, 24, 18}) // Name, IP, Vendor, Model, Location
	table.SortColumn = 0                                   // Default sort by Name
	table.SortAscending = true

	return DevicesModel{
		width:       width,
		height:      height,
		client:      client,
		loading:     true,
		spinner:     s,
		searchInput: ti,
		table:       table,
	}
}

func (m DevicesModel) Init() tea.Cmd {
	return loadDevicesCmd(m.client)
}

func (m DevicesModel) Update(msg tea.Msg) (DevicesModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case devicesLoadMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.devices = msg.devices
			m.applyFilter()
		}
		return m, nil

	case deviceDetailLoadMsg:
		m.detailLoading = false
		m.detailErr = msg.err
		m.detailDevice = msg.device
		return m, nil

	case devicePortsLoadMsg:
		m.portsLoading = false
		m.portsErr = msg.err
		m.ports = msg.ports
		m.portsCursor = 0
		return m, nil

	case deviceNeighborsLoadMsg:
		m.neighborsLoading = false
		m.neighborsErr = msg.err
		m.neighbors = msg.neighbors
		return m, nil

	case deviceVlansLoadMsg:
		m.vlansLoading = false
		m.vlansErr = msg.err
		m.vlans = msg.vlans
		return m, nil

	case portNodesLoadMsg:
		m.portNodesLoading = false
		m.portNodesErr = msg.err
		m.portNodes = msg.nodes
		return m, nil

	case tea.MouseMsg:
		if !m.inDetail {
			// Check for header click (sorting)
			if col := m.table.HandleHeaderClick(msg, 0, 2); col >= 0 {
				// Clicked on column header
				if m.table.SortColumn == col {
					// Toggle sort direction
					m.table.SortAscending = !m.table.SortAscending
				} else {
					// New column, default to ascending
					m.table.SortColumn = col
					m.table.SortAscending = true
				}
				m.applyFilter()
				return m, nil
			}
			m.table.HandleMouse(msg, 0)
		}
		return m, nil

	case tea.KeyMsg:
		if m.inDetail {
			return m.updateDetail(msg)
		}
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateList(msg)

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m DevicesModel) updateSearch(msg tea.KeyMsg) (DevicesModel, tea.Cmd) {
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
			return m, searchDevicesCmd(m.client, m.searchQuery)
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

func (m DevicesModel) updateList(msg tea.KeyMsg) (DevicesModel, tea.Cmd) {
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
	case "enter":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			ip := getStringField(m.filtered[m.cursor], "ip")
			if ip != "" {
				m.inDetail = true
				m.selectedIP = ip
				m.detailTab = 0
				m.detailLoading = true
				m.showPortNodes = false
				return m, loadDeviceDetailCmd(m.client, ip)
			}
		}
	case "r":
		m.loading = true
		m.err = nil
		if m.searchQuery != "" {
			return m, searchDevicesCmd(m.client, m.searchQuery)
		}
		return m, loadDevicesCmd(m.client)
	}
	return m, nil
}

func (m DevicesModel) updateDetail(msg tea.KeyMsg) (DevicesModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "b":
		m.inDetail = false
		m.showPortNodes = false
		return m, nil
	case "left":
		if m.detailTab > 0 {
			m.detailTab--
			m.showPortNodes = false
			return m, m.loadCurrentTab()
		}
	case "right":
		if m.detailTab < 3 {
			m.detailTab++
			m.showPortNodes = false
			return m, m.loadCurrentTab()
		}
	case "up":
		if m.detailTab == 1 && m.portsCursor > 0 {
			m.portsCursor--
		}
	case "down":
		if m.detailTab == 1 && m.portsCursor < len(m.ports)-1 {
			m.portsCursor++
		}
	case "enter":
		if m.detailTab == 1 && len(m.ports) > 0 && m.portsCursor < len(m.ports) {
			port := getStringField(m.ports[m.portsCursor], "port")
			if port != "" {
				m.showPortNodes = true
				m.portNodesIP = m.selectedIP
				m.portNodesPort = port
				m.portNodesLoading = true
				return m, loadPortNodesCmd(m.client, m.selectedIP, port)
			}
		}
	case "r":
		return m, m.loadCurrentTab()
	}
	return m, nil
}

func (m DevicesModel) loadCurrentTab() tea.Cmd {
	switch m.detailTab {
	case 0:
		m.detailLoading = true
		return loadDeviceDetailCmd(m.client, m.selectedIP)
	case 1:
		m.portsLoading = true
		return loadDevicePortsCmd(m.client, m.selectedIP)
	case 2:
		m.neighborsLoading = true
		return loadDeviceNeighborsCmd(m.client, m.selectedIP)
	case 3:
		m.vlansLoading = true
		return loadDeviceVlansCmd(m.client, m.selectedIP)
	}
	return nil
}

func (m *DevicesModel) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.devices
	} else {
		query := strings.ToLower(m.searchQuery)
		var filtered []map[string]interface{}
		for _, dev := range m.devices {
			name := strings.ToLower(getStringField(dev, "device_name"))
			ip := strings.ToLower(getStringField(dev, "ip"))
			vendor := strings.ToLower(getStringField(dev, "vendor"))
			model := strings.ToLower(getStringField(dev, "model"))
			location := strings.ToLower(getStringField(dev, "location"))
			if strings.Contains(name, query) || strings.Contains(ip, query) ||
				strings.Contains(vendor, query) || strings.Contains(model, query) ||
				strings.Contains(location, query) {
				filtered = append(filtered, dev)
			}
		}
		m.filtered = filtered
	}
	
	// Apply sorting
	m.sortDevices()
	
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
		m.scrollOffset = 0
	}
}

func (m *DevicesModel) sortDevices() {
	if len(m.filtered) == 0 {
		return
	}

	for i := 0; i < len(m.filtered)-1; i++ {
		for j := i + 1; j < len(m.filtered); j++ {
			swap := false
			d1 := m.filtered[i]
			d2 := m.filtered[j]

			switch m.table.SortColumn {
			case 0: // Name
				val1 := strings.ToLower(getStringField(d1, "device_name"))
				val2 := strings.ToLower(getStringField(d2, "device_name"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 1: // IP
				val1 := strings.ToLower(getStringField(d1, "ip"))
				val2 := strings.ToLower(getStringField(d2, "ip"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 2: // Vendor
				val1 := strings.ToLower(getStringField(d1, "vendor"))
				val2 := strings.ToLower(getStringField(d2, "vendor"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 3: // Model
				val1 := strings.ToLower(getStringField(d1, "model"))
				val2 := strings.ToLower(getStringField(d2, "model"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			case 4: // Location
				val1 := strings.ToLower(getStringField(d1, "location"))
				val2 := strings.ToLower(getStringField(d2, "location"))
				if m.table.SortAscending {
					swap = val1 > val2
				} else {
					swap = val1 < val2
				}
			}

			if swap {
				m.filtered[i], m.filtered[j] = m.filtered[j], m.filtered[i]
			}
		}
	}
}

// --- View ---

func (m *DevicesModel) View() string {
	if m.inDetail {
		return m.viewDetail()
	}
	return m.viewList()
}

func (m *DevicesModel) viewList() string {
	header := TitleStyle.Render("🖥  Devices") +
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
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Loading devices...")
	}
	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s\n\nPress 'r' to retry", m.err.Error())))
	}
	if len(m.filtered) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, "",
			WarningStyle.Render("No devices found."))
	}

	// Set HeaderY for click detection
	m.table.HeaderY = 7

	table := m.renderDeviceTable()
	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render(
		fmt.Sprintf("  %d-%d of %d  ·  ↑↓ navigate  ·  click header to sort  ·  Enter select  ·  / search",
			m.scrollOffset+1, minInt(m.scrollOffset+m.height-8, len(m.filtered)), len(m.filtered)))

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, table, "", footer)
}

func (m DevicesModel) renderDeviceTable() string {
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}

	// Render header and separator using ResizableTable
	headers := []string{"Name", "IP", "Vendor", "Model", "Location"}
	headerRow := m.table.RenderHeader(headers)
	sep := m.table.RenderSeparator()

	var rows []string
	rows = append(rows, headerRow, sep)

	end := minInt(m.scrollOffset+maxVisible, len(m.filtered))
	for i := m.scrollOffset; i < end; i++ {
		dev := m.filtered[i]
		values := []string{
			orNA(shortName(getStringField(dev, "device_name"))),
			orNA(getStringField(dev, "ip")),
			orNA(getStringField(dev, "vendor")),
			orNA(getStringField(dev, "model")),
			orNA(getStringField(dev, "location")),
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

func (m DevicesModel) viewDetail() string {
	header := TitleStyle.Render("🖥  Device Detail") +
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render(fmt.Sprintf("  [%s]", m.selectedIP)) +
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ← Esc/b")

	tabs := []string{"Info", "Ports", "Neighbors", "VLANs"}
	var tabBar []string
	for i, t := range tabs {
		if i == m.detailTab {
			tabBar = append(tabBar, TabActiveStyle.Render(t))
		} else {
			tabBar = append(tabBar, TabInactiveStyle.Render(t))
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabBar...)

	var content string
	switch m.detailTab {
	case 0:
		content = m.viewDeviceInfo()
	case 1:
		content = m.viewDevicePorts()
	case 2:
		content = m.viewDeviceNeighbors()
	case 3:
		content = m.viewDeviceVlans()
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, tabRow, "", content)
}

func (m DevicesModel) viewDeviceInfo() string {
	if m.detailLoading {
		return lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading..."
	}
	if m.detailErr != nil {
		return ErrorStyle.Render(fmt.Sprintf("⚠  %s  (press 'r' to retry)", m.detailErr.Error()))
	}
	if m.detailDevice == nil {
		return WarningStyle.Render("No device data")
	}

	d := m.detailDevice
	fields := []struct{ key, val string }{
		{"Name", orNA(shortName(getStringField(d, "device_name")))},
		{"FQDN", orNA(getStringField(d, "dns"))},
		{"IP", orNA(getStringField(d, "ip"))},
		{"Model", orNA(getStringField(d, "model"))},
		{"Vendor", orNA(getStringField(d, "vendor"))},
		{"OS", orNA(getStringField(d, "os"))},
		{"OS Version", orNA(getStringField(d, "version"))},
		{"Location", orNA(getStringField(d, "location"))},
		{"Contact", orNA(getStringField(d, "contact"))},
		{"Serial", orNA(getStringField(d, "serial"))},
		{"Uptime", func() string {
			if u, ok := d["uptime"]; ok {
				return formatUptime(u)
			}
			return "N/A"
		}()},
		{"Last Polled", formatTime(getStringField(d, "last_discover"))},
	}

	var lines []string
	for _, f := range fields {
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top,
			DetailKeyStyle.Render(f.key),
			DetailValStyle.Render(f.val)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m DevicesModel) viewDevicePorts() string {
	if m.portsLoading {
		return lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading ports..."
	}
	if m.portsErr != nil {
		return ErrorStyle.Render(fmt.Sprintf("⚠  %s  (press 'r' to retry)", m.portsErr.Error()))
	}
	if len(m.ports) == 0 {
		return WarningStyle.Render("No ports found")
	}

	if m.showPortNodes {
		return m.viewPortNodes()
	}

	colPort := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(14).Render("Port")
	colStatus := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(8).Render("Status")
	colSpeed := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(12).Render("Speed")
	colVlan := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(8).Render("VLAN")
	colDescr := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(24).Render("Description")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colPort, colStatus, colSpeed, colVlan, colDescr)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 14) + "┼" + strings.Repeat("─", 8) + "┼" +
			strings.Repeat("─", 12) + "┼" + strings.Repeat("─", 8) + "┼" + strings.Repeat("─", 24))

	var rows []string
	rows = append(rows, headerRow, sep)

	maxVisible := m.height - 14
	if maxVisible < 5 {
		maxVisible = 5
	}
	end := minInt(maxVisible, len(m.ports))

	for i := 0; i < end; i++ {
		p := m.ports[i]
		port := truncate(orNA(getStringField(p, "port")), 12)
		status := StatusDown
		if getStringField(p, "up") == "up" {
			status = StatusUp
		}
		speed := orNA(getStringField(p, "speed"))
		if speed != "N/A" {
			speed += " Mbps"
		}
		vlan := orNA(getStringField(p, "vlan"))
		descr := truncate(orNA(getStringField(p, "descr")), 22)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(14).Foreground(ColorText).Render(port),
			lipgloss.NewStyle().Width(8).Render(status),
			lipgloss.NewStyle().Width(12).Foreground(ColorTextDim).Render(speed),
			lipgloss.NewStyle().Width(8).Foreground(ColorTextMuted).Render(vlan),
			lipgloss.NewStyle().Width(24).Foreground(ColorTextMuted).Render(descr))

		if i == m.portsCursor {
			rows = append(rows, ActiveRowStyle.Render(row))
		} else {
			rows = append(rows, NormalRowStyle.Render(row))
		}
	}

	footer := lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  ↑↓ navigate  ·  Enter → active nodes")
	rows = append(rows, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m DevicesModel) viewPortNodes() string {
	header := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(
		fmt.Sprintf("Nodes on %s:%s", m.portNodesIP, m.portNodesPort)) +
		lipgloss.NewStyle().Foreground(ColorTextMuted).Render("  (Esc back)")

	if m.portNodesLoading {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View())+" Loading nodes...")
	}
	if m.portNodesErr != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			ErrorStyle.Render(fmt.Sprintf("⚠  %s", m.portNodesErr.Error())))
	}
	if len(m.portNodes) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "",
			WarningStyle.Render("No active nodes on this port"))
	}

	var lines []string
	lines = append(lines, header, "")
	for _, n := range m.portNodes {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf(
			"  MAC: %-20s  IP: %-16s  DNS: %s",
			orNA(getStringField(n, "mac")),
			orNA(getStringField(n, "ip")),
			orNA(getStringField(n, "dns")))),
			lipgloss.NewStyle().Foreground(ColorTextMuted).Render(fmt.Sprintf(
				"  VLAN: %-8s  Vendor: %-20s  LastSeen: %s",
				orNA(getStringField(n, "vlan")),
				orNA(getNestedString(n, "manufacturer", "company")),
				formatTime(getStringField(n, "time_last")))),
			"")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m DevicesModel) viewDeviceNeighbors() string {
	if m.neighborsLoading {
		return lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading neighbors..."
	}
	if m.neighborsErr != nil {
		return ErrorStyle.Render(fmt.Sprintf("⚠  %s  (press 'r' to retry)", m.neighborsErr.Error()))
	}
	if m.neighbors == nil {
		return WarningStyle.Render("No neighbor data")
	}

	data, ok := m.neighbors["data"].(map[string]interface{})
	if !ok {
		return WarningStyle.Render("No neighbor data found")
	}

	// Build node name lookup
	nodeNames := map[string]string{}
	if nodes, ok := data["nodes"].([]interface{}); ok {
		for _, n := range nodes {
			if node, ok := n.(map[string]interface{}); ok {
				id := getStringField(node, "ID")
				label := getStringField(node, "LABEL")
				nodeNames[id] = shortName(label)
			}
		}
	}

	links, ok := data["links"].([]interface{})
	if !ok || len(links) == 0 {
		return WarningStyle.Render("No neighbors found")
	}

	colRemote := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(24).Render("Remote Device")
	colIP := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(16).Render("IP")
	colSpeed := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(12).Render("Speed")
	colInfo := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(30).Render("Link Info")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colRemote, colIP, colSpeed, colInfo)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 24) + "┼" + strings.Repeat("─", 16) + "┼" +
			strings.Repeat("─", 12) + "┼" + strings.Repeat("─", 30))

	var rows []string
	rows = append(rows, headerRow, sep)

	for _, l := range links {
		link, ok := l.(map[string]interface{})
		if !ok {
			continue
		}
		isFrom := getStringField(link, "FROMID") == m.selectedIP
		var remoteID string
		if isFrom {
			remoteID = getStringField(link, "TOID")
		} else {
			remoteID = getStringField(link, "FROMID")
		}
		remoteName := nodeNames[remoteID]
		if remoteName == "" {
			remoteName = remoteID
		}
		speed := orNA(getStringField(link, "SPEED"))
		// Strip HTML from INFOSTRING
		info := getStringField(link, "INFOSTRING")
		info = stripHTML(info)
		info = truncate(info, 28)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(24).Foreground(ColorText).Render(truncate(remoteName, 22)),
			lipgloss.NewStyle().Width(16).Foreground(ColorTextDim).Render(truncate(remoteID, 14)),
			lipgloss.NewStyle().Width(12).Foreground(ColorTextMuted).Render(speed),
			lipgloss.NewStyle().Width(30).Foreground(ColorTextMuted).Render(info))
		rows = append(rows, NormalRowStyle.Render(row))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m DevicesModel) viewDeviceVlans() string {
	if m.vlansLoading {
		return lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.spinner.View()) + " Loading VLANs..."
	}
	if m.vlansErr != nil {
		return ErrorStyle.Render(fmt.Sprintf("⚠  %s  (press 'r' to retry)", m.vlansErr.Error()))
	}
	if len(m.vlans) == 0 {
		return WarningStyle.Render("No VLANs found")
	}

	colVlan := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(10).Render("VLAN")
	colName := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Width(30).Render("Name")
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, colVlan, colName)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render(
		strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 30))

	var rows []string
	rows = append(rows, headerRow, sep)

	for _, v := range m.vlans {
		vlan := orNA(getStringField(v, "vlan"))
		name := orNA(getStringField(v, "description"))
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(10).Foreground(ColorText).Render(vlan),
			lipgloss.NewStyle().Width(30).Foreground(ColorTextDim).Render(name))
		rows = append(rows, NormalRowStyle.Render(row))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			if r == '\n' {
				result.WriteString(" | ")
			} else {
				result.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(result.String())
}
