package util

import "testing"

func TestNumPunct_Small(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{99, "99"},
		{100, "100"},
		{123, "123"},
		{999, "999"},
	}
	for _, c := range cases {
		if got := NumPunct(c.in); got != c.want {
			t.Errorf("NumPunct(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNumPunct_Thousand(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{1000, "1,000"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
		{999999999, "999,999,999"},
	}
	for _, c := range cases {
		if got := NumPunct(c.in); got != c.want {
			t.Errorf("NumPunct(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNumPunct_Negative(t *testing.T) {
	if got := NumPunct(-1234); got != "-1,234" {
		t.Errorf("NumPunct(-1234) = %q, want -1,234", got)
	}
	if got := NumPunct(-999); got != "-999" {
		t.Errorf("NumPunct(-999) = %q, want -999", got)
	}
}

func TestNumPunct_Max(t *testing.T) {
	if got := NumPunct(2000000000); got != "2,000,000,000" {
		t.Errorf("NumPunct(2B) = %q, want 2,000,000,000", got)
	}
}
