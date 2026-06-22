package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/duhl/timeon/internal/config"
	"github.com/duhl/timeon/internal/macos"
	"github.com/duhl/timeon/internal/report"
	"github.com/duhl/timeon/internal/tracker"
)

func main() {
	if len(os.Args) < 2 {
		printTodayReport()
		return
	}

	switch os.Args[1] {
	case "daemon":
		runDaemon()
	case "report", "today":
		printDayReport(0)
	case "day":
		daysAgo, err := optionalOffset(os.Args[2:], "day")
		if err != nil {
			exitUsage(err)
		}
		printDayReport(daysAgo)
	case "week":
		weeksAgo, err := optionalOffset(os.Args[2:], "week")
		if err != nil {
			exitUsage(err)
		}
		printWeekReport(weeksAgo)
	case "diagnose":
		runDiagnose()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func optionalOffset(args []string, command string) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}
	if len(args) > 1 {
		return 0, fmt.Errorf("%s accepts at most one offset", command)
	}

	n, err := strconv.Atoi(args[0])
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s offset must be a non-negative integer", command)
	}
	return n, nil
}

func exitUsage(err error) {
	fmt.Fprintf(os.Stderr, "%v\n\n", err)
	printUsage()
	os.Exit(1)
}

func runDaemon() {
	cfg, err := config.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	tr := tracker.New(cfg)
	if err := tr.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "tracker error: %v\n", err)
		os.Exit(1)
	}
}

func printTodayReport() {
	printDayReport(0)
}

func printDayReport(daysAgo int) {
	cfg, err := config.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	date := reportDate(time.Now(), daysAgo)
	path := cfg.ReportPath(date)

	content, err := report.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "no report for %s yet — is the tracker running?\n", date)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "read report: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(content)
}

func reportDate(now time.Time, daysAgo int) string {
	return now.AddDate(0, 0, -daysAgo).Format("2006-01-02")
}

func printWeekReport(weeksAgo int) {
	cfg, err := config.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	content, err := report.Week(cfg.ReportsDir, weekReference(time.Now(), weeksAgo))
	if err != nil {
		fmt.Fprintf(os.Stderr, "week report: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(content)
}

func weekReference(now time.Time, weeksAgo int) time.Time {
	return now.AddDate(0, 0, -7*weeksAgo)
}

func runDiagnose() {
	diag := macos.DiagnoseFrontmost()
	fmt.Printf("selected:      %s\n", diag.Selected)
	fmt.Printf("lsappinfo:     %s\n", diag.LSAppInfo)
	fmt.Printf("accessibility: %s\n", diag.AX)
	fmt.Printf("system events: %s\n", diag.SystemEvents)
	fmt.Printf("nsworkspace:   %s\n", diag.Workspace)
	fmt.Printf("idle seconds:  %.0f\n", macos.IdleSeconds())
	fmt.Printf("screen locked: %v\n", macos.ScreenLocked())
	fmt.Printf("accessibility granted: %v\n", macos.AccessibilityTrusted())
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `timeon — macOS active time tracker

Usage:
  timeon              Print today's report (same as "timeon report")
  timeon daemon       Run the background tracker
  timeon report       Print today's markdown report
  timeon today        Alias for report
  timeon day [n]      Print today's report, or n days ago
  timeon week [n]     Print this week's daily totals, or n weeks ago
  timeon diagnose     Show current frontmost app detection

Environment:
  TIMEON_REPORTS_DIR  Override reports directory (default: ~/git/timeon/reports)
`)
}
