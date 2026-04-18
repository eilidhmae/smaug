package util

import (
	"fmt"
	"log"
)

// BugSink, when non-nil, captures BUG-prefixed messages in addition to
// writing them through the stdlib logger. Intended as a test seam only;
// production leaves it nil so messages flow through `log.Printf` as
// before. Set via SetBugSink and restore with a `t.Cleanup` to avoid
// leaking state between tests.
var BugSink func(msg string)

// SetBugSink installs (or clears) the test-only bug-capture callback.
// Callers typically use `t.Cleanup(func(){ util.SetBugSink(nil) })` to
// ensure restoration.
func SetBugSink(fn func(msg string)) {
	BugSink = fn
}

// Bug logs a message with a "BUG:" prefix, matching the SMAUG bug() macro.
// If a BugSink is installed it also receives the formatted message.
func Bug(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("BUG: %s", msg)
	if BugSink != nil {
		BugSink(msg)
	}
}

// LogString writes a general log message.
func LogString(str string) {
	log.Print(str)
}

// LogStringPlus writes a log message annotated with a log type and level.
// The type and level values correspond to SMAUG's LOG_* constants and
// immortal trust levels.
func LogStringPlus(str string, logType int, level int) {
	log.Printf("[type=%d level=%d] %s", logType, level, str)
}
