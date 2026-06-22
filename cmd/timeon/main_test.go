package main

import (
	"strings"
	"testing"
	"time"
)

func TestOptionalOffset(t *testing.T) {
	t.Run("defaults to zero", func(t *testing.T) {
		got, err := optionalOffset(nil, "week")
		if err != nil {
			t.Fatalf("optionalOffset() error = %v", err)
		}
		if got != 0 {
			t.Fatalf("optionalOffset() = %d, want 0", got)
		}
	})

	t.Run("parses one non-negative integer", func(t *testing.T) {
		got, err := optionalOffset([]string{"3"}, "day")
		if err != nil {
			t.Fatalf("optionalOffset() error = %v", err)
		}
		if got != 3 {
			t.Fatalf("optionalOffset() = %d, want 3", got)
		}
	})

	t.Run("rejects negative integers", func(t *testing.T) {
		_, err := optionalOffset([]string{"-1"}, "week")
		if err == nil {
			t.Fatal("optionalOffset() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "non-negative integer") {
			t.Fatalf("optionalOffset() error = %q, want non-negative integer", err)
		}
	})

	t.Run("rejects extra arguments", func(t *testing.T) {
		_, err := optionalOffset([]string{"1", "2"}, "day")
		if err == nil {
			t.Fatal("optionalOffset() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "at most one offset") {
			t.Fatalf("optionalOffset() error = %q, want at most one offset", err)
		}
	})
}

func TestReportDate(t *testing.T) {
	now := time.Date(2026, time.June, 22, 12, 0, 0, 0, time.UTC)

	got := reportDate(now, 3)
	if got != "2026-06-19" {
		t.Fatalf("reportDate() = %q, want 2026-06-19", got)
	}
}

func TestWeekReference(t *testing.T) {
	now := time.Date(2026, time.June, 22, 12, 0, 0, 0, time.UTC)

	got := weekReference(now, 1)
	want := time.Date(2026, time.June, 15, 12, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("weekReference() = %v, want %v", got, want)
	}
}
