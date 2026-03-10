package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  john IS world  ",
			expected: []string{"john", "is", "world"},
		},
		{
			input:    "  AAAHHH NOOO WHYY  world  ",
			expected: []string{"aaahhh", "nooo", "whyy", "world"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q) returned %d words, expected %d", c.input, len(actual), len(c.expected))
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%q) returned word %q at index %d, expected %q", c.input, word, i, expectedWord)
			}
		}
	}
}
