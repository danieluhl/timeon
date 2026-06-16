package main

import (
	"context"
	"fmt"
	"os"
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
		printTodayReport()
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
	cfg, err := config.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	today := time.Now().Format("2006-01-02")
	path := cfg.ReportPath(today)

	content, err := report.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "no report for today (%s) yet — is the tracker running?\n", today)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "read report: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(content)
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
  timeon diagnose     Show current frontmost app detection

Environment:
  TIMEON_REPORTS_DIR  Override reports directory (default: ~/git/timeon/reports)
`)
}
