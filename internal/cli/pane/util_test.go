package pane

import "testing"

func TestFractionToPercent(t *testing.T) {
	cases := []struct {
		size         string
		expectedSize string
		expectedOK   bool
	}{
		{size: "0.3", expectedSize: "30%", expectedOK: true},
		{size: ".5", expectedSize: "50%", expectedOK: true},
		{size: "0.29", expectedSize: "29%", expectedOK: true},
		{size: "0.755", expectedSize: "76%", expectedOK: true},
		{size: "1", expectedOK: false},
		{size: "0", expectedOK: false},
		{size: "5", expectedOK: false},
		{size: "1.5", expectedOK: false},
		{size: "abc", expectedOK: false},
	}
	for _, c := range cases {
		got, ok := fractionToPercent(c.size)
		if ok != c.expectedOK || (ok && got != c.expectedSize) {
			t.Errorf("fractionToPercent(%q) = (%q, %v), want (%q, %v)", c.size, got, ok, c.expectedSize, c.expectedOK)
		}
	}
}

func TestMaxSplitSize(t *testing.T) {
	cases := []struct {
		avail    int
		expected int
	}{
		{avail: 0, expected: 0},
		{avail: 1, expected: 0},
		{avail: 2, expected: 0},
		{avail: 3, expected: 1},
		{avail: 50, expected: 48},
	}
	for _, c := range cases {
		if got := maxSplitSize(c.avail); got != c.expected {
			t.Errorf("maxSplitSize(%d) = %d, want %d", c.avail, got, c.expected)
		}
	}
}
