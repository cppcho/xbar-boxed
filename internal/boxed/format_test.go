package boxed

import "testing"

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
		got := FormatDuration(tt.seconds)
		if got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}
