package util

import (
	"math"
	"testing"
)

const iterations = 10000

// ---------------------------------------------------------------------------
// NumberRange
// ---------------------------------------------------------------------------

func TestNumberRange_FromEqualsTo(t *testing.T) {
	for i := 0; i < 100; i++ {
		if got := NumberRange(5, 5); got != 5 {
			t.Fatalf("NumberRange(5,5) = %d, want 5", got)
		}
	}
}

func TestNumberRange_FromGreaterThanTo(t *testing.T) {
	for i := 0; i < 100; i++ {
		if got := NumberRange(10, 3); got != 10 {
			t.Fatalf("NumberRange(10,3) = %d, want 10", got)
		}
	}
}

func TestNumberRange_Statistical(t *testing.T) {
	counts := make(map[int]int)
	sum := 0
	for i := 0; i < iterations; i++ {
		v := NumberRange(1, 6)
		if v < 1 || v > 6 {
			t.Fatalf("NumberRange(1,6) = %d, out of bounds [1,6]", v)
		}
		counts[v]++
		sum += v
	}

	mean := float64(sum) / float64(iterations)
	if math.Abs(mean-3.5) > 0.15 {
		t.Errorf("mean of NumberRange(1,6) over %d trials = %.4f, want ~3.5", iterations, mean)
	}

	// Every face should appear at least once in 10000 rolls.
	for face := 1; face <= 6; face++ {
		if counts[face] == 0 {
			t.Errorf("face %d never appeared in %d trials", face, iterations)
		}
	}

	// Verify there is variance: not all values are the same.
	if len(counts) < 2 {
		t.Errorf("only %d distinct values produced — expected variance", len(counts))
	}
}

// ---------------------------------------------------------------------------
// NumberPercent
// ---------------------------------------------------------------------------

func TestNumberPercent(t *testing.T) {
	seen := make(map[int]bool)
	for i := 0; i < iterations; i++ {
		v := NumberPercent()
		if v < 1 || v > 100 {
			t.Fatalf("NumberPercent() = %d, out of bounds [1,100]", v)
		}
		seen[v] = true
	}
	// With 10000 trials and 100 buckets, every value should appear.
	if len(seen) != 100 {
		t.Errorf("saw %d distinct values out of 100 expected", len(seen))
	}
}

// ---------------------------------------------------------------------------
// NumberDoor
// ---------------------------------------------------------------------------

func TestNumberDoor(t *testing.T) {
	seen := make(map[int]bool)
	for i := 0; i < iterations; i++ {
		v := NumberDoor()
		if v < 0 || v > 9 {
			t.Fatalf("NumberDoor() = %d, out of bounds [0,9]", v)
		}
		seen[v] = true
	}
	if len(seen) != 10 {
		t.Errorf("saw %d distinct values out of 10 expected", len(seen))
	}
}

// ---------------------------------------------------------------------------
// NumberBits
// ---------------------------------------------------------------------------

func TestNumberBits_WidthZero(t *testing.T) {
	for i := 0; i < 100; i++ {
		if got := NumberBits(0); got != 0 {
			t.Fatalf("NumberBits(0) = %d, want 0", got)
		}
	}
}

func TestNumberBits_WidthOne(t *testing.T) {
	seen := make(map[int]bool)
	for i := 0; i < iterations; i++ {
		v := NumberBits(1)
		if v != 0 && v != 1 {
			t.Fatalf("NumberBits(1) = %d, want 0 or 1", v)
		}
		seen[v] = true
	}
	if len(seen) != 2 {
		t.Errorf("NumberBits(1) produced %d distinct values, want 2", len(seen))
	}
}

func TestNumberBits_WidthEight(t *testing.T) {
	min, max := math.MaxInt, 0
	for i := 0; i < iterations; i++ {
		v := NumberBits(8)
		if v < 0 || v > 255 {
			t.Fatalf("NumberBits(8) = %d, out of bounds [0,255]", v)
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	// Statistical: with 10000 uniform draws from 256 values, the observed
	// range should be wide. Exact coverage is not guaranteed, but min
	// should be low and max should be high.
	if min > 10 {
		t.Errorf("NumberBits(8) minimum = %d, expected close to 0", min)
	}
	if max < 245 {
		t.Errorf("NumberBits(8) maximum = %d, expected close to 255", max)
	}
}

// ---------------------------------------------------------------------------
// DiceRoll
// ---------------------------------------------------------------------------

func TestDiceRoll_SizeZero(t *testing.T) {
	if got := DiceRoll(3, 0); got != 0 {
		t.Fatalf("DiceRoll(3,0) = %d, want 0", got)
	}
}

func TestDiceRoll_SizeOne(t *testing.T) {
	if got := DiceRoll(5, 1); got != 5 {
		t.Fatalf("DiceRoll(5,1) = %d, want 5", got)
	}
}

func TestDiceRoll_1d6(t *testing.T) {
	for i := 0; i < iterations; i++ {
		v := DiceRoll(1, 6)
		if v < 1 || v > 6 {
			t.Fatalf("DiceRoll(1,6) = %d, out of bounds [1,6]", v)
		}
	}
}

func TestDiceRoll_2d6_Statistical(t *testing.T) {
	sum := 0
	for i := 0; i < iterations; i++ {
		v := DiceRoll(2, 6)
		if v < 2 || v > 12 {
			t.Fatalf("DiceRoll(2,6) = %d, out of bounds [2,12]", v)
		}
		sum += v
	}
	mean := float64(sum) / float64(iterations)
	if math.Abs(mean-7.0) > 0.2 {
		t.Errorf("mean of DiceRoll(2,6) over %d trials = %.4f, want ~7.0", iterations, mean)
	}
}

// ---------------------------------------------------------------------------
// NumberFuzzy
// ---------------------------------------------------------------------------

func TestNumberFuzzy_Range(t *testing.T) {
	seen := make(map[int]bool)
	for i := 0; i < iterations; i++ {
		v := NumberFuzzy(10)
		if v < 9 || v > 11 {
			t.Fatalf("NumberFuzzy(10) = %d, out of bounds [9,11]", v)
		}
		seen[v] = true
	}
	if len(seen) != 3 {
		t.Errorf("NumberFuzzy(10) produced %d distinct values, want 3 (9, 10, 11)", len(seen))
	}
}

func TestNumberFuzzy_NeverBelowOne(t *testing.T) {
	for i := 0; i < iterations; i++ {
		v := NumberFuzzy(1)
		if v < 1 || v > 2 {
			t.Fatalf("NumberFuzzy(1) = %d, out of bounds [1,2]", v)
		}
	}
}

// UMIN/UMAX/URANGE tests are in strings_test.go
