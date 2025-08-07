package main

import (
	"testing"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"a4bc2d5e", "aaaabccddddde", false},
		{"abcd", "abcd", false},
		{"", "", false},
		{"45", "", true},
		{"aaa0b", "aab", false},
		{"\\4\\5", "45", false},
		{"qwe\\4\\5", "qwe45", false},
		{"qwe\\45", "qwe44444", false},
		{"qwe\\\\5", "qwe\\\\\\", false},
		{"qwe\\", "", true},
		{"a10", "", true},
	}

	for _, tt := range tests {
		got, err := Unpack(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("Unpack(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
		}
		if got != tt.expected {
			t.Errorf("Unpack(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}
