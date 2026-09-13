package main

import "testing"

func TestParseCommands(t *testing.T) {
	tests := []struct {
		input, name, since, text string
		ok                       bool
	}{
		{"/start abc", "start", "abc", "", true}, {" /PAUSE ", "pause", "", "", true},
		{"keep it short", "steer", "", "keep it short", true}, {"/wat", "", "", "", false}, {" ", "", "", "", false},
	}
	for _, test := range tests {
		got, ok := ParseCommand(test.input)
		if ok != test.ok || got.Name != test.name || got.Since != test.since || got.Text != test.text {
			t.Errorf("ParseCommand(%q) = %#v, %v", test.input, got, ok)
		}
	}
}
