package util

import (
	"fmt"
	"log"
)

// Bug logs a message with a "BUG:" prefix, matching the SMAUG bug() macro.
func Bug(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("BUG: %s", msg)
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
