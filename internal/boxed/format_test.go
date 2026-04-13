package boxed

import (
	"testing"
	"time"
)

func TestFormatTimer(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "0:00"},
		{1, "0:01"},
		{59, "0:59"},
		{60, "1:00"},
		{90, "1:30"},
		{119, "1:59"},
		{120, "2:00"},
		{599, "9:59"},
		{600, "10:00"},
		{1500, "25:00"},
		{3599, "59:59"},
		{3600, "1:00:00"},
		{3660, "1:01:00"},
		{3661, "1:01:01"},
		{5400, "1:30:00"},
		{7200, "2:00:00"},
		{7261, "2:01:01"},
	}
	for _, tt := range tests {
		got := FormatTimer(time.Duration(tt.seconds) * time.Second)
		if got != tt.want {
			t.Errorf("FormatTimer(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "0s"},
		{1, "1s"},
		{59, "59s"},
		{60, "1m"},
		{90, "1m"},
		{119, "1m"},
		{120, "2m"},
		{3599, "59m"},
		{3600, "1h"},
		{3660, "1h1m"},
		{3661, "1h1m"},
		{7200, "2h"},
		{7261, "2h1m"},
	}
	for _, tt := range tests {
		got := FormatDuration(time.Duration(tt.seconds) * time.Second)
		if got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}
