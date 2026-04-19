//go:build !windows

package main

import (
	"flag"
	"reflect"
	"testing"

	"github.com/eilidhmae/smaug/internal/boot"
)

// G5 argv-parser tests. The parser contract is defined in the executable
// plan (smaug-go/doc/plan-phase6-hotboot.md §G5):
//
//   (nil, nil) → flag absent, signals normal boot
//   (v,   nil) → flag present + valid
//   (nil, err) → flag present but malformed; main should log.Fatal
//
// Mutation gates pinned here (see §G5 Mutation gates):
//   - Gate 1: swap ln_fd / descriptor roles → TestParseHotbootArgs_OneListenerOneDesc fails.
//   - Gate 2: drop the ':' split in descriptor parse → TestParseHotbootArgs_Malformed_Rejected fails.

func TestParseHotbootArgs_Absent(t *testing.T) {
	args := []string{"-port", "4000", "-data", "../db"}
	got, err := ParseHotbootArgv(args)
	if err != nil {
		t.Fatalf("ParseHotbootArgv returned err=%v, want nil", err)
	}
	if got != nil {
		t.Fatalf("ParseHotbootArgv returned %+v, want nil (signals normal boot)", got)
	}
}

func TestParseHotbootArgs_OneListenerOneDesc(t *testing.T) {
	args := []string{"--hotboot-recover", "3", "4:0"}
	got, err := ParseHotbootArgv(args)
	if err != nil {
		t.Fatalf("ParseHotbootArgv returned err=%v, want nil", err)
	}
	if got == nil {
		t.Fatal("ParseHotbootArgv returned nil, want non-nil")
	}
	if got.LnFD != 3 {
		t.Errorf("LnFD = %d, want 3", got.LnFD)
	}
	wantSessions := []boot.FDSession{{FD: 4, SessionIdx: 0}}
	if !reflect.DeepEqual(got.Sessions, wantSessions) {
		t.Errorf("Sessions = %+v, want %+v", got.Sessions, wantSessions)
	}
}

func TestParseHotbootArgs_OneListenerThreeDescs(t *testing.T) {
	// Paired twin for Gate 1: if the parser swaps ln/desc roles this
	// multi-session test fails even more loudly.
	args := []string{"--hotboot-recover", "3", "4:0", "5:1", "6:2"}
	got, err := ParseHotbootArgv(args)
	if err != nil {
		t.Fatalf("ParseHotbootArgv err = %v", err)
	}
	if got == nil {
		t.Fatal("parsed nil, want value")
	}
	if got.LnFD != 3 {
		t.Errorf("LnFD = %d, want 3", got.LnFD)
	}
	want := []boot.FDSession{
		{FD: 4, SessionIdx: 0},
		{FD: 5, SessionIdx: 1},
		{FD: 6, SessionIdx: 2},
	}
	if !reflect.DeepEqual(got.Sessions, want) {
		t.Errorf("Sessions = %+v, want %+v", got.Sessions, want)
	}
}

func TestParseHotbootArgs_Malformed_Rejected(t *testing.T) {
	args := []string{"--hotboot-recover", "3", "4"} // descriptor missing :idx
	got, err := ParseHotbootArgv(args)
	if err == nil {
		t.Fatalf("ParseHotbootArgv err = nil, want non-nil (descriptor %q missing ':idx')", "4")
	}
	if got != nil {
		t.Errorf("ParseHotbootArgv returned %+v on malformed args, want nil", got)
	}
}

func TestParseHotbootArgs_MalformedDescriptor_NonNumericFD(t *testing.T) {
	// Twin for Gate 2: confirms malformed-descriptor branch fires on non-numeric too.
	args := []string{"--hotboot-recover", "3", "abc:0"}
	_, err := ParseHotbootArgv(args)
	if err == nil {
		t.Fatal("expected error on non-numeric descriptor FD")
	}
}

