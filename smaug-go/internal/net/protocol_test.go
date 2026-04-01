package net

import (
	"bytes"
	"testing"
)

func TestTelnetNeg(t *testing.T) {
	seq := TelnetNeg(WILL, TELOPT_COMPRESS2)
	if len(seq) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(seq))
	}
	if seq[0] != IAC || seq[1] != WILL || seq[2] != TELOPT_COMPRESS2 {
		t.Errorf("unexpected sequence: %v", seq)
	}
}

func TestTelnetSubneg(t *testing.T) {
	data := []byte{1, 2, 3}
	seq := TelnetSubneg(TELOPT_MSDP, data)

	// Should be: IAC SB MSDP 1 2 3 IAC SE
	if len(seq) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(seq))
	}
	if seq[0] != IAC || seq[1] != SB || seq[2] != TELOPT_MSDP {
		t.Error("bad subneg header")
	}
	if seq[len(seq)-2] != IAC || seq[len(seq)-1] != SE {
		t.Error("bad subneg footer")
	}
}

func TestBuildMSSPPayload(t *testing.T) {
	info := &MSSPInfo{
		Name:    "TestMUD",
		Players: 5,
		Uptime:  1000,
		Port:    4000,
	}
	payload := BuildMSSPPayload(info)

	if len(payload) == 0 {
		t.Fatal("MSSP payload should not be empty")
	}

	// Should start with IAC SB MSSP
	if payload[0] != IAC || payload[1] != SB || payload[2] != TELOPT_MSSP {
		t.Error("MSSP payload should start with IAC SB MSSP")
	}

	// Should end with IAC SE
	if payload[len(payload)-2] != IAC || payload[len(payload)-1] != SE {
		t.Error("MSSP payload should end with IAC SE")
	}

	// Should contain NAME
	if !bytes.Contains(payload, []byte("NAME")) {
		t.Error("MSSP payload should contain NAME variable")
	}
	if !bytes.Contains(payload, []byte("TestMUD")) {
		t.Error("MSSP payload should contain server name")
	}
}

func TestBuildMSDPReport(t *testing.T) {
	report := BuildMSDPReport([]MSDPVariable{
		{"HEALTH", "100"},
		{"MANA", "50"},
	})

	if len(report) == 0 {
		t.Fatal("MSDP report should not be empty")
	}

	// Should contain variable names and values
	if !bytes.Contains(report, []byte("HEALTH")) {
		t.Error("MSDP report should contain HEALTH")
	}
	if !bytes.Contains(report, []byte("100")) {
		t.Error("MSDP report should contain value 100")
	}
}

func TestMSDPCharReport(t *testing.T) {
	report := MSDPCharReport("Gandalf", 100, 200, 50, 100, 80, 120)

	if !bytes.Contains(report, []byte("Gandalf")) {
		t.Error("should contain character name")
	}
	if !bytes.Contains(report, []byte("CHARACTER_NAME")) {
		t.Error("should contain CHARACTER_NAME variable")
	}
}

