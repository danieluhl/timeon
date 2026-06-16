package tracker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duhl/timeon/internal/config"
	"github.com/duhl/timeon/internal/macos"
	"github.com/duhl/timeon/internal/report"
	"github.com/duhl/timeon/internal/state"
)

const stateSaveInterval = 10 * time.Second

type Tracker struct {
	cfg          *config.Config
	log          *log.Logger
	dirty        bool
	lastSave     time.Time
	lastLoggedApp string
}

func New(cfg *config.Config) *Tracker {
	return &Tracker{
		cfg: cfg,
		log: log.New(os.Stderr, "timeon: ", log.LstdFlags),
	}
}

func (t *Tracker) Run(ctx context.Context) error {
	if err := t.cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("create reports directory: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	dayState, err := state.Load(t.cfg.StatePath())
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if dayState == nil || dayState.Date != today {
		if dayState != nil && dayState.Date != today {
			if err := t.writeReport(dayState); err != nil {
				t.log.Printf("final report for %s: %v", dayState.Date, err)
			}
		}
		dayState = state.NewDay(today)
	}

	if err := t.writeReport(dayState); err != nil {
		t.log.Printf("initial report: %v", err)
	}

	if !macos.AccessibilityTrusted() {
		macos.RequestAccessibility()
		t.log.Printf("accessibility permission is recommended for accurate app detection; enable timeon in System Settings → Privacy & Security → Accessibility")
	}

	ticker := time.NewTicker(config.PollInterval)
	reportTicker := time.NewTicker(config.ReportInterval)
	defer ticker.Stop()
	defer reportTicker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	lastPoll := time.Now()
	if !dayState.LastPoll.IsZero() {
		lastPoll = dayState.LastPoll
	}

	for {
		select {
		case <-ctx.Done():
			return t.shutdown(dayState)
		case sig := <-sigCh:
			t.log.Printf("received %s, shutting down", sig)
			return t.shutdown(dayState)
		case <-reportTicker.C:
			if err := t.writeReport(dayState); err != nil {
				t.log.Printf("periodic report: %v", err)
			}
		case now := <-ticker.C:
			for {
				dayState, lastPoll = t.tick(dayState, lastPoll, now)
				select {
				case now = <-ticker.C:
					continue
				default:
					goto doneTicks
				}
			}
		doneTicks:
			t.maybeSave(dayState, false)
		}
	}
}

func (t *Tracker) tick(dayState *state.DayState, lastPoll, now time.Time) (*state.DayState, time.Time) {
	today := now.Format("2006-01-02")
	if dayState.Date != today {
		if err := t.writeReport(dayState); err != nil {
			t.log.Printf("midnight report for %s: %v", dayState.Date, err)
		}
		dayState = state.NewDay(today)
		lastPoll = now
	}

	elapsed := now.Sub(lastPoll)
	if elapsed > config.SleepGapThreshold {
		t.log.Printf("sleep/lock gap detected (%s), skipping", elapsed.Round(time.Second))
		dayState.LastPoll = now
		t.maybeSave(dayState, true)
		return dayState, now
	}

	if macos.ScreenLocked() {
		dayState.LastPoll = now
		t.maybeSave(dayState, true)
		return dayState, now
	}

	if macos.IdleSeconds() >= config.IdleThreshold.Seconds() {
		dayState.LastPoll = now
		t.maybeSave(dayState, true)
		return dayState, now
	}

	app := macos.FrontmostApp()
	if app != t.lastLoggedApp {
		t.log.Printf("active app: %s", app)
		t.lastLoggedApp = app
	}
	seconds := int(config.PollInterval.Seconds())
	dayState.AddActive(app, seconds, now)
	t.dirty = true

	return dayState, now
}

func (t *Tracker) maybeSave(dayState *state.DayState, force bool) {
	if !force && (!t.dirty || time.Since(t.lastSave) < stateSaveInterval) {
		return
	}
	if err := dayState.Save(t.cfg.StatePath()); err != nil {
		t.log.Printf("save state: %v", err)
		return
	}
	t.dirty = false
	t.lastSave = time.Now()
}

func (t *Tracker) writeReport(dayState *state.DayState) error {
	path := t.cfg.ReportPath(dayState.Date)
	if err := report.Write(path, dayState); err != nil {
		return err
	}
	dayState.LastReport = time.Now()
	t.dirty = true
	return dayState.Save(t.cfg.StatePath())
}

func (t *Tracker) shutdown(dayState *state.DayState) error {
	if err := t.writeReport(dayState); err != nil {
		return err
	}
	t.log.Println("shutdown complete")
	return nil
}
