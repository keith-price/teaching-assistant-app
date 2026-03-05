package calendar

import (
	"testing"
	"time"
)

func TestParseEventTime(t *testing.T) {
	loc := getWIBLocation()

	tests := []struct {
		name       string
		timeStr    string
		expectErr  bool
		expectTime time.Time
	}{
		{
			name:      "Valid RFC3339 UTC to WIB",
			timeStr:   "2026-03-05T10:00:00Z",
			expectErr: false,
			// UTC 10:00 is WIB 17:00
			expectTime: time.Date(2026, 3, 5, 17, 0, 0, 0, loc),
		},
		{
			name:      "Valid RFC3339 WIB with offset",
			timeStr:   "2026-03-05T15:00:00+07:00",
			expectErr: false,
			// Already in WIB
			expectTime: time.Date(2026, 3, 5, 15, 0, 0, 0, loc),
		},
		{
			name:      "Valid All-Day Event (Date only)",
			timeStr:   "2026-03-05",
			expectErr: false,
			// All day event parsed into WIB timezone directly
			expectTime: time.Date(2026, 3, 5, 0, 0, 0, 0, loc),
		},
		{
			name:       "Invalid Time String",
			timeStr:    "invalid-time",
			expectErr:  true,
			expectTime: time.Time{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsedTime, err := parseEventTime(tc.timeStr, loc)
			if (err != nil) != tc.expectErr {
				t.Fatalf("expected error: %v, got: %v", tc.expectErr, err)
			}
			if !tc.expectErr && !parsedTime.Equal(tc.expectTime) {
				t.Errorf("expected time: %v, got: %v", tc.expectTime, parsedTime)
			}
			if !tc.expectErr && parsedTime.Location() != loc {
				t.Errorf("expected location WIB, got: %v", parsedTime.Location())
			}
		})
	}
}

func TestWIBLocation(t *testing.T) {
	loc := getWIBLocation()
	if loc.String() != "WIB" {
		t.Errorf("Expected location name WIB, got %s", loc.String())
	}

	// Test that offset is UTC+7
	now := time.Now().In(loc)
	_, offset := now.Zone()
	expectedOffset := 7 * 3600
	if offset != expectedOffset {
		t.Errorf("Expected offset %d, got %d", expectedOffset, offset)
	}
}
