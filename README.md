# Netdisco TUI

<div align="center">

![Netdisco TUI](https://img.shields.io/badge/Netdisco-TUI-00ffff?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**A beautiful Terminal User Interface for Netdisco Network Discovery**

*Network Discovery Made Beautiful*

</div>

---

## ✨ Features

- 🎨 **6 Beautiful Themes** - Cyberpunk, Nord, Dracula, Tokyo Night, Gruvbox, Monokai
- 🖱️ **Full Mouse Support** - Click, drag, and resize columns
- 📊 **Sortable Tables** - Click column headers to sort (all columns, all tabs)
- 🔍 **Real-time Search** - Filter devices, VLANs, subnets, and more
- 🌐 **Multiple Views**:
  - 📱 **Devices** - Browse and search network devices
  - 🔍 **Nodes** - Search by IP or MAC address
  - 🌐 **VLANs** - View VLAN inventory across your network
  - 📊 **Subnets** - Subnet utilization and IP inventory
  - 📋 **Reports** - Recently added devices, vendor statistics, and more
- 🎭 **Animated Splash Screen** - Eye-catching startup animation
- 💾 **Theme Persistence** - Your theme choice is saved across sessions
- 🪟 **Cross-Platform** - Works on Linux, macOS, and Windows

---

## 🚀 Quick Start

### Prerequisites

- **Netdisco** server with API access
- **Go** 1.21 or later (for building from source)

### Installation

#### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/netdisco-tui.git
cd netdisco-tui

# Build for your platform
go build -o netdisco-tui .

# Or build for Windows
GOOS=windows GOARCH=amd64 go build -o netdisco-tui.exe .
```

#### Download Binary

Download the latest release from the [Releases](https://github.com/yourusername/netdisco-tui/releases) page.

---

## ⚙️ Configuration

Netdisco TUI uses environment variables for configuration:

### Required Variables

```bash
# Your Netdisco API token (required)
export NETDISCO_TOKEN="your-api-token-here"

# Your Netdisco server URL
export NETDISCO_URL="https://netdisco.example.com/api/v1"
```

### Optional Variables

```bash
# HTTP timeout in seconds (default: 30)
export NETDISCO_TIMEOUT=30

# Maximum retry attempts (default: 3)
export NETDISCO_RETRIES=3
```

### Platform-Specific Setup

#### Linux / macOS

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
export NETDISCO_URL="https://netdisco.example.com/api/v1"
export NETDISCO_TOKEN="your-api-token-here"
```

Then reload:
```bash
source ~/.bashrc  # or ~/.zshrc
```

#### Windows (PowerShell)

**Temporary (current session):**
```powershell
$env:NETDISCO_URL = "https://netdisco.example.com/api/v1"
$env:NETDISCO_TOKEN = "your-api-token-here"
```

**Permanent (system-wide):**
1. Press `Win + X` → System
2. Click "Advanced system settings"
3. Click "Environment Variables"
4. Add new variables under "User variables"

**Or use a batch file:**

Create `run-netdisco.bat`:
```batch
@echo off
set NETDISCO_URL=https://netdisco.example.com/api/v1
set NETDISCO_TOKEN=your-api-token-here
netdisco-tui.exe
```

---

## 🎮 Usage

### Starting the Application

```bash
# Linux/macOS
./netdisco-tui

# Windows (PowerShell)
.\netdisco-tui.exe

# Or double-click netdisco-tui.exe
```

### Keyboard Shortcuts

#### Global

- `1-5` - Switch between tabs (Devices, Nodes, VLANs, Subnets, Reports)
- `T` - Open theme selector
- `q` or `Ctrl+C` - Quit
- `←` `→` - Navigate between tabs
- `↑` `↓` - Navigate lists

#### Tab-Specific

**Devices Tab:**
- `/` - Search devices
- `Enter` - View device details
- `Esc` or `b` - Back from detail view
- `r` - Refresh device list

**Nodes Tab:**
- `/` - Search by IP or MAC
- `Enter` - Search
- `Esc` - Clear search

**VLANs Tab:**
- `/` - Search VLANs
- `r` - Refresh VLAN list

**Subnets Tab:**
- `Enter` - Search for subnet or view IP inventory
- `Esc` or `b` - Back to subnet list

**Reports Tab:**
- `←` `→` - Switch between report types
- `+` `-` - Adjust time range (Recent Devices report)
- `r` - Refresh current report

### Mouse Controls

- **Click column headers** - Sort by that column (toggles ascending/descending)
- **Click tabs** - Switch between views
- **Click theme boxes** - Select a theme (when theme menu is open)
- **Drag column separators** - Resize columns (╪ indicators)

### Themes

Press `T` to open the theme selector, then:
- Click on a theme preview, or
- Press `1-6` to select a theme

**Available themes:**
1. **Cyberpunk** - Neon blues and magentas
2. **Nord** - Cool Nordic palette
3. **Dracula** - Purple and pink dark theme
4. **Tokyo Night** - Blue Japanese-inspired theme
5. **Gruvbox** - Retro warm colors
6. **Monokai** - Classic coding theme

Your theme selection is automatically saved and restored on next launch.

---

## 📊 Features by Tab

### 📱 Devices

- Browse all network devices
- Search by name, IP, vendor, model, or location
- View detailed device information
- See device ports, neighbors, and VLANs
- Sortable by all columns

### 🔍 Nodes

- Search for nodes by IP or MAC address
- View node details and sightings
- See where a MAC has been seen across your network

### 🌐 VLANs

- View VLAN inventory across your network
- See how many devices and ports use each VLAN
- Search and filter VLANs
- Sort by ID, name, device count, or port count

### 📊 Subnets

- View subnet utilization
- Drill down into IP inventory for any subnet
- See used vs. free addresses
- Sort by utilization percentage

### 📋 Reports

Multiple report types:
- **Recent Devices** - Devices added in the last N days
- **All Devices** - Complete device inventory
- **By Vendor** - Device count and percentage by vendor
- **VLANs** - VLAN summary report

---

## 🛠️ Development

### Project Structure

```
netdisco-tui/
├── main.go                    # Application entry point
├── internal/
│   ├── api/
│   │   └── client.go         # Netdisco API client
│   ├── config/
│   │   └── config.go         # Configuration management
│   └── tui/
│       ├── root.go           # Root UI model
│       ├── splash.go         # Animated splash screen
│       ├── devices.go        # Devices tab
│       ├── nodes.go          # Nodes tab
│       ├── vlans.go          # VLANs tab
│       ├── subnets.go        # Subnets tab
│       ├── reports.go        # Reports tab
│       ├── table.go          # Resizable table component
│       ├── themes.go         # Theme definitions
│       ├── styles.go         # UI styles
│       └── helpers.go        # Utility functions
├── go.mod
├── go.sum
└── README.md
```

### Building

```bash
# Build for current platform
go build -o netdisco-tui .

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o netdisco-tui-linux .
GOOS=darwin GOARCH=amd64 go build -o netdisco-tui-macos .
GOOS=windows GOARCH=amd64 go build -o netdisco-tui.exe .
```

### Dependencies

This project uses:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style and layout
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

---

## 🐛 Troubleshooting

### Authentication Failed

**Error:** `authentication failed — check NETDISCO_TOKEN`

**Solution:** Verify your API token is correct and has proper permissions.

### Connection Issues

**Error:** `request failed after N retries`

**Solution:** 
- Check that `NETDISCO_URL` is correct
- Verify network connectivity
- Ensure Netdisco server is accessible
- Check for SSL certificate issues (the app accepts self-signed certificates)

### Mouse Not Working

**Solution:**
- Ensure you're using a terminal that supports mouse input (Windows Terminal, iTerm2, modern Linux terminals)
- On Windows, use Windows Terminal or PowerShell for best results
- CMD.exe has limited mouse support

### Theme Not Persisting

**Solution:** Ensure you have write permissions to your home directory. Theme is saved to `~/.netdisco-tui-theme`.

---

## 📝 License

MIT License - See LICENSE file for details

---

## 👤 Author

**Mike Guimaraes**

---

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
- Designed for [Netdisco](https://github.com/netdisco/netdisco)
- Inspired by modern terminal UIs

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

<div align="center">

**⭐ Star this repository if you find it useful! ⭐**

Made with ❤️ and ☕ by Mike Guimaraes

</div>
