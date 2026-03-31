package util

import "testing"

func TestOneArgument(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFirst string
		wantRest  string
	}{
		{"normal word", "sword rest of line", "sword", "rest of line"},
		{"single-quoted arg", "'hello world' rest", "hello world", "rest"},
		{"double-quoted arg", `"hello world" rest`, "hello world", "rest"},
		{"leading whitespace", "   sword rest", "sword", "rest"},
		{"empty input", "", "", ""},
		{"multi-word rest", "get sword from chest", "get", "sword from chest"},
		{"lowercases first arg", "SWORD rest", "sword", "rest"},
		{"single word no rest", "sword", "sword", ""},
		{"quoted no rest", "'hello world'", "hello world", ""},
		{"only whitespace", "   ", "", ""},
		{"mixed case quoted", `"HeLLo WoRLd" rest`, "hello world", "rest"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, rest := OneArgument(tt.input)
			if first != tt.wantFirst {
				t.Errorf("first = %q, want %q", first, tt.wantFirst)
			}
			if rest != tt.wantRest {
				t.Errorf("rest = %q, want %q", rest, tt.wantRest)
			}
		})
	}
}

func TestSmashTilde(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no tildes", "hello world", "hello world"},
		{"one tilde", "hello~world", "hello-world"},
		{"multiple tildes", "a~b~c~d", "a-b-c-d"},
		{"tilde-only string", "~~~", "---"},
		{"empty string", "", ""},
		{"tilde at edges", "~hello~", "-hello-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SmashTilde(tt.input)
			if got != tt.want {
				t.Errorf("SmashTilde(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "hello world", "Hello world"},
		{"empty", "", ""},
		{"already capitalized", "Hello", "Hello"},
		{"all caps", "HELLO WORLD", "Hello world"},
		{"single char lower", "h", "H"},
		{"single char upper", "H", "H"},
		{"mixed case", "hELLO wORLD", "Hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Capitalize(tt.input)
			if got != tt.want {
				t.Errorf("Capitalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsName(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		namelist string
		want     bool
	}{
		{"prefix match", "swo", "sword shield", true},
		{"exact match", "sword", "sword shield", true},
		{"no match", "axe", "sword shield", false},
		{"empty str", "", "sword shield", false},
		{"empty namelist", "sword", "", false},
		{"multi-word namelist", "shi", "sword shield armor", true},
		{"case insensitive str upper", "SWORD", "sword shield", true},
		{"case insensitive namelist upper", "sword", "SWORD SHIELD", true},
		{"prefix of second word", "sh", "sword shield", true},
		{"single word namelist exact", "sword", "sword", true},
		{"single word namelist prefix", "swo", "sword", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsName(tt.str, tt.namelist)
			if got != tt.want {
				t.Errorf("IsName(%q, %q) = %v, want %v", tt.str, tt.namelist, got, tt.want)
			}
		})
	}
}

func TestIsNameExact(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		namelist string
		want     bool
	}{
		{"exact match", "sword", "sword shield", true},
		{"prefix does not count", "swo", "sword shield", false},
		{"empty str", "", "sword shield", false},
		{"empty namelist", "sword", "", false},
		{"case insensitive", "SWORD", "sword shield", true},
		{"second word match", "shield", "sword shield", true},
		{"no match", "axe", "sword shield", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNameExact(tt.str, tt.namelist)
			if got != tt.want {
				t.Errorf("IsNameExact(%q, %q) = %v, want %v", tt.str, tt.namelist, got, tt.want)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"positive", "123", true},
		{"negative", "-456", true},
		{"zero", "0", true},
		{"non-numeric", "abc", false},
		{"empty", "", false},
		{"just minus", "-", false},
		{"mixed 123abc", "123abc", false},
		{"leading zero", "007", true},
		{"negative zero", "-0", true},
		{"space in middle", "1 2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNumber(tt.input)
			if got != tt.want {
				t.Errorf("IsNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNumberArgument(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNum int
		wantArg string
	}{
		{"dotted number prefix", "3.sword", 3, "sword"},
		{"no dot defaults to 1", "sword", 1, "sword"},
		{"zero prefix", "0.sword", 0, "sword"},
		{"non-numeric prefix", "abc.sword", 1, "abc.sword"},
		{"negative prefix", "-1.sword", -1, "sword"},
		{"large number", "99.potion", 99, "potion"},
		{"dot at end", "3.", 3, ""},
		{"empty before dot", ".sword", 1, ".sword"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, arg := NumberArgument(tt.input)
			if num != tt.wantNum {
				t.Errorf("NumberArgument(%q) number = %d, want %d", tt.input, num, tt.wantNum)
			}
			if arg != tt.wantArg {
				t.Errorf("NumberArgument(%q) arg = %q, want %q", tt.input, arg, tt.wantArg)
			}
		})
	}
}

func TestUMIN(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"a smaller", 1, 5, 1},
		{"b smaller", 5, 1, 1},
		{"equal", 3, 3, 3},
		{"negatives", -10, -5, -10},
		{"mixed sign", -3, 3, -3},
		{"zeros", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UMIN(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("UMIN(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestUMAX(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"a larger", 5, 1, 5},
		{"b larger", 1, 5, 5},
		{"equal", 3, 3, 3},
		{"negatives", -10, -5, -5},
		{"mixed sign", -3, 3, 3},
		{"zeros", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UMAX(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("UMAX(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestURANGE(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c int
		want    int
	}{
		{"within range", 1, 5, 10, 5},
		{"below min", 1, -5, 10, 1},
		{"above max", 1, 15, 10, 10},
		{"at min boundary", 1, 1, 10, 1},
		{"at max boundary", 1, 10, 10, 10},
		{"all same", 5, 5, 5, 5},
		{"negatives within", -10, -5, -1, -5},
		{"negatives below", -10, -20, -1, -10},
		{"negatives above", -10, 0, -1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := URANGE(tt.a, tt.b, tt.c)
			if got != tt.want {
				t.Errorf("URANGE(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, got, tt.want)
			}
		})
	}
}
