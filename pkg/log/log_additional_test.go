package log

import (
	"testing"
	"time"
)

func TestParseRelativeTime(t *testing.T) {
	tests := []struct {
		name         string
		relativeTime string
		wantError    bool
		duration     time.Duration
	}{
		{
			name:         "empty string",
			relativeTime: "",
			wantError:    false,
		},
		{
			name:         "seconds",
			relativeTime: "-30s",
			wantError:    false,
			duration:     30 * time.Second,
		},
		{
			name:         "minutes",
			relativeTime: "-15m",
			wantError:    false,
			duration:     15 * time.Minute,
		},
		{
			name:         "hours",
			relativeTime: "-2h",
			wantError:    false,
			duration:     2 * time.Hour,
		},
		{
			name:         "days",
			relativeTime: "-3d",
			wantError:    false,
			duration:     3 * 24 * time.Hour,
		},
		{
			name:         "invalid format",
			relativeTime: "-3x",
			wantError:    true,
		},
		{
			name:         "invalid number",
			relativeTime: "-xh",
			wantError:    true,
		},
		{
			name:         "positive value",
			relativeTime: "3h",
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			result, err := parseRelativeTime(tt.relativeTime)

			if tt.wantError {
				if err == nil {
					t.Errorf("parseRelativeTime(%q) expected error, got nil", tt.relativeTime)
				}
				return
			}

			if err != nil {
				t.Errorf("parseRelativeTime(%q) unexpected error: %v", tt.relativeTime, err)
				return
			}

			if tt.relativeTime == "" {
				if result != nil {
					t.Errorf("parseRelativeTime(%q) expected nil result, got %v", tt.relativeTime, result)
				}
				return
			}

			// Check that the result is within a reasonable range of now - duration
			expected := now.Add(-tt.duration)
			diff := result.Sub(expected)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("parseRelativeTime(%q) result not within 1 second of expected: got %v, expected around %v",
					tt.relativeTime, result, expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "test",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			item:     "test",
			expected: false,
		},
		{
			name:     "item exists",
			slice:    []string{"a", "b", "test", "c"},
			item:     "test",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "test",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    []string{"a", "b", "", "c"},
			item:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains(%v, %q) = %v, expected %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}
