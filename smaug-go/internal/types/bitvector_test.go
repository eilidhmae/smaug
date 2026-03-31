package types

import (
	"testing"
)

// ---------------------------------------------------------------------------
// 1. IsSet / Set / Remove / Toggle
// ---------------------------------------------------------------------------

func TestBitVector_IsSet_Set(t *testing.T) {
	tests := []struct {
		name string
		bit  int
		want bool // expected IsSet result after Set
	}{
		{"bit 0", 0, true},
		{"bit 31 (boundary first element)", 31, true},
		{"bit 32 (first bit second element)", 32, true},
		{"bit 127 (last valid bit)", 127, true},
		{"bit -1 (out of range)", -1, false},
		{"bit 128 (out of range)", 128, false},
		{"bit 999 (out of range)", 999, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bv BitVector
			bv.Set(tc.bit)
			got := bv.IsSet(tc.bit)
			if got != tc.want {
				t.Errorf("after Set(%d): IsSet = %v, want %v", tc.bit, got, tc.want)
			}
			// For valid bits, verify only the intended bit changed.
			if tc.want {
				for i := 0; i < MaxBits; i++ {
					if i == tc.bit {
						continue
					}
					if bv.IsSet(i) {
						t.Errorf("bit %d unexpectedly set after Set(%d)", i, tc.bit)
					}
				}
			}
		})
	}
}

func TestBitVector_Remove(t *testing.T) {
	tests := []struct {
		name string
		bit  int
	}{
		{"bit 0", 0},
		{"bit 31", 31},
		{"bit 32", 32},
		{"bit 127", 127},
		{"bit -1 (out of range)", -1},
		{"bit 128 (out of range)", 128},
		{"bit 999 (out of range)", 999},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bv BitVector
			bv.Set(tc.bit) // may be no-op for OOB
			bv.Remove(tc.bit)
			if bv.IsSet(tc.bit) {
				t.Errorf("bit %d still set after Remove", tc.bit)
			}
			if !bv.IsEmpty() {
				t.Errorf("bitvector not empty after Set+Remove on single bit %d", tc.bit)
			}
		})
	}
}