func TestCompressDecompress(t *testing.T) {
	original := []byte("Hello, World! This is a test of MCCP2 compression.")

	compressed, err := CompressData(original)
	if err != nil {
		t.Fatalf("CompressData: %v", err)
	}

	decompressed, err := DecompressData(compressed)
	if err != nil {
		t.Fatalf("DecompressData: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("round-trip failed: got %q, want %q", decompressed, original)
	}
}

func TestCompressData_Smaller(t *testing.T) {
	// Repetitive data should compress well
	data := bytes.Repeat([]byte("AAAA"), 1000)

	compressed, err := CompressData(data)
	if err != nil {
		t.Fatalf("CompressData: %v", err)
	}

	if len(compressed) >= len(data) {
		t.Errorf("compressed size %d should be smaller than original %d", len(compressed), len(data))
	}
}

func TestMCCPStartSequence(t *testing.T) {
	seq := MCCPStartSequence()
	// IAC SB COMPRESS2 IAC SE
	if len(seq) != 5 {
		t.Fatalf("expected 5 bytes, got %d", len(seq))
	}
	if seq[0] != IAC || seq[1] != SB || seq[2] != TELOPT_COMPRESS2 {
		t.Error("bad MCCP start header")
	}
	if seq[3] != IAC || seq[4] != SE {
		t.Error("bad MCCP start footer")
	}
}

// --------------- stripTelnetIAC ---------------

func TestStripTelnetIAC_NoIAC(t *testing.T) {
	input := []byte("hello world")
	got := stripTelnetIAC(input)
	if !bytes.Equal(got, input) {
		t.Errorf("stripTelnetIAC(%q) = %q, want %q", input, got, input)
	}
}

func TestStripTelnetIAC_EmptyInput(t *testing.T) {
	got := stripTelnetIAC([]byte{})
	if len(got) != 0 {
		t.Errorf("stripTelnetIAC(empty) = %v, want empty", got)
	}
}

func TestStripTelnetIAC_WillDoWontDont(t *testing.T) {
	tests := []struct {
		name string
		cmd  byte
	}{
		{"WILL", WILL},
		{"WONT", WONT},
		{"DO", DO},
		{"DONT", DONT},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// IAC <cmd> <option> should be stripped (3 bytes)
			input := []byte{'A', IAC, tt.cmd, TELOPT_ECHO, 'B'}
			got := stripTelnetIAC(input)
			want := []byte{'A', 'B'}
			if !bytes.Equal(got, want) {
				t.Errorf("stripTelnetIAC with %s: got %v, want %v", tt.name, got, want)
			}
		})
	}
}

func TestStripTelnetIAC_TwoByteCommands(t *testing.T) {
	// NOP (241), GA (249), and other 2-byte commands
	cmds := []byte{241, GA, 242, 243, 244, 245, 246}
	for _, cmd := range cmds {
		input := []byte{'X', IAC, cmd, 'Y'}
		got := stripTelnetIAC(input)
		want := []byte{'X', 'Y'}
		if !bytes.Equal(got, want) {
			t.Errorf("stripTelnetIAC with cmd %d: got %v, want %v", cmd, got, want)
		}
	}
}

func TestStripTelnetIAC_EscapedFF(t *testing.T) {
	// IAC IAC -> one literal 0xFF
	input := []byte{'A', IAC, IAC, 'B'}
	got := stripTelnetIAC(input)
	want := []byte{'A', IAC, 'B'}
	if !bytes.Equal(got, want) {
		t.Errorf("escaped 0xFF: got %v, want %v", got, want)
	}
}

func TestStripTelnetIAC_MultipleSequences(t *testing.T) {
	// Mix of WILL, escaped FF, and two-byte command
	input := []byte{
		'H', 'i',
		IAC, WILL, TELOPT_ECHO,       // 3-byte, stripped
		IAC, IAC,                       // escaped, keep one 0xFF
		IAC, GA,                        // 2-byte, stripped
		'!',
	}
	got := stripTelnetIAC(input)
	want := []byte{'H', 'i', IAC, '!'}
	if !bytes.Equal(got, want) {
		t.Errorf("multiple sequences: got %v, want %v", got, want)
	}
}

func TestStripTelnetIAC_TruncatedWILL(t *testing.T) {
	// IAC WILL at end of data (missing option byte) — should skip past end safely
	input := []byte{'A', IAC, WILL}
	got := stripTelnetIAC(input)
	// i += 3 goes past end, so 'A' is the only output
	want := []byte{'A'}
	if !bytes.Equal(got, want) {
		t.Errorf("truncated WILL: got %v, want %v", got, want)
	}
}

func TestStripTelnetIAC_TrailingIAC(t *testing.T) {
	// Lone IAC at end (no following byte) — kept as literal
	input := []byte{'X', IAC}
	got := stripTelnetIAC(input)
	// The condition data[i] == iac && i+1 < len(data) fails, so IAC is kept
	want := []byte{'X', IAC}
	if !bytes.Equal(got, want) {
		t.Errorf("trailing IAC: got %v, want %v", got, want)
	}
}

