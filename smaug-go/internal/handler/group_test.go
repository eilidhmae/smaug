package handler

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// TestIsSameGroup covers the equivalence relation ported from C
// `is_same_group` (`act_comm.c:4293-4301`). SMAUG's group graph is flat —
// followers point at a leader whose own Leader is nil — so transitive-
// through-a-chain must not be accepted (see TestIsSameGroup_TransitiveViaChain).
func TestIsSameGroup(t *testing.T) {
	a := &types.CharData{Name: "A"}
	b := &types.CharData{Name: "B"}

	tests := []struct {
		name string
		x, y *types.CharData
		want bool
	}{
		{"self (leaderless)", a, a, true},
		{"two distinct leaderless chars", a, b, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSameGroup(tc.x, tc.y); got != tc.want {
				t.Errorf("IsSameGroup(%s, %s) = %v, want %v",
					tc.x.Name, tc.y.Name, got, tc.want)
			}
		})
	}
}

func TestIsSameGroup_NilInputs(t *testing.T) {
	a := &types.CharData{Name: "A"}
	if IsSameGroup(nil, a) {
		t.Error("IsSameGroup(nil, a) should be false")
	}
	if IsSameGroup(a, nil) {
		t.Error("IsSameGroup(a, nil) should be false")
	}
	if IsSameGroup(nil, nil) {
		t.Error("IsSameGroup(nil, nil) should be false")
	}
}

func TestIsSameGroup_LeaderAndFollower(t *testing.T) {
	leader := &types.CharData{Name: "Leader"}
	follower := &types.CharData{Name: "Follower", Leader: leader}

	if !IsSameGroup(leader, follower) {
		t.Error("leader and follower should be in the same group")
	}
	if !IsSameGroup(follower, leader) {
		t.Error("same-group relation must be symmetric")
	}
}

func TestIsSameGroup_TwoFollowersOfSameLeader(t *testing.T) {
	leader := &types.CharData{Name: "Leader"}
	f1 := &types.CharData{Name: "F1", Leader: leader}
	f2 := &types.CharData{Name: "F2", Leader: leader}

	if !IsSameGroup(f1, f2) {
		t.Error("two followers of the same leader should be in the same group")
	}
	if !IsSameGroup(f2, f1) {
		t.Error("symmetry: two followers of the same leader")
	}
}

func TestIsSameGroup_FollowersOfDifferentLeaders(t *testing.T) {
	l1 := &types.CharData{Name: "L1"}
	l2 := &types.CharData{Name: "L2"}
	f1 := &types.CharData{Name: "F1", Leader: l1}
	f2 := &types.CharData{Name: "F2", Leader: l2}

	if IsSameGroup(f1, f2) {
		t.Error("followers of different leaders should NOT be in the same group")
	}
	if IsSameGroup(l1, l2) {
		t.Error("two leaders of different groups should NOT be in the same group")
	}
	if IsSameGroup(l1, f2) {
		t.Error("leader of group 1 should NOT be in the same group as a follower of group 2")
	}
}

// TestIsSameGroup_TransitiveViaChain documents that SMAUG's leader graph is
// flat — C's `is_same_group` does not recurse. If A leads B and B's Leader
// field is reassigned to point at C as if C followed B, the predicate uses
// B's direct Leader (C), not A. Modeling that here: leader=A has no Leader,
// B.Leader = A, but if someone sets C.Leader = B (B is not a leader), then
// is_same_group(A, C) examines C.Leader=B vs A.Leader=nil and returns false.
// This mirrors C behavior — callers must collapse chains before calling.
func TestIsSameGroup_TransitiveViaChain(t *testing.T) {
	a := &types.CharData{Name: "A"}
	b := &types.CharData{Name: "B", Leader: a}
	c := &types.CharData{Name: "C", Leader: b}

	// b is in group with a (direct follower).
	if !IsSameGroup(a, b) {
		t.Error("a and b must be same group")
	}
	// c's Leader is b (not a), and b is not its own leader, so
	// normalizing c -> b, a -> a gives b != a — NOT same group.
	// This matches C exactly.
	if IsSameGroup(a, c) {
		t.Error("C-matching: a and c must NOT be same group when Leader graph is not collapsed")
	}
	// b and c: normalize c -> b, b -> a -> a. b != a — not same group.
	if IsSameGroup(b, c) {
		t.Error("C-matching: b and c must NOT be same group under flat-graph semantics")
	}
}
