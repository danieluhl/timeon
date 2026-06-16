package report

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/duhl/timeon/internal/state"
)

type AppUsage struct {
	Name    string
	Seconds int
}

func Write(path string, s *state.DayState) error {
	var b strings.Builder
	now := time.Now()

	b.WriteString(fmt.Sprintf("# Time Report — %s\n\n", s.Date))
	b.WriteString(fmt.Sprintf("**Last updated:** %s\n\n", now.Format("2006-01-02 15:04:05")))

	b.WriteString("## Total Active Time\n\n")
	b.WriteString(fmt.Sprintf("**%s** (%d seconds)\n\n", formatDuration(s.TotalSeconds), s.TotalSeconds))

	b.WriteString("## Active Time by Application\n\n")
	apps := sortedApps(s)
	if len(apps) == 0 {
		b.WriteString("_No active time recorded yet._\n\n")
	} else {
		b.WriteString("| Application | Active Time | % of Day |\n")
		b.WriteString("|---|---|---|\n")
		for _, app := range apps {
			pct := 0.0
			if s.TotalSeconds > 0 {
				pct = float64(app.Seconds) / float64(s.TotalSeconds) * 100
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %.1f%% |\n",
				escapePipe(app.Name),
				formatDuration(app.Seconds),
				pct,
			))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Timeline (10-minute blocks)\n\n")
	blocks := timelineBlocks(s, now)
	if len(blocks) == 0 {
		b.WriteString("_No activity recorded yet._\n")
	} else {
		for _, block := range blocks {
			end := state.BlockEndLabel(block.Key)
			activeMin := block.Seconds / 60
			remainder := block.Seconds % 60
			label := fmt.Sprintf("%d minutes active", activeMin)
			if remainder > 0 {
				label = fmt.Sprintf("%d minutes %d seconds active", activeMin, remainder)
			}
			if activeMin == 0 && remainder == 0 {
				label = "0 minutes active"
			}
			b.WriteString(fmt.Sprintf("- **%s - %s:** %s\n", block.Key, end, label))
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func sortedApps(s *state.DayState) []AppUsage {
	apps := make([]AppUsage, 0, len(s.AppSeconds))
	for name, seconds := range s.AppSeconds {
		apps = append(apps, AppUsage{Name: name, Seconds: seconds})
	}
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Seconds == apps[j].Seconds {
			return apps[i].Name < apps[j].Name
		}
		return apps[i].Seconds > apps[j].Seconds
	})
	return apps
}

type blockEntry struct {
	Key     string
	Seconds int
}

func timelineBlocks(s *state.DayState, now time.Time) []blockEntry {
	loc := now.Location()
	dayStart, err := time.ParseInLocation("2006-01-02", s.Date, loc)
	if err != nil {
		return nil
	}

	end := now.Truncate(10 * time.Minute).Add(10 * time.Minute)
	dayEnd := dayStart.Add(24 * time.Hour)
	if end.After(dayEnd) {
		end = dayEnd
	}

	var blocks []blockEntry
	for t := dayStart; t.Before(end); t = t.Add(10 * time.Minute) {
		key := t.Format("15:04")
		seconds := s.BlockSeconds[key]
		if seconds > 0 {
			blocks = append(blocks, blockEntry{Key: key, Seconds: seconds})
		}
	}
	return blocks
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", seconds)
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