func TestStripTelnetIAC_OnlyIACSequences(t *testing.T) {
	// All IAC commands, no real data
	input := []byte{IAC, WILL, TELOPT_ECHO, IAC, DO, TELOPT_NAWS, IAC, GA}
	got := stripTelnetIAC(input)
	if len(got) != 0 {
		t.Errorf("all-IAC input: got %v, want empty", got)
	}
}

func TestStripTelnetIAC_MultipleEscapedFF(t *testing.T) {
	// IAC IAC IAC IAC -> two literal 0xFF bytes
	input := []byte{IAC, IAC, IAC, IAC}
	got := stripTelnetIAC(input)
	want := []byte{IAC, IAC}
	if !bytes.Equal(got, want) {
		t.Errorf("double escaped FF: got %v, want %v", got, want)
	}
}

// --------------- stripCR ---------------

func TestStripCR_NoCR(t *testing.T) {
	input := []byte("hello world")
	got := stripCR(input)
	if !bytes.Equal(got, input) {
		t.Errorf("stripCR(%q) = %q, want %q", input, got, input)
	}
}

func TestStripCR_Empty(t *testing.T) {
	got := stripCR([]byte{})
	if len(got) != 0 {
		t.Errorf("stripCR(empty) = %v, want empty", got)
	}
}

func TestStripCR_AllCR(t *testing.T) {
	input := []byte{'\r', '\r', '\r'}
	got := stripCR(input)
	if len(got) != 0 {
		t.Errorf("stripCR(all \\r) = %v, want empty", got)
	}
}

func TestStripCR_Mixed(t *testing.T) {
	input := []byte("hello\r\nworld\r")
	got := stripCR(input)
	want := []byte("hello\nworld")
	if !bytes.Equal(got, want) {
		t.Errorf("stripCR mixed: got %q, want %q", got, want)
	}
}

func TestStripCR_OnlyCRLF(t *testing.T) {
	input := []byte("\r\n\r\n")
	got := stripCR(input)
	want := []byte("\n\n")
	if !bytes.Equal(got, want) {
		t.Errorf("stripCR CRLF: got %v, want %v", got, want)
	}
}

// --------------- TelnetNeg additional ---------------

func TestTelnetNeg_AllVerbs(t *testing.T) {
	verbs := []struct {
		name string
		verb byte
	}{
		{"WILL", WILL},
		{"WONT", WONT},
		{"DO", DO},
		{"DONT", DONT},
	}
	for _, tt := range verbs {
		t.Run(tt.name, func(t *testing.T) {
			seq := TelnetNeg(tt.verb, TELOPT_ECHO)
			if len(seq) != 3 {
				t.Fatalf("len=%d, want 3", len(seq))
			}
			if seq[0] != IAC {
				t.Errorf("seq[0]=%d, want IAC(%d)", seq[0], IAC)
			}
			if seq[1] != tt.verb {
				t.Errorf("seq[1]=%d, want %d", seq[1], tt.verb)
			}
			if seq[2] != TELOPT_ECHO {
				t.Errorf("seq[2]=%d, want TELOPT_ECHO(%d)", seq[2], TELOPT_ECHO)
			}
		})
	}
}

func TestTelnetNeg_AllOptions(t *testing.T) {
	options := []struct {
		name   string
		option byte
	}{
		{"ECHO", TELOPT_ECHO},
		{"TTYPE", TELOPT_TTYPE},
		{"NAWS", TELOPT_NAWS},
		{"COMPRESS2", TELOPT_COMPRESS2},
		{"MSDP", TELOPT_MSDP},
		{"MSSP", TELOPT_MSSP},
	}
	for _, tt := range options {
		t.Run(tt.name, func(t *testing.T) {
			seq := TelnetNeg(DO, tt.option)
			if seq[2] != tt.option {
				t.Errorf("option byte = %d, want %d", seq[2], tt.option)
			}
		})
	}
}

