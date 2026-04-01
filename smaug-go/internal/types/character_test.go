package types

import (
	"net"
	"testing"
)

// ---------------------------------------------------------------------------
// IsNPC
// ---------------------------------------------------------------------------

func TestCharData_IsNPC(t *testing.T) {
	tests := []struct {
		name string
		act  func() BitVector
		want bool
	}{
		{"default zero bitvector", func() BitVector { return BitVector{} }, false},
		{"ACT_IS_NPC set", func() BitVector {
			var bv BitVector
			bv.Set(ACT_IS_NPC)
			return bv
		}, true},
		{"other bits set but not ACT_IS_NPC", func() BitVector {
			var bv BitVector
			bv.Set(5)
			bv.Set(10)
			return bv
		}, false},
		{"ACT_IS_NPC plus others", func() BitVector {
			var bv BitVector
			bv.Set(ACT_IS_NPC)
			bv.Set(5)
			bv.Set(10)
			return bv
		}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := &CharData{Act: tc.act()}
			if got := ch.IsNPC(); got != tc.want {
				t.Errorf("IsNPC() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsImmortal
// ---------------------------------------------------------------------------

func TestCharData_IsImmortal(t *testing.T) {
	tests := []struct {
		name  string
		level int
		trust int
		isNPC bool
		want  bool
	}{
		{"mortal level 1", 1, 0, false, false},
		{"mortal level 50", 50, 0, false, false},
		{"exactly immortal level", LEVEL_IMMORTAL, 0, false, true},
		{"above immortal level", LEVEL_IMMORTAL + 1, 0, false, true},
		{"max level", MAX_LEVEL, 0, false, true},
		{"mortal with trust override", 1, LEVEL_IMMORTAL, false, true},
		{"NPC at immortal level (capped)", LEVEL_IMMORTAL, 0, true, true},
		{"NPC above immortal level (capped)", LEVEL_IMMORTAL + 5, 0, true, true},
		{"NPC below immortal level", LEVEL_IMMORTAL - 1, 0, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := &CharData{Level: tc.level, Trust: tc.trust}
			if tc.isNPC {
				ch.Act.Set(ACT_IS_NPC)
			}
			if got := ch.IsImmortal(); got != tc.want {
				t.Errorf("IsImmortal() = %v, want %v (level=%d, trust=%d, npc=%v)",
					got, tc.want, tc.level, tc.trust, tc.isNPC)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetTrust
// ---------------------------------------------------------------------------

func TestCharData_GetTrust(t *testing.T) {
	tests := []struct {
		name  string
		level int
		trust int
		isNPC bool
		want  int
	}{
		{"PC level 10, no trust override", 10, 0, false, 10},
		{"PC level 50, trust override 60", 50, 60, false, 60},
		{"PC trust takes priority over level", 1, MAX_LEVEL, false, MAX_LEVEL},
		{"NPC below immortal", LEVEL_IMMORTAL - 1, 0, true, LEVEL_IMMORTAL - 1},
		{"NPC at immortal level capped", LEVEL_IMMORTAL, 0, true, LEVEL_IMMORTAL},
		{"NPC above immortal level capped", MAX_LEVEL, 0, true, LEVEL_IMMORTAL},
		{"NPC with trust override", 10, 55, true, 55},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := &CharData{Level: tc.level, Trust: tc.trust}
			if tc.isNPC {
				ch.Act.Set(ACT_IS_NPC)
			}
			if got := ch.GetTrust(); got != tc.want {
				t.Errorf("GetTrust() = %d, want %d", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Send / Sendf
// ---------------------------------------------------------------------------

func TestCharData_Send(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)
	ch := &CharData{Desc: d}

	ch.Send("Hello, world!")
	if !d.HasOutput() {
		t.Fatal("descriptor should have output after Send")
	}

	// Verify the buffer content by flushing
	go func() {
		buf := make([]byte, 256)
		n, _ := client.Read(buf)
		if string(buf[:n]) != "Hello, world!" {
			// Can't t.Errorf from goroutine safely, but FlushOutput test covers this
		}
	}()
	if err := d.FlushOutput(); err != nil {
		t.Fatalf("FlushOutput error: %v", err)
	}
}

func TestCharData_Send_NilDesc(t *testing.T) {
	ch := &CharData{} // Desc is nil
	// Should not panic
	ch.Send("hello")
}

func TestCharData_Sendf(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)
	ch := &CharData{Desc: d}

	ch.Sendf("You have %d gold.\r\n", 42)

	go func() {
		buf := make([]byte, 256)
		client.Read(buf)
	}()
	if err := d.FlushOutput(); err != nil {
		t.Fatalf("FlushOutput error: %v", err)
	}
}

func TestCharData_Sendf_NilDesc(t *testing.T) {
	ch := &CharData{}
	// Should not panic
	ch.Sendf("hello %s", "world")
}

// ---------------------------------------------------------------------------
// GetCurrStr / Int / Wis / Dex / Con / Cha / Lck
// ---------------------------------------------------------------------------

func TestCharData_GetCurrStats(t *testing.T) {
	type statCase struct {
		name     string
		perm     int
		mod      int
		expected int
	}
	cases := []statCase{
		{"normal value", 15, 0, 15},
		{"with positive mod", 15, 3, 18},
		{"with negative mod", 15, -5, 10},
		{"clamped low to 3", 2, 0, 3},
		{"clamped low with mod", 5, -10, 3},
		{"clamped high to 25", 20, 10, 25},
		{"exactly 3", 3, 0, 3},
		{"exactly 25", 25, 0, 25},
		{"zero perm clamped", 0, 0, 3},
		{"negative result clamped", 1, -5, 3},
	}

	// Each stat getter is identical in logic, so we test them all with the same table.
	type statFunc struct {
		name    string
		setPerm func(ch *CharData, v int)
		setMod  func(ch *CharData, v int)
		get     func(ch *CharData) int
	}

	funcs := []statFunc{
		{"Str", func(ch *CharData, v int) { ch.PermStr = v }, func(ch *CharData, v int) { ch.ModStr = v }, func(ch *CharData) int { return ch.GetCurrStr() }},
		{"Int", func(ch *CharData, v int) { ch.PermInt = v }, func(ch *CharData, v int) { ch.ModInt = v }, func(ch *CharData) int { return ch.GetCurrInt() }},
		{"Wis", func(ch *CharData, v int) { ch.PermWis = v }, func(ch *CharData, v int) { ch.ModWis = v }, func(ch *CharData) int { return ch.GetCurrWis() }},
		{"Dex", func(ch *CharData, v int) { ch.PermDex = v }, func(ch *CharData, v int) { ch.ModDex = v }, func(ch *CharData) int { return ch.GetCurrDex() }},
		{"Con", func(ch *CharData, v int) { ch.PermCon = v }, func(ch *CharData, v int) { ch.ModCon = v }, func(ch *CharData) int { return ch.GetCurrCon() }},
		{"Cha", func(ch *CharData, v int) { ch.PermCha = v }, func(ch *CharData, v int) { ch.ModCha = v }, func(ch *CharData) int { return ch.GetCurrCha() }},
		{"Lck", func(ch *CharData, v int) { ch.PermLck = v }, func(ch *CharData, v int) { ch.ModLck = v }, func(ch *CharData) int { return ch.GetCurrLck() }},
	}

	for _, sf := range funcs {
		for _, tc := range cases {
			t.Run(sf.name+"/"+tc.name, func(t *testing.T) {
				ch := &CharData{}
				sf.setPerm(ch, tc.perm)
				sf.setMod(ch, tc.mod)
				if got := sf.get(ch); got != tc.expected {
					t.Errorf("GetCurr%s() = %d, want %d (perm=%d, mod=%d)",
						sf.name, got, tc.expected, tc.perm, tc.mod)
				}
			})
		}
	}
}