func TestBitVector_Toggle(t *testing.T) {
	tests := []struct {
		name    string
		bit     int
		inRange bool
	}{
		{"bit 0", 0, true},
		{"bit 31", 31, true},
		{"bit 32", 32, true},
		{"bit 127", 127, true},
		{"bit -1 (out of range)", -1, false},
		{"bit 128 (out of range)", 128, false},
		{"bit 999 (out of range)", 999, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bv BitVector

			// First toggle: 0 -> 1 (for valid bits)
			bv.Toggle(tc.bit)
			if tc.inRange {
				if !bv.IsSet(tc.bit) {
					t.Errorf("bit %d not set after first Toggle", tc.bit)
				}
			} else {
				if !bv.IsEmpty() {
					t.Errorf("bitvector changed after Toggle on out-of-range bit %d", tc.bit)
				}
			}

			// Second toggle: 1 -> 0
			bv.Toggle(tc.bit)
			if bv.IsSet(tc.bit) {
				t.Errorf("bit %d still set after second Toggle", tc.bit)
			}
			if !bv.IsEmpty() {
				t.Errorf("bitvector not empty after double Toggle on bit %d", tc.bit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Clear
// ---------------------------------------------------------------------------

func TestBitVector_Clear(t *testing.T) {
	var bv BitVector
	bv.Set(0)
	bv.Set(31)
	bv.Set(32)
	bv.Set(64)
	bv.Set(127)

	bv.Clear()

	if !bv.IsEmpty() {
		t.Errorf("bitvector not empty after Clear: %s", bv.String())
	}
	for i := 0; i < MaxBits; i++ {
		if bv.IsSet(i) {
			t.Errorf("bit %d set after Clear", i)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// 3. IsEmpty
// ---------------------------------------------------------------------------

func TestBitVector_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		bv   BitVector
		want bool
	}{
		{"zero vector", BitVector{}, true},
		{"bit 0 set", BitVector{1, 0, 0, 0}, false},
		{"bit in element 3", BitVector{0, 0, 0, 1}, false},
		{"all elements nonzero", BitVector{1, 2, 3, 4}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.bv.IsEmpty(); got != tc.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. Equal
// ---------------------------------------------------------------------------

func TestBitVector_Equal(t *testing.T) {
	tests := []struct {
		name string
		a, b BitVector
		want bool
	}{
		{"both zero", BitVector{}, BitVector{}, true},
		{"identical non-zero", BitVector{1, 2, 3, 4}, BitVector{1, 2, 3, 4}, true},
		{"differ in element 0", BitVector{1, 0, 0, 0}, BitVector{2, 0, 0, 0}, false},
		{"differ in element 3", BitVector{0, 0, 0, 1}, BitVector{0, 0, 0, 2}, false},
		{"one zero one not", BitVector{}, BitVector{0, 0, 0, 1}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Equal(tc.b); got != tc.want {
				t.Errorf("Equal() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. Or / And
// ---------------------------------------------------------------------------

func TestBitVector_Or(t *testing.T) {
	tests := []struct {
		name   string
		a, b   BitVector
		expect BitVector
	}{
		{"both zero", BitVector{}, BitVector{}, BitVector{}},
		{"a zero", BitVector{}, BitVector{0xFF, 0, 0, 0}, BitVector{0xFF, 0, 0, 0}},
		{"disjoint bits", BitVector{0x0F, 0, 0, 0}, BitVector{0xF0, 0, 0, 0}, BitVector{0xFF, 0, 0, 0}},
		{"overlapping bits", BitVector{0xFF, 0, 0, 0}, BitVector{0xF0, 0, 0, 0}, BitVector{0xFF, 0, 0, 0}},
		{"all elements",
			BitVector{1, 2, 4, 8},
			BitVector{16, 32, 64, 128},
			BitVector{17, 34, 68, 136}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Or(tc.b)
			if got != tc.expect {
				t.Errorf("Or() = %v, want %v", got, tc.expect)
			}
		})
	}
}

func TestBitVector_And(t *testing.T) {
	tests := []struct {
		name   string
		a, b   BitVector
		expect BitVector
	}{
		{"both zero", BitVector{}, BitVector{}, BitVector{}},
		{"no overlap", BitVector{0x0F, 0, 0, 0}, BitVector{0xF0, 0, 0, 0}, BitVector{}},
		{"full overlap", BitVector{0xFF, 0, 0, 0}, BitVector{0xFF, 0, 0, 0}, BitVector{0xFF, 0, 0, 0}},
		{"partial overlap", BitVector{0xFF, 0, 0, 0}, BitVector{0xF0, 0, 0, 0}, BitVector{0xF0, 0, 0, 0}},
		{"all elements",
			BitVector{0xFF, 0xFF, 0xFF, 0xFF},
			BitVector{0x0F, 0xF0, 0x33, 0xCC},
			BitVector{0x0F, 0xF0, 0x33, 0xCC}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.And(tc.b)
			if got != tc.expect {
				t.Errorf("And() = %v, want %v", got, tc.expect)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 6. HasAny
// ---------------------------------------------------------------------------

func TestBitVector_HasAny(t *testing.T) {
	tests := []struct {
		name string
		a, b BitVector
		want bool
	}{
		{"both zero", BitVector{}, BitVector{}, false},
		{"no overlap", BitVector{0x0F, 0, 0, 0}, BitVector{0xF0, 0, 0, 0}, false},
		{"overlap in element 0", BitVector{0xFF, 0, 0, 0}, BitVector{0x01, 0, 0, 0}, true},
		{"overlap in element 3", BitVector{0, 0, 0, 0x80000000}, BitVector{0, 0, 0, 0x80000000}, true},
		{"disjoint across elements", BitVector{1, 0, 0, 0}, BitVector{0, 1, 0, 0}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.HasAny(tc.b); got != tc.want {
				t.Errorf("HasAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. SetFrom
// ---------------------------------------------------------------------------

func TestBitVector_SetFrom(t *testing.T) {
	var bv BitVector
	bv.Set(0)
	bv.Set(64)

	other := BitVector{0, 0xFF, 0, 0x01}
	bv.SetFrom(other)

	expect := BitVector{1, 0xFF, 1, 0x01}
	if bv != expect {
		t.Errorf("SetFrom: got %v, want %v", bv, expect)
	}

	// SetFrom with zero should be a no-op.
	before := bv
	bv.SetFrom(BitVector{})
	if bv != before {
		t.Errorf("SetFrom(zero) changed bitvector")
	}
}

// ---------------------------------------------------------------------------
// 8. String
// ---------------------------------------------------------------------------

func TestBitVector_String(t *testing.T) {
	tests := []struct {
		name string
		bv   BitVector
		want string
	}{
		{"zero vector", BitVector{}, "0 0 0 0"},
		{"single element", BitVector{42, 0, 0, 0}, "42 0 0 0"},
		{"all elements", BitVector{1, 2, 3, 4}, "1 2 3 4"},
		{"max uint32", BitVector{0xFFFFFFFF, 0, 0, 0}, "4294967295 0 0 0"},
		{"bit 31 set", BitVector{0x80000000, 0, 0, 0}, "2147483648 0 0 0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.bv.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 9. ParseBitVector
// ---------------------------------------------------------------------------

func TestBitVector_ParseBitVector(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    BitVector
		wantErr bool
	}{
		{"empty string", "", BitVector{}, false},
		{"whitespace only", "   ", BitVector{}, false},
		{"single value", "42", BitVector{42, 0, 0, 0}, false},
		{"two values", "1 2", BitVector{1, 2, 0, 0}, false},
		{"four values", "10 20 30 40", BitVector{10, 20, 30, 40}, false},
		{"extra values beyond 4", "1 2 3 4 5 6", BitVector{1, 2, 3, 4}, false},
		{"max uint32", "4294967295 0 0 0", BitVector{0xFFFFFFFF, 0, 0, 0}, false},
		{"leading/trailing spaces", "  1 2 3 4  ", BitVector{1, 2, 3, 4}, false},
		{"invalid non-numeric", "abc", BitVector{}, true},
		{"invalid second element", "1 abc", BitVector{}, true},
		{"negative number", "-1", BitVector{}, true},
		{"overflow uint32", "4294967296", BitVector{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseBitVector(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseBitVector(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("ParseBitVector(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 10. BitVectorFromInt
// ---------------------------------------------------------------------------

func TestBitVector_BitVectorFromInt(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		want  BitVector
	}{
		{"zero", 0, BitVector{}},
		{"one", 1, BitVector{1, 0, 0, 0}},
		{"max", 0xFFFFFFFF, BitVector{0xFFFFFFFF, 0, 0, 0}},
		{"arbitrary", 0xDEADBEEF, BitVector{0xDEADBEEF, 0, 0, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BitVectorFromInt(tc.flags)
			if got != tc.want {
				t.Errorf("BitVectorFromInt(%#x) = %v, want %v", tc.flags, got, tc.want)
			}
			// Verify other elements are zero.
			if got[1] != 0 || got[2] != 0 || got[3] != 0 {
				t.Errorf("BitVectorFromInt(%#x) has non-zero elements beyond [0]", tc.flags)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 11. Bits
// ---------------------------------------------------------------------------

func TestBitVector_Bits(t *testing.T) {
	tests := []struct {
		name string
		bv   BitVector
		want []int
	}{
		{"zero vector", BitVector{}, nil},
		{"bit 0 only", BitVector{1, 0, 0, 0}, []int{0}},
		{"bits 0 and 1", BitVector{3, 0, 0, 0}, []int{0, 1}},
		{"bit 31", BitVector{0x80000000, 0, 0, 0}, []int{31}},
		{"bit 32", BitVector{0, 1, 0, 0}, []int{32}},
		{"bit 127", BitVector{0, 0, 0, 0x80000000}, []int{127}},
		{"scattered across elements",
			BitVector{1, 1, 1, 1},
			[]int{0, 32, 64, 96}},
		{"multiple bits per element",
			BitVector{0x05, 0, 0, 0}, // bits 0 and 2
			[]int{0, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.bv.Bits()
			if !intSliceEqual(got, tc.want) {
				t.Errorf("Bits() = %v, want %v", got, tc.want)
			}
		})
	}
}

// intSliceEqual compares two int slices, treating nil and empty as equal.
func intSliceEqual(a, b []int) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
