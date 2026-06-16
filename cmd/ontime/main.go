package main

import (
	"fmt"
	"os"
	"time"

	"github.com/duhl/timeon/internal/config"
	"github.com/duhl/timeon/internal/report"
)

func main() {
	cfg, err := config.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	today := time.Now().Format("2006-01-02")
	content, err := report.Read(cfg.ReportPath(today))
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
