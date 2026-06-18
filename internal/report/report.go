package report

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/duhl/timeon/internal/state"
)

const minAppSeconds = 60

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
		if seconds < minAppSeconds {
			continue
		}
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