// --------------- TelnetSubneg additional ---------------

func TestTelnetSubneg_EmptyData(t *testing.T) {
	seq := TelnetSubneg(TELOPT_MSSP, nil)
	// IAC SB MSSP IAC SE
	if len(seq) != 5 {
		t.Fatalf("len=%d, want 5", len(seq))
	}
	if seq[0] != IAC || seq[1] != SB || seq[2] != TELOPT_MSSP {
		t.Error("bad header")
	}
	if seq[3] != IAC || seq[4] != SE {
		t.Error("bad footer")
	}
}

func TestTelnetSubneg_LargePayload(t *testing.T) {
	data := bytes.Repeat([]byte{0x42}, 256)
	seq := TelnetSubneg(TELOPT_MSDP, data)
	// IAC SB MSDP <256 bytes> IAC SE = 261
	if len(seq) != 261 {
		t.Fatalf("len=%d, want 261", len(seq))
	}
	// Verify payload is intact
	payload := seq[3 : len(seq)-2]
	if !bytes.Equal(payload, data) {
		t.Error("payload corrupted")
	}
}

// --------------- CompressData / DecompressData edge cases ---------------

func TestCompressData_Empty(t *testing.T) {
	compressed, err := CompressData([]byte{})
	if err != nil {
		t.Fatalf("CompressData(empty): %v", err)
	}
	decompressed, err := DecompressData(compressed)
	if err != nil {
		t.Fatalf("DecompressData: %v", err)
	}
	if len(decompressed) != 0 {
		t.Errorf("expected empty, got %d bytes", len(decompressed))
	}
}

func TestDecompressData_InvalidInput(t *testing.T) {
	_, err := DecompressData([]byte{0x01, 0x02, 0x03})
	if err == nil {
		t.Error("DecompressData with invalid input should error")
	}
}

func TestCompressDecompress_LargeData(t *testing.T) {
	// 100KB of mixed data
	original := make([]byte, 100*1024)
	for i := range original {
		original[i] = byte(i % 251)
	}
	compressed, err := CompressData(original)
	if err != nil {
		t.Fatalf("CompressData: %v", err)
	}
	decompressed, err := DecompressData(compressed)
	if err != nil {
		t.Fatalf("DecompressData: %v", err)
	}
	if !bytes.Equal(original, decompressed) {
		t.Error("round-trip failed for large data")
	}
}

func TestCompressDecompress_BinaryData(t *testing.T) {
	// Data with all byte values 0-255
	original := make([]byte, 256)
	for i := range original {
		original[i] = byte(i)
	}
	compressed, err := CompressData(original)
	if err != nil {
		t.Fatalf("CompressData: %v", err)
	}
	decompressed, err := DecompressData(compressed)
	if err != nil {
		t.Fatalf("DecompressData: %v", err)
	}
	if !bytes.Equal(original, decompressed) {
		t.Error("round-trip failed for binary data")
	}
}

// --------------- BuildMSDPReport additional ---------------

func TestBuildMSDPReport_EmptyVars(t *testing.T) {
	report := BuildMSDPReport(nil)
	// Should be: IAC SB MSDP IAC SE (no data)
	if len(report) != 5 {
		t.Fatalf("len=%d, want 5", len(report))
	}
	if report[0] != IAC || report[1] != SB || report[2] != TELOPT_MSDP {
		t.Error("bad header")
	}
	if report[3] != IAC || report[4] != SE {
		t.Error("bad footer")
	}
}

func TestBuildMSDPReport_SingleVar(t *testing.T) {
	report := BuildMSDPReport([]MSDPVariable{
		{"HP", "42"},
	})
	// Verify structure: IAC SB MSDP <VAR> HP <VAL> 42 IAC SE
	if report[0] != IAC || report[1] != SB || report[2] != TELOPT_MSDP {
		t.Error("bad header")
	}
	inner := report[3 : len(report)-2]
	if inner[0] != MSDP_VAR {
		t.Errorf("expected MSDP_VAR, got %d", inner[0])
	}
	if !bytes.Contains(inner, []byte("HP")) {
		t.Error("missing variable name HP")
	}
	// Find MSDP_VAL
	valIdx := bytes.IndexByte(inner, MSDP_VAL)
	if valIdx < 0 {
		t.Fatal("missing MSDP_VAL")
	}
	if !bytes.Contains(inner[valIdx:], []byte("42")) {
		t.Error("missing value 42")
	}
}

