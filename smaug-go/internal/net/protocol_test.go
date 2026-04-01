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
