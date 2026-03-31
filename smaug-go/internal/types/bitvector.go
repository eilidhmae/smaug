package types

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// IntBits is the number of bits per element, matching C INTBITS.
	IntBits = 32

	// XBI is the number of uint32 elements in a BitVector, matching C XBI.
	XBI = 4

	// MaxBits is the total number of usable bits (XBI * IntBits).
	MaxBits = XBI * IntBits // 128
)

// BitVector is a 128-bit extended bitvector matching the C EXT_BV struct.
// Layout is identical to the C struct: an array of XBI (4) unsigned 32-bit
// integers, with bit indexing via bits[bit/32] & (1 << (bit%32)).
type BitVector [XBI]uint32

// IsSet returns true if the given bit position is set.
// Out-of-range bits always return false.
func (bv BitVector) IsSet(bit int) bool {
	if bit < 0 || bit >= MaxBits {
		return false
	}
	return bv[bit/IntBits]&(1<<uint(bit%IntBits)) != 0
}

// Set sets the given bit position. Out-of-range bits are ignored.
func (bv *BitVector) Set(bit int) {
	if bit < 0 || bit >= MaxBits {
		return
	}
	bv[bit/IntBits] |= 1 << uint(bit%IntBits)
}

// Remove clears the given bit position. Out-of-range bits are ignored.
func (bv *BitVector) Remove(bit int) {
	if bit < 0 || bit >= MaxBits {
		return
	}
	bv[bit/IntBits] &^= 1 << uint(bit%IntBits)
}

// Toggle flips the given bit position. Out-of-range bits are ignored.
func (bv *BitVector) Toggle(bit int) {
	if bit < 0 || bit >= MaxBits {
		return
	}
	bv[bit/IntBits] ^= 1 << uint(bit%IntBits)
}

// Clear zeroes all bits.
func (bv *BitVector) Clear() {
	bv[0] = 0
	bv[1] = 0
	bv[2] = 0
	bv[3] = 0
}

// IsEmpty returns true if all bits are zero.
func (bv BitVector) IsEmpty() bool {
	return bv[0] == 0 && bv[1] == 0 && bv[2] == 0 && bv[3] == 0
}

// Equal returns true if both bitvectors have exactly the same bits set.
func (bv BitVector) Equal(other BitVector) bool {
	return bv == other
}

// Or returns a new BitVector that is the bitwise OR of bv and other.
func (bv BitVector) Or(other BitVector) BitVector {
	return BitVector{
		bv[0] | other[0],
		bv[1] | other[1],
		bv[2] | other[2],
		bv[3] | other[3],
	}
}

// And returns a new BitVector that is the bitwise AND of bv and other.
func (bv BitVector) And(other BitVector) BitVector {
	return BitVector{
		bv[0] & other[0],
		bv[1] & other[1],
		bv[2] & other[2],
		bv[3] & other[3],
	}
}

// HasAny returns true if bv and other share any set bits.
func (bv BitVector) HasAny(other BitVector) bool {
	return (bv[0]&other[0])|(bv[1]&other[1])|(bv[2]&other[2])|(bv[3]&other[3]) != 0
}

// SetFrom copies all set bits from other into bv (bitwise OR in place),
// matching the C xSET_BITS macro behaviour.
func (bv *BitVector) SetFrom(other BitVector) {
	bv[0] |= other[0]
	bv[1] |= other[1]
	bv[2] |= other[2]
	bv[3] |= other[3]
}

// String serialises the bitvector as space-separated decimal uint32 values,
// matching the C save format used in area files. Trailing zero elements are
// always included so the output is always exactly four values.
func (bv BitVector) String() string {
	return fmt.Sprintf("%d %d %d %d", bv[0], bv[1], bv[2], bv[3])
}

// ParseBitVector parses a bitvector from space-separated decimal uint32 values
// as written by the C codebase. Fewer than XBI values is permitted; missing
// elements default to zero. Extra values beyond XBI are silently ignored.
func ParseBitVector(s string) (BitVector, error) {
	var bv BitVector
	s = strings.TrimSpace(s)
	if s == "" {
		return bv, nil
	}
	fields := strings.Fields(s)
	for i := 0; i < len(fields) && i < XBI; i++ {
		v, err := strconv.ParseUint(fields[i], 10, 32)
		if err != nil {
			return bv, fmt.Errorf("bitvector element %d: %w", i, err)
		}
		bv[i] = uint32(v)
	}
	return bv, nil
}

// BitVectorFromInt converts a single 32-bit flag value (used by non-extended
// bitvectors in the C codebase) into a BitVector. The value is placed in the
// first element; all other elements are zeroed.
func BitVectorFromInt(flags uint32) BitVector {
	return BitVector{flags, 0, 0, 0}
}

// Bits returns a slice of all bit positions that are currently set,
// in ascending order. Useful for iteration and debugging.
func (bv BitVector) Bits() []int {
	var set []int
	for i := 0; i < MaxBits; i++ {
		if bv[i/IntBits]&(1<<uint(i%IntBits)) != 0 {
			set = append(set, i)
		}
	}
	return set
}