func TestBuildMSDPReport_MultipleVars(t *testing.T) {
	report := BuildMSDPReport([]MSDPVariable{
		{"A", "1"},
		{"B", "2"},
		{"C", "3"},
	})
	inner := report[3 : len(report)-2]
	// Count MSDP_VAR bytes
	varCount := 0
	for _, b := range inner {
		if b == MSDP_VAR {
			varCount++
		}
	}
	if varCount != 3 {
		t.Errorf("expected 3 MSDP_VAR markers, got %d", varCount)
	}
}

// --------------- MSDPCharReport additional ---------------

func TestMSDPCharReport_VerifyAllFields(t *testing.T) {
	report := MSDPCharReport("TestPC", 75, 100, 30, 50, 60, 80)
	inner := report[3 : len(report)-2]

	expects := []string{
		"CHARACTER_NAME", "TestPC",
		"HEALTH", "75",
		"HEALTH_MAX", "100",
		"MANA", "30",
		"MANA_MAX", "50",
		"MOVEMENT", "60",
		"MOVEMENT_MAX", "80",
	}
	for _, s := range expects {
		if !bytes.Contains(inner, []byte(s)) {
			t.Errorf("MSDPCharReport missing %q", s)
		}
	}
}

// --------------- BuildMSSPPayload additional ---------------

func TestBuildMSSPPayload_AllFields(t *testing.T) {
	info := &MSSPInfo{
		Name:    "DragonMUD",
		Players: 42,
		Uptime:  86400,
		Port:    9999,
	}
	payload := BuildMSSPPayload(info)

	// Verify all expected fields are in the payload
	expects := []string{
		"NAME", "DragonMUD",
		"PLAYERS", "42",
		"UPTIME", "86400",
		"CODEBASE", "SMAUG 1.8 (Go)",
		"FAMILY", "DikuMUD",
		"PORT", "9999",
		"HOSTNAME", "localhost",
		"ANSI", "1",
		"MCCP", "1",
	}
	for _, s := range expects {
		if !bytes.Contains(payload, []byte(s)) {
			t.Errorf("MSSP payload missing %q", s)
		}
	}
}

func TestBuildMSSPPayload_ZeroValues(t *testing.T) {
	info := &MSSPInfo{
		Name: "ZeroMUD",
	}
	payload := BuildMSSPPayload(info)
	if !bytes.Contains(payload, []byte("ZeroMUD")) {
		t.Error("should contain server name")
	}
	if !bytes.Contains(payload, []byte("PLAYERS")) {
		t.Error("should contain PLAYERS field even with 0")
	}
}

func TestBuildMSSPPayload_VarValStructure(t *testing.T) {
	info := &MSSPInfo{Name: "Test", Players: 1, Uptime: 1, Port: 1}
	payload := BuildMSSPPayload(info)
	inner := payload[3 : len(payload)-2]

	// Each variable should be preceded by MSSP_VAR, each value by MSSP_VAL
	varCount := 0
	valCount := 0
	for _, b := range inner {
		if b == MSSP_VAR {
			varCount++
		}
		if b == MSSP_VAL {
			valCount++
		}
	}
	// 8 variables: NAME, PLAYERS, UPTIME, CODEBASE, FAMILY, PORT, HOSTNAME, ANSI, MCCP = 9
	if varCount != 9 {
		t.Errorf("MSSP_VAR count = %d, want 9", varCount)
	}
	if valCount != 9 {
		t.Errorf("MSSP_VAL count = %d, want 9", valCount)
	}
}
