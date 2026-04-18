package util

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// TestAct_ColorPrefix_AT_HIT verifies that an act() call with aType =
// AT_HIT prepends the mapped color code (&R) and appends a reset (&D)
// around the formatted message delivered to the recipient. Matches
// plan-tranche-c.md G2 acceptance criterion A8.
func TestAct_ColorPrefix_AT_HIT(t *testing.T) {
	ch, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	vch, vc := makeChar(t, "Bob", types.SEX_MALE)
	putInRoom(ch, vch)

	Act(types.AT_HIT, "$n hits $N", ch, vch, nil, nil, types.TO_VICT)

	got := readOutput(t, vch, vc)
	if !strings.HasPrefix(got, "&R") {
		t.Errorf("AT_HIT should prepend &R; got %q", got)
	}
	if !strings.Contains(got, "Alice hits Bob") {
		t.Errorf("message body should still contain formatted text; got %q", got)
	}
	// &D reset must appear before the trailing \n\r so subsequent lines
	// are not tinted red.
	if !strings.HasSuffix(got, "&D\n\r") {
		t.Errorf("output should end with &D\\n\\r; got %q", got)
	}
}

// TestAct_ColorPrefix_AT_HITME verifies the attacker/victim AT
// differentiation: the same two-call pattern with AT_HIT for TO_CHAR
// and AT_HITME for TO_VICT produces different prefixes per recipient.
func TestAct_ColorPrefix_AT_HITME(t *testing.T) {
	ch, cc := makeChar(t, "Alice", types.SEX_FEMALE)
	vch, vc := makeChar(t, "Bob", types.SEX_MALE)
	putInRoom(ch, vch)

	Act(types.AT_HIT, "You punch $N.", ch, vch, nil, nil, types.TO_CHAR)
	Act(types.AT_HITME, "$n punches you.", ch, vch, nil, nil, types.TO_VICT)

	gotCh := readOutput(t, ch, cc)
	if !strings.HasPrefix(gotCh, "&R") {
		t.Errorf("attacker (AT_HIT) should get &R prefix; got %q", gotCh)
	}
	gotVch := readOutput(t, vch, vc)
	if !strings.HasPrefix(gotVch, "&r") {
		t.Errorf("victim (AT_HITME) should get &r (dark red) prefix; got %q", gotVch)
	}
}

// TestAct_AT_PLAIN_NoPrefix verifies the zero-cost path: AT_PLAIN
// (mapped to empty-string in atColorCode) preserves the pre-migration
// uncolored output. This is what existing Go tests with AT_PLAIN rely on.
func TestAct_AT_PLAIN_NoPrefix(t *testing.T) {
	ch, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	vch, vc := makeChar(t, "Bob", types.SEX_MALE)
	putInRoom(ch, vch)

	Act(types.AT_PLAIN, "$n waves.", ch, vch, nil, nil, types.TO_VICT)

	got := readOutput(t, vch, vc)
	if strings.HasPrefix(got, "&") {
		t.Errorf("AT_PLAIN should NOT prepend any color code; got %q", got)
	}
	if strings.Contains(got, "&D") {
		t.Errorf("AT_PLAIN should NOT append &D reset; got %q", got)
	}
	if !strings.Contains(got, "Alice waves.") {
		t.Errorf("message body should be preserved; got %q", got)
	}
}

// TestAct_UnknownAType_NoPrefix_NoPanic verifies that passing an AT
// value not in atColorCode (e.g., a future AT_ constant whose mapping
// hasn't been populated yet) is safe: no prefix is added and no panic.
// Matches plan-tranche-c.md G2 acceptance criterion A8.
func TestAct_UnknownAType_NoPrefix_NoPanic(t *testing.T) {
	ch, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	vch, vc := makeChar(t, "Bob", types.SEX_MALE)
	putInRoom(ch, vch)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Act with unmapped AT value should not panic: %v", r)
		}
	}()
	// Use a high number far outside the AT_COLORBASE enum range.
	Act(9999, "$n waves.", ch, vch, nil, nil, types.TO_VICT)

	got := readOutput(t, vch, vc)
	if strings.HasPrefix(got, "&") {
		t.Errorf("unknown AT should produce no color prefix; got %q", got)
	}
	if !strings.Contains(got, "Alice waves.") {
		t.Errorf("message should still deliver; got %q", got)
	}
}

// TestAct_ColorPrefix_AT_IMMORT_GTELL verifies the channel-style codes
// ship a visible prefix (yellow / magenta respectively). Thin coverage
// but pins the table for the other commonly-used constants.
func TestAct_ColorPrefix_AT_IMMORT_GTELL(t *testing.T) {
	ch, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	vch, vc := makeChar(t, "Bob", types.SEX_MALE)
	putInRoom(ch, vch)

	Act(types.AT_IMMORT, "[Immortal] $n: hello", ch, vch, nil, nil, types.TO_VICT)
	if got := readOutput(t, vch, vc); !strings.HasPrefix(got, "&Y") {
		t.Errorf("AT_IMMORT should prepend &Y; got %q", got)
	}

	Act(types.AT_GTELL, "$n tells the group: 'hi'", ch, vch, nil, nil, types.TO_VICT)
	if got := readOutput(t, vch, vc); !strings.HasPrefix(got, "&P") {
		t.Errorf("AT_GTELL should prepend &P (magenta); got %q", got)
	}
}
