# UI Test Checklist

## Fixed Issues
- ✅ Header bar width calculation fixed
- ✅ Footer no longer uses fixed width
- ✅ Theme menu ESC key now closes menu
- ✅ Number keys (1-6) only work for themes when theme menu is open
- ✅ Number keys (1-5) only work for tabs when theme menu is closed
- ✅ Tab click detection updated for new layout
- ✅ Reports sub-tab click detection updated
- ✅ TitleStyle no longer has background to prevent layout conflicts
- ✅ **CRITICAL: Content height overflow fixed** - Header no longer disappears!
  - Root cause: Content taller than terminal caused scrolling
  - Fix: Height constraints enforce max content height
  - All tabs protected from overflow
  - Header always stays visible regardless of data size or clicks

## Layout Structure (for reference)
```
Row 0:     Header bar "⚡ NETDISCO TUI ... Theme: X [T]"
Rows 1-3:  Main tabs (📱 Devices | 🔍 Nodes | 🌐 VLANs | 📊 Subnets | 📋 Reports)
Row 4+:    Content area
           - For Reports tab: Rows 4-6 are Reports sub-tabs
```

## Test Scenarios

### Main Navigation
- [ ] Click on each main tab (Devices, Nodes, VLANs, Subnets, Reports)
- [ ] Use keyboard 1-5 to switch main tabs
- [ ] Use ← → arrows to navigate main tabs
- [ ] Verify header stays visible on all tabs

### Reports Tab
- [ ] Click on Reports sub-tabs (Recent, All Devices, By Vendor, VLANs)
- [ ] Use ← → arrows to navigate Reports sub-tabs
- [ ] Verify all report types load correctly

### Theme System
- [ ] Press T to open theme menu
- [ ] Press 1-6 to select each theme
- [ ] Press ESC to close theme menu without selecting
- [ ] Press T again to close theme menu
- [ ] Verify themes apply correctly to all UI elements
- [ ] Switch between tabs with different themes active

### Tables
- [ ] Click and drag column separators (╪ symbols) to resize
- [ ] Verify columns stay aligned after resizing
- [ ] Test scrolling with ↑↓ on large lists (Devices, Subnets, VLANs)
- [ ] Test Page Up/Down on Subnets tab
- [ ] Verify table headers have consistent styling

### Search & Filtering
- [ ] Use / to activate search on Devices tab
- [ ] Use / to activate search on VLANs tab
- [ ] Test ESC to clear search
- [ ] Verify search results display correctly

### Node Lookup
- [ ] Enter IP/MAC address
- [ ] Tab to focus on results
- [ ] Use ↑↓ to navigate results
- [ ] Press Enter to view node details

### Subnet Utilization
- [ ] Enter CIDR (e.g., 10.0.0.0/24)
- [ ] Test scrolling with large result sets
- [ ] Press Enter to view IP inventory
- [ ] Test ESC/b to go back

### Visual Elements
- [ ] Verify emojis display correctly on all tabs
- [ ] Check border colors are visible
- [ ] Confirm active row highlighting works
- [ ] Test error/warning message styling
- [ ] Verify spinner appears during loading

## Known Limitations
- Mouse hover states not available (Bubbletea limitation)
- Terminal color support varies by terminal emulator
- Some emoji may not render in all terminals

## Recommended Test Environment
- Terminal: Modern terminal with 256-color support
- Size: At least 120x40 characters
- Font: Monospace with good emoji support
- Environment: `NETDISCO_URL` and `NETDISCO_TOKEN` set

## Quick Test Commands
```bash
# Set environment
export NETDISCO_URL="https://your-netdisco-server"
export NETDISCO_TOKEN="your-token"

# Run app
./netdisco-tui

# Test sequence:
# 1. Press T, select theme 2 (Nord)
# 2. Click through all 5 main tabs
# 3. On Reports, click through all 4 sub-tabs
# 4. On Devices, press / and search "switch"
# 5. Test column resizing by dragging ╪
```

## Report Issues
If you find bugs, note:
1. Which tab you were on
2. What action you took
3. What happened vs expected behavior
4. Your terminal emulator and size
