package net

import (
	"strings"
	"testing"
)

func TestProcessColors_ANSIEnabled(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bright red and reset",
			input: "&RHello&D",
			want:  "\033[1;31mHello\033[0m",
		},
		{
			name:  "green then blue",
			input: "&GGreen &BBlue",
			want:  "\033[1;32mGreen \033[1;34mBlue",
		},
		{
			name:  "escaped ampersand",
			input: "&&literal",
			want:  "&literal",
		},
		{
			name:  "no codes fast path",
			input: "no codes",
			want:  "no codes",
		},
		{
			name:  "reset only",
			input: "&D",
			want:  "\033[0m",
		},
		{
			name:  "two codes in sequence",
			input: "&W&R",
			want:  "\033[1;37m\033[1;31m",
		},
		{
			name:  "trailing ampersand no next char",
			input: "hello&",
			want:  "hello&",
		},
		{
			name:  "unknown code literal passthrough",
			input: "&ZStuff",
			want:  "&ZStuff",
		},
		// Bright / bold foreground codes
		{name: "code R", input: "&R", want: "\033[1;31m"},
		{name: "code G", input: "&G", want: "\033[1;32m"},
		{name: "code Y", input: "&Y", want: "\033[1;33m"},
		{name: "code B", input: "&B", want: "\033[1;34m"},
		{name: "code P", input: "&P", want: "\033[1;35m"},
		{name: "code C", input: "&C", want: "\033[1;36m"},
		{name: "code W", input: "&W", want: "\033[1;37m"},
		{name: "code O", input: "&O", want: "\033[0;33m"},
		// Dark / normal foreground codes
		{name: "code r", input: "&r", want: "\033[0;31m"},
		{name: "code g", input: "&g", want: "\033[0;32m"},
		{name: "code y", input: "&y", want: "\033[0;33m"},
		{name: "code b", input: "&b", want: "\033[0;34m"},
		{name: "code p", input: "&p", want: "\033[0;35m"},
		{name: "code c", input: "&c", want: "\033[0;36m"},
		{name: "code w", input: "&w", want: "\033[0;37m"},
		{name: "code o", input: "&o", want: "\033[0;33m"},
		// Reset codes
		{name: "code D reset", input: "&D", want: "\033[0m"},
		{name: "code d reset", input: "&d", want: "\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessColors(tt.input, true)
			if got != tt.want {
				t.Errorf("ProcessColors(%q, true) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestProcessColors_ANSIDisabled(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "codes stripped",
			input: "&RHello&D",
			want:  "Hello",
		},
		{
			name:  "escaped ampersand preserved",
			input: "&&literal",
			want:  "&literal",
		},
		{
			name:  "no codes unchanged",
			input: "no codes",
			want:  "no codes",
		},
		{
			name:  "two codes both stripped",
			input: "&W&R",
			want:  "",
		},
		{
			name:  "mixed codes and text stripped",
			input: "&GHello &Bworld&D",
			want:  "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessColors(tt.input, false)
			if got != tt.want {
				t.Errorf("ProcessColors(%q, false) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestProcessColors_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		ansiEnabled bool
		want        string
	}{
		{
			name:        "empty string",
			input:       "",
			ansiEnabled: true,
			want:        "",
		},
		{
			name:        "empty string ansi disabled",
			input:       "",
			ansiEnabled: false,
			want:        "",
		},
		{
			name:        "lone ampersand",
			input:       "&",
			ansiEnabled: true,
			want:        "&",
		},
		{
			name:        "lone ampersand ansi disabled",
			input:       "&",
			ansiEnabled: false,
			want:        "&",
		},
		{
			name:        "long string with many codes",
			input:       strings.Repeat("&RHi&D ", 500),
			ansiEnabled: true,
			want:        strings.Repeat("\033[1;31mHi\033[0m ", 500),
		},
		{
			name:        "long string codes stripped",
			input:       strings.Repeat("&RHi&D ", 500),
			ansiEnabled: false,
			want:        strings.Repeat("Hi ", 500),
		},
		{
			name:        "multiple escaped ampersands",
			input:       "&&&&&&",
			ansiEnabled: true,
			want:        "&&&",
		},
		{
			name:        "escaped ampersand before code",
			input:       "&&R",
			ansiEnabled: true,
			want:        "&R",
		},
		{
			name:        "code then trailing ampersand",
			input:       "&R&",
			ansiEnabled: true,
			want:        "\033[1;31m&",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessColors(tt.input, tt.ansiEnabled)
			if got != tt.want {
				t.Errorf("ProcessColors(%q, %v) = %q, want %q", tt.input, tt.ansiEnabled, got, tt.want)
			}
		})
	}
}
