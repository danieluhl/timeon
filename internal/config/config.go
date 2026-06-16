package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	PollInterval      = 1 * time.Second
	IdleThreshold     = 5 * time.Minute
	ReportInterval    = 10 * time.Minute
	BlockDuration     = 10 * time.Minute
	SleepGapThreshold = 5 * time.Second
	StateFileName     = ".tracker-state.json"
)

type Config struct {
	ReportsDir string
}

func Default() (*Config, error) {
	dir := os.Getenv("TIMEON_REPORTS_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(home, "git", "timeon", "reports")
	}
	return &Config{ReportsDir: dir}, nil
}

func (c *Config) StatePath() string {
	return filepath.Join(c.ReportsDir, StateFileName)
}

func (c *Config) ReportPath(date string) string {
	return filepath.Join(c.ReportsDir, date+".md")
}

func (c *Config) EnsureDirs() error {
	return os.MkdirAll(c.ReportsDir, 0o755)
}
