package report

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const minAverageSeconds = 3600

var totalSecondsRE = regexp.MustCompile(`\((\d+) seconds\)`)

type dayTotal struct {
	date    string
	weekday string
	seconds int
	hasData bool
}

func Week(reportsDir string, ref time.Time) (string, error) {
	ref = ref.In(time.Local)
	monday := weekMonday(ref)

	var days []dayTotal
	for i := range 7 {
		d := monday.AddDate(0, 0, i)
		date := d.Format("2006-01-02")
		path := reportPath(reportsDir, date)
		seconds, hasData, err := totalSecondsFromFile(path)
		if err != nil {
			return "", err
		}
		days = append(days, dayTotal{
			date:    date,
			weekday: d.Format("Mon"),
			seconds: seconds,
			hasData: hasData,
		})
	}

	return formatWeek(monday, days), nil
}

func weekMonday(ref time.Time) time.Time {
	ref = time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
	daysSinceMonday := (int(ref.Weekday()) + 6) % 7
	return ref.AddDate(0, 0, -daysSinceMonday)
}

func reportPath(reportsDir, date string) string {
	return filepath.Join(reportsDir, date+".md")
}

func totalSecondsFromFile(path string) (seconds int, hasData bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, err
	}

	match := totalSecondsRE.FindSubmatch(data)
	if match == nil {
		return 0, true, fmt.Errorf("parse total time in %s", path)
	}

	n, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return 0, true, fmt.Errorf("parse total time in %s: %w", path, err)
	}
	return n, true, nil
}

func formatWeek(monday time.Time, days []dayTotal) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Time Report — Week of %s\n\n", monday.Format("2006-01-02")))
	b.WriteString("| Day | Date | Active Time |\n")
	b.WriteString("|---|---|---|\n")

	var qualifying []int
	for _, day := range days {
		active := "—"
		if day.hasData {
			active = formatDuration(day.seconds)
			if day.seconds > minAverageSeconds {
				qualifying = append(qualifying, day.seconds)
			}
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", day.weekday, day.date, active))
	}

	b.WriteString("\n## Daily Average\n\n")
	if len(qualifying) == 0 {
		b.WriteString("_No days with more than 1 hour of active time this week._\n")
		return b.String()
	}

	var sum int
	for _, s := range qualifying {
		sum += s
	}
	avg := sum / len(qualifying)
	b.WriteString(fmt.Sprintf("**%s** (%d day", formatDuration(avg), len(qualifying)))
	if len(qualifying) == 1 {
		b.WriteString(" with >1h active time)\n")
	} else {
		b.WriteString("s with >1h active time)\n")
	}

	return b.String()
}
