package ui

import (
	"strconv"
	"time"
)

func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return strconv.Itoa(mins) + "m ago"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return strconv.Itoa(hours) + "h ago"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return strconv.Itoa(days) + "d ago"
	default:
		return t.Format("Jan 2")
	}
}

func GetSessionAge(t time.Time) int {
	if t.IsZero() {
		return 2
	}

	diff := time.Since(t)
	switch {
	case diff < 24*time.Hour:
		return 0
	case diff < 7*24*time.Hour:
		return 1
	default:
		return 2
	}
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	val := float64(bytes) / float64(div)
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return strconv.FormatFloat(val, 'f', 1, 64) + " " + units[exp]
}