func TestParseHotbootArgs_MalformedDescriptor_NonNumericIdx(t *testing.T) {
	args := []string{"--hotboot-recover", "3", "4:xyz"}
	_, err := ParseHotbootArgv(args)
	if err == nil {
		t.Fatal("expected error on non-numeric descriptor idx")
	}
}

func TestParseHotbootArgs_ZeroFDRejected(t *testing.T) {
	// stdin reserved — listener FD 0 is malformed.
	args := []string{"--hotboot-recover", "0", "4:0"}
	got, err := ParseHotbootArgv(args)
	if err == nil {
		t.Fatal("expected error on ln_fd=0 (stdin reserved)")
	}
	if got != nil {
		t.Errorf("ParseHotbootArgv returned %+v on ln_fd=0, want nil", got)
	}
}

func TestParseHotbootArgs_MissingArgs(t *testing.T) {
	// --hotboot-recover with no payload at all.
	if _, err := ParseHotbootArgv([]string{"--hotboot-recover"}); err == nil {
		t.Error("expected error on empty payload")
	}
	// Just ln_fd, no descriptor tokens.
	if _, err := ParseHotbootArgv([]string{"--hotboot-recover", "3"}); err == nil {
		t.Error("expected error on missing descriptor")
	}
}

func TestParseHotbootArgs_OrderIndependent(t *testing.T) {
	// --hotboot-recover appears BEFORE -port and -data. All three must parse.
	args := []string{"--hotboot-recover", "3", "4:0", "-port", "4000", "-data", "../db"}

	got, err := ParseHotbootArgv(args)
	if err != nil {
		t.Fatalf("ParseHotbootArgv err = %v", err)
	}
	if got == nil {
		t.Fatal("parsed nil")
	}
	if got.LnFD != 3 {
		t.Errorf("LnFD = %d, want 3", got.LnFD)
	}
	if len(got.Sessions) != 1 || got.Sessions[0].FD != 4 || got.Sessions[0].SessionIdx != 0 {
		t.Errorf("Sessions = %+v, want [{FD:4 SessionIdx:0}]", got.Sessions)
	}

	// After stripping, `flag` should happily consume -port and -data.
	stripped := StripHotbootArgv(args)
	wantStripped := []string{"-port", "4000", "-data", "../db"}
	if !reflect.DeepEqual(stripped, wantStripped) {
		t.Fatalf("StripHotbootArgv = %v, want %v", stripped, wantStripped)
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	port := fs.Int("port", 0, "")
	data := fs.String("data", "", "")
	if err := fs.Parse(stripped); err != nil {
		t.Fatalf("flag.Parse on stripped: %v", err)
	}
	if *port != 4000 {
		t.Errorf("-port = %d, want 4000", *port)
	}
	if *data != "../db" {
		t.Errorf("-data = %q, want %q", *data, "../db")
	}
}

func TestParseHotbootArgs_OrderIndependent_FlagAfterOtherFlags(t *testing.T) {
	// Mirror: --hotboot-recover sits AFTER -port/-data. Must still parse,
	// and strip must preserve the leading flags.
	args := []string{"-port", "4000", "-data", "../db", "--hotboot-recover", "3", "4:0"}

	got, err := ParseHotbootArgv(args)
	if err != nil {
		t.Fatalf("ParseHotbootArgv err = %v", err)
	}
	if got == nil || got.LnFD != 3 || len(got.Sessions) != 1 {
		t.Fatalf("parsed %+v, want {LnFD:3, Sessions:[{4,0}]}", got)
	}

	stripped := StripHotbootArgv(args)
	wantStripped := []string{"-port", "4000", "-data", "../db"}
	if !reflect.DeepEqual(stripped, wantStripped) {
		t.Errorf("StripHotbootArgv = %v, want %v", stripped, wantStripped)
	}
}

func TestStripHotbootArgv_NoOpWhenAbsent(t *testing.T) {
	args := []string{"-port", "4000", "-data", "../db"}
	stripped := StripHotbootArgv(args)
	if !reflect.DeepEqual(stripped, args) {
		t.Errorf("StripHotbootArgv modified args when flag absent: got %v", stripped)
	}
}
