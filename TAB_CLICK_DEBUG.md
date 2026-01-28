# Tab Click Detection - Debug Guide

## Current Implementation (2026-01-28 14:46 CST)

### Fixed Position Thresholds

**Main Tabs:**
- X: 0-20 → Devices
- X: 20-40 → Nodes
- X: 40-60 → VLANs
- X: 60-80 → Subnets
- X: 80+ → Reports

**Reports Sub-Tabs:**
- X: 0-18 → Recent
- X: 18-36 → All Devices
- X: 36-54 → By Vendor
- X: 54-120 → VLANs

## How to Adjust Thresholds

If clicking still doesn't work correctly, adjust the numbers in:
1. `internal/tui/root.go` - Main tab thresholds
2. `internal/tui/reports.go` - Reports sub-tab thresholds

### Symptoms and Fixes

**"Have to click too far to the right"**
- Thresholds are TOO LARGE
- Reduce the numbers by 5-10

Example fix:
```go
// Change from:
if x >= 0 && x < 20    → Devices
if x >= 20 && x < 40   → Nodes

// To:
if x >= 0 && x < 15    → Devices
if x >= 15 && x < 30   → Nodes
```

**"Wrong tab activates when clicking"**
- Thresholds might be TOO SMALL or TOO LARGE
- Observe which tab SHOULD activate vs which DOES activate
- Adjust the boundary between them

**"Can't click tabs at all"**
- Y coordinate range might be wrong
- Check if Y detection (rows 1-4 for main, 4-10 for reports) needs adjustment

## Testing Procedure

1. **Test main tabs (Devices, Nodes, VLANs, Subnets, Reports):**
   - Click far left edge of "Devices" text
   - Click middle of "Devices" text
   - Click far right edge of "Devices" text
   - Repeat for each tab

2. **Test Reports sub-tabs:**
   - Go to Reports tab
   - Click far left, middle, right of each sub-tab name
   - Note which clicks work, which don't

3. **Record observations:**
   - If "Devices" tab requires clicking at X=15, but threshold starts at X=0
   - Then tabs are shifted ~15 chars to the right
   - Adjust ALL thresholds by adding 15

## Understanding Terminal Coordinates

Terminal uses character-based coordinates:
- X=0 is leftmost column
- X increases to the right
- Y=0 is top row
- Y increases downward

Lipgloss rendering adds:
- Borders: 1 char each side
- Padding: 2 chars each side (configurable)
- Margins: 1 char right (configurable)

A tab labeled "📱 Devices" renders roughly as:
```
┌─────────────────┐
│  📱 Devices     │
└─────────────────┘
```
Width = 1 (border) + 2 (padding) + 2 (emoji) + 1 (space) + 7 (Devices) + padding/space + 1 (border) = ~15-18 chars

## Quick Threshold Adjustment Template

```go
// Main tabs (adjust these numbers in root.go)
if x >= START1 && x < END1 { return m.switchTab(TabDevices) }
if x >= START2 && x < END2 { return m.switchTab(TabNodes) }
// ... etc

// Where:
// START1 = 0 (or left padding if tabs don't start at edge)
// END1 = START1 + estimated tab width (15-20 chars)
// START2 = END1
// END2 = START2 + tab width
// ... and so on
```

## Current Status

These thresholds are ESTIMATES based on:
- Tab bar wrapper: Padding(1, 2) = 2 chars left padding
- Each tab: ~15-20 chars wide
- Left-aligned layout

If they don't match reality, user feedback will help calibrate them.
