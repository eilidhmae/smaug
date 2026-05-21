# Adversary Review

**Target**: `internal/net/server.go`
**Timestamp**: 2026-05-16T18:08:32Z
**Model**: local-mlx-dscoder/mlx-community/deepseek-coder-33b-instruct-hf-4bit-mlx


```go
adversary-review
verdict: FAIL
confidence: high
artifact:
  path: internal/net/server.go
  sha256: 6b0febabd42f8a28
  lines_reviewed: 1-182
findings:
  - id: F1
    severity: critical
    category: race-condition
    file: internal/net/server.go
    line: 47
    line_end: 52
    message: >
      Concurrent access to the server map without mutex protection.
      Multiple goroutines can call Store() simultaneously, leading to a fatal map race detected at runtime.
    suggested_fix: >
      Wrap reads/writes in sync.RWMutex, or replace with sync.Map.
  - id: F2
    severity: major
    category: error-handling
    file: internal/net/server.go
    line: 92
    line_end: 92
    message: >
      Error from json.Unmarshal is discarded. Malformed session data will silently produce a zero-value Session struct.
    suggested_fix: >
      Return wrapped error: fmt.Errorf("decodesession:%w",err)
  - id: F3
    severity: minor
    category: correctness
    file: internal/net/server.go
    line: 178
    line_end: 178
    message: >
      The server's listener is not closed before rewriting it. This can lead to a race between the old goroutine's close and the new goroutine's accept.
    suggested_fix: >
      Add a lock around the listener rewriting to ensure it's not being closed while the new goroutine is accepting.
mechanical_baseline: ran: true
passed: false
failures: "- gotest: race catches it deterministically under TestResumeFromPause_RestartsAccept"
```

The code review found several issues, including a race condition in the server map, an error from json.Unmarshal being discarded, and a potential race between the old goroutine's close and the new goroutine's accept. The suggested fixes include wrapping reads/writes in sync.RWMutex, returning a wrapped error from json.Unmarshal, and adding a lock around the listener rewriting.
