package util

import (
	crand "crypto/rand"
	"math/rand/v2"
)

// rng is the package-level random number generator, seeded from crypto/rand.
var rng *rand.Rand

func init() {
	var seed [32]byte
	if _, err := crand.Read(seed[:]); err != nil {
		panic("util: failed to seed RNG from crypto/rand: " + err.Error())
	}
	rng = rand.New(rand.NewChaCha8(seed))
}

// NumberRange returns a random integer in the inclusive range [from, to].
// If from >= to it returns from. Mirrors number_range() from db.c.
func NumberRange(from, to int) int {
	if to-from < 1 {
		return from
	}
	return rng.IntN(to-from+1) + from
}

// NumberPercent returns a random integer in the range [1, 100].
// Mirrors number_percent() from db.c.
func NumberPercent() int {
	return rng.IntN(100) + 1
}

// NumberDoor returns a random integer in the range [0, 9], covering the
// ten possible exit directions (N/S/E/W/U/D + diagonals).
// Mirrors number_door() from db.c.
func NumberDoor() int {
	return rng.IntN(10)
}

// NumberBits returns a random non-negative integer using only the lowest
// 'width' bits (i.e. in range [0, 2^width - 1]).
// Mirrors number_bits() from db.c.
func NumberBits(width int) int {
	if width <= 0 {
		return 0
	}
	return int(rng.Uint64() & uint64((1<<width)-1))
}

// DiceRoll rolls 'number' dice each with 'size' sides and returns the sum.
// Mirrors dice() from db.c.
func DiceRoll(number, size int) int {
	switch size {
	case 0:
		return 0
	case 1:
		return number
	}
	sum := 0
	for i := 0; i < number; i++ {
		sum += NumberRange(1, size)
	}
	return sum
}

// NumberFuzzy returns number-1, number, or number+1 at random, but never
// less than 1. Mirrors number_fuzzy() from db.c.
func NumberFuzzy(number int) int {
	switch NumberBits(2) {
	case 0:
		number--
	case 3:
		number++
	}
	return UMAX(1, number)
}
