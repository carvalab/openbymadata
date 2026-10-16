package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetters(t *testing.T) {
	m := map[string]interface{}{
		"s": "x", "f64": 1.5, "f32": float32(2.5), "i": 3, "i64": int64(4), "bad": []int{1},
	}

	assert.Equal(t, "x", GetString(m, "s"))
	assert.Equal(t, "", GetString(m, "f64"))
	assert.Equal(t, "", GetString(m, "missing"))

	for key, want := range map[string]float64{"f64": 1.5, "f32": 2.5, "i": 3, "i64": 4, "bad": 0, "missing": 0} {
		assert.Equal(t, want, GetFloat64(m, key), key)
	}
	for key, want := range map[string]int64{"f64": 1, "f32": 2, "i": 3, "i64": 4, "bad": 0, "missing": 0} {
		assert.Equal(t, want, GetInt64(m, key), key)
	}
}

func TestGetTime(t *testing.T) {
	tests := map[string]time.Time{
		"2026-10-16T09:57:00Z":  time.Date(2026, 10, 16, 9, 57, 0, 0, time.UTC),
		"2026-10-16T09:57:00":   time.Date(2026, 10, 16, 9, 57, 0, 0, time.UTC),
		"2026-10-16 09:57:00.0": time.Date(2026, 10, 16, 9, 57, 0, 0, time.UTC),
		"2026-12-18":            time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC),
		"18/12/2026":            {},
	}
	for in, want := range tests {
		assert.Equal(t, want, GetTime(map[string]interface{}{"d": in}, "d"), in)
	}
	assert.True(t, GetTime(map[string]interface{}{"d": 1.0}, "d").IsZero())
	assert.True(t, GetTime(nil, "d").IsZero())
}

func TestParseTradeTime(t *testing.T) {
	now := time.Now()

	got := ParseTradeTime("13:45:51")
	assert.Equal(t, now.YearDay(), got.YearDay())
	assert.Equal(t, 13, got.Hour())
	assert.Equal(t, 51, got.Second())

	got = ParseTradeTime("13:45")
	assert.Equal(t, 45, got.Minute())

	assert.Equal(t, time.Date(2026, 10, 16, 13, 45, 0, 0, time.UTC), ParseTradeTime("2026-10-16 13:45:00"))

	// Empty and unparseable values fall back to "now".
	assert.WithinDuration(t, now, ParseTradeTime(""), time.Minute)
	assert.WithinDuration(t, now, ParseTradeTime("garbage"), time.Minute)
}
