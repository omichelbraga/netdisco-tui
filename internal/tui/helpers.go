package tui

import (
	"fmt"
	"strings"
	"time"
)

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%.0f", f)
		}
		if v != nil {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func getNestedString(m map[string]interface{}, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			return getStringField(current, key)
		}
		if nested, ok := current[key]; ok {
			if nestedMap, ok := nested.(map[string]interface{}); ok {
				current = nestedMap
			} else {
				return ""
			}
		} else {
			return ""
		}
	}
	return ""
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func shortName(fqdn string) string {
	if fqdn == "" {
		return "N/A"
	}
	parts := strings.SplitN(fqdn, ".", 2)
	return parts[0]
}

func formatTime(t string) string {
	if t == "" {
		return "N/A"
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05 -0700",
	}
	for _, f := range formats {
		if parsed, err := time.Parse(f, t); err == nil {
			return parsed.Format("2006-01-02 15:04")
		}
	}
	return t
}

func formatUptime(val interface{}) string {
	var secs float64
	switch v := val.(type) {
	case float64:
		secs = v / 100
	case string:
		return v
	default:
		return "N/A"
	}
	d := time.Duration(secs) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
