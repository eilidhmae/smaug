package handler

import "github.com/eilidhmae/smaug/internal/types"

// IsSameGroup reports whether two characters belong to the same adventuring
// group. Ports C `is_same_group` from `act_comm.c:4293-4301`: normalize each
// character to its group leader (itself if `Leader == nil`) and compare the
// two leaders by pointer identity. This is an equivalence relation —
// reflexive, symmetric, and transitive within the flat leader graph SMAUG
// uses (followers point at the leader; the leader's own `Leader` is nil).
//
// Nil inputs always return false. A character is always in the same group
// as itself (including self-pairs of NPCs, per C behavior).
func IsSameGroup(a, b *types.CharData) bool {
	if a == nil || b == nil {
		return false
	}
	la := a
	if a.Leader != nil {
		la = a.Leader
	}
	lb := b
	if b.Leader != nil {
		lb = b.Leader
	}
	return la == lb
}
