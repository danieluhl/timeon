package state

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type DayState struct {
	Date         string         `json:"date"`
	TotalSeconds int            `json:"total_seconds"`
	AppSeconds   map[string]int `json:"app_seconds"`
	BlockSeconds map[string]int `json:"block_seconds"`
	LastPoll     time.Time      `json:"last_poll"`
	LastReport   time.Time      `json:"last_report"`
}

func NewDay(date string) *DayState {
	return &DayState{
		Date:         date,
		AppSeconds:   make(map[string]int),
		BlockSeconds: make(map[string]int),
		LastPoll:     time.Now(),
	}
}

func Load(path string) (*DayState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var s DayState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.AppSeconds == nil {
		s.AppSeconds = make(map[string]int)
	}
	if s.BlockSeconds == nil {
		s.BlockSeconds = make(map[string]int)
	}
	return &s, nil
}

func (s *DayState) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *DayState) AddActive(app string, seconds int, at time.Time) {
	s.TotalSeconds += seconds
	s.AppSeconds[app] += seconds
	blockKey := BlockKey(at)
	s.BlockSeconds[blockKey] += seconds
	s.LastPoll = at
}

func BlockKey(t time.Time) string {
	block := t.Truncate(10 * time.Minute)
	return block.Format("15:04")
}

func BlockEndLabel(startKey string) string {
	t, err := time.Parse("15:04", startKey)
	if err != nil {
		return startKey
	}
	return t.Add(10 * time.Minute).Format("15:04")
}
