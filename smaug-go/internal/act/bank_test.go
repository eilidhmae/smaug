package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func setupBankRoom(ch *types.CharData) *types.CharData {
	room := &types.RoomIndexData{Vnum: 9100, Name: "Bank"}
	handler.CharToRoom(ch, room)
	banker := &types.CharData{
		Name:       "banker",
		ShortDescr: "the banker",
	}
	banker.Act.Set(types.ACT_IS_NPC)
	banker.Act.Set(types.ACT_BANKER)
	handler.CharToRoom(banker, room)
	return banker
}

func TestDoBank(t *testing.T) {
	tests := []struct {
		name       string
		argument   string
		gold       int
		gbalance   int
		noBanker   bool
		isNPC      bool
		wantGold   int
		wantBal    int
		wantOutput string
	}{
		{
			name:       "no argument shows usage",
			argument:   "",
			gold:       500,
			gbalance:   0,
			wantGold:   500,
			wantBal:    0,
			wantOutput: "Syntax:",
		},
		{
			name:       "no banker in room",
			argument:   "balance",
			gold:       500,
			gbalance:   0,
			noBanker:   true,
			wantGold:   500,
			wantBal:    0,
			wantOutput: "can't do that here",
		},
		{
			name:     "NPC tries to use bank",
			argument: "balance",
			gold:     500,
			gbalance: 0,
			isNPC:    true,
			wantGold: 500,
			wantBal:  0,
		},
		{
			name:       "balance shows GBalance",
			argument:   "balance",
			gold:       500,
			gbalance:   1234,
			wantGold:   500,
			wantBal:    1234,
			wantOutput: "1234",
		},
		{
			name:       "deposit 100",
			argument:   "deposit 100",
			gold:       500,
			gbalance:   0,
			wantGold:   400,
			wantBal:    100,
			wantOutput: "deposit",
		},
		{
			name:       "deposit all",
			argument:   "deposit all",
			gold:       500,
			gbalance:   50,
			wantGold:   0,
			wantBal:    550,
			wantOutput: "deposit",
		},
		{
			name:       "deposit 0 is error",
			argument:   "deposit 0",
			gold:       500,
			gbalance:   0,
			wantGold:   500,
			wantBal:    0,
			wantOutput: "much",
		},
		{
			name:       "deposit negative is error",
			argument:   "deposit -5",
			gold:       500,
			gbalance:   0,
			wantGold:   500,
			wantBal:    0,
			wantOutput: "much",
		},
		{
			name:       "deposit insufficient gold",
			argument:   "deposit 999",
			gold:       500,
			gbalance:   0,
			wantGold:   500,
			wantBal:    0,
			wantOutput: "don't have that much",
		},
		{
			name:       "withdraw 50",
			argument:   "withdraw 50",
			gold:       100,
			gbalance:   200,
			wantGold:   150,
			wantBal:    150,
			wantOutput: "withdraw",
		},
		{
			name:       "withdraw all",
			argument:   "withdraw all",
			gold:       100,
			gbalance:   300,
			wantGold:   400,
			wantBal:    0,
			wantOutput: "withdraw",
		},
		{
			name:       "withdraw insufficient balance",
			argument:   "withdraw 999",
			gold:       100,
			gbalance:   200,
			wantGold:   100,
			wantBal:    200,
			wantOutput: "don't have that much",
		},
		{
			name:       "withdraw 0 is error",
			argument:   "withdraw 0",
			gold:       100,
			gbalance:   200,
			wantGold:   100,
			wantBal:    200,
			wantOutput: "much",
		},
		{
			name:       "unknown subcommand shows usage",
			argument:   "foo",
			gold:       500,
			gbalance:   0,
			wantGold:   500,
			wantBal:    0,
			wantOutput: "Syntax:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, client := makeTestChar("Tester")
			defer client.Close()
			defer ch.Desc.Conn.Close()

			ch.Gold = tt.gold
			ch.PCData.GBalance = tt.gbalance

			if tt.isNPC {
				ch.Act.Set(types.ACT_IS_NPC)
				DoBank(ch, tt.argument)
				// NPC should produce no output and no state change
				if ch.Gold != tt.wantGold {
					t.Errorf("gold = %d, want %d", ch.Gold, tt.wantGold)
				}
				return
			}

			if tt.noBanker {
				room := &types.RoomIndexData{Vnum: 9100, Name: "Empty Room"}
				handler.CharToRoom(ch, room)
			} else {
				setupBankRoom(ch)
			}

			DoBank(ch, tt.argument)

			out := readOutput(ch, client)

			if tt.wantOutput != "" && !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantOutput)) {
				t.Errorf("output = %q, want substring %q", out, tt.wantOutput)
			}

			if ch.Gold != tt.wantGold {
				t.Errorf("gold = %d, want %d", ch.Gold, tt.wantGold)
			}

			if ch.PCData.GBalance != tt.wantBal {
				t.Errorf("gbalance = %d, want %d", ch.PCData.GBalance, tt.wantBal)
			}
		})
	}
}
