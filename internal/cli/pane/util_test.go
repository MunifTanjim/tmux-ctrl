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

func TestMoveReinvokeArgs(t *testing.T) {
	cases := []struct {
		name                    string
		paneID, direction, size string
		target                  string
		edge                    bool
		expected                string
	}{
		{name: "relative", paneID: "%3", direction: "left", expected: `"tmux-ctrl" pane move -p %3 -d left`},
		{name: "edge", paneID: "%3", direction: "top", edge: true, expected: `"tmux-ctrl" pane move -p %3 -d top --edge`},
		{name: "corner", paneID: "%3", direction: "top-left", edge: true, expected: `"tmux-ctrl" pane move -p %3 -d top-left --edge`},
		{name: "swap", paneID: "%3", direction: "swap", expected: `"tmux-ctrl" pane move -p %3 -d swap`},
		{name: "with size", paneID: "%3", direction: "left", size: "30%", expected: `"tmux-ctrl" pane move -p %3 -d left --size 30%`},
		{name: "picker template", direction: "left", target: "%%", expected: `"tmux-ctrl" pane move -d left --target %%`},
		{name: "picker src option", paneID: "#{q:@tmux_ctrl_pane_move_src}", direction: "left", target: "%%", expected: `"tmux-ctrl" pane move -p #{q:@tmux_ctrl_pane_move_src} -d left --target %%`},
		{name: "size and target", paneID: "%3", direction: "left", size: "30%", target: "%%", expected: `"tmux-ctrl" pane move -p %3 -d left --size 30% --target %%`},
	}
	for _, c := range cases {
		got := moveReinvokeArgs("tmux-ctrl", c.paneID, c.direction, c.size, c.target, c.edge)
		if got != c.expected {
			t.Errorf("%s: moveReinvokeArgs = %q, want %q", c.name, got, c.expected)
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
