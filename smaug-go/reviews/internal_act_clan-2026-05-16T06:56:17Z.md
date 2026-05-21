# Adversary Review

**Target**: `internal/act/clan.go`
**Timestamp**: 2026-05-16T06:56:17Z
**Model**: local-mlx-deepseek/mlx-community/DeepSeek-Coder-V2-Lite-Instruct-4bit-mlx

```yaml
adversary-review:
  verdict: CONCERNS
  confidence: medium
  artifact:
    path: internal/act/clan.go
    sha256: 1c50151df99d3de5
    lines_reviewed: 1-505
  findings:
    - id: F1
      severity: major
      category: maintainability
      file: internal/act/clan.go
      line: 47
      line_end: 52
      message: >
        Concurrent access to the session map without mutex protection.
        Multiple goroutines can call Store() simultaneously, leading to a
        fatal map race detected at runtime.
      suggested_fix: >
        Wrap reads/writes in sync.RWMutex, or replace with sync.Map.
    - id: F2
      severity: major
      category: error-handling
      file: internal/act/clan.go
      line: 92
      line_end: 92
      message: >
        Error from json.Unmarshal is discarded. Malformed session data will
        silently produce a zero-value Session struct.
      suggested_fix: >
        Return wrapped error: fmt.Errorf("decode session: %w", err)
    - id: F3
      severity: major
      category: maintainability
      file: internal/act/clan.go
      line: 178
      line_end: 178
      message: >
        UnreadNotesFor function does not check if the note is read before returning the count.
      suggested_fix: >
        Add a check to see if the note is read before returning the count.
    - id: F4
      severity: major
      category: maintainability
      file: internal/act/clan.go
      line: 400
      line_end: 400
      message: >
        clanWithdrawAllowed function does not check if the character is a leader or officer before allowing withdrawal.
      suggested_fix: >
        Add a check to see if the character is a leader or officer before allowing withdrawal.
```

```adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/clan.go
  sha256: 1c50151df99d3de5
  lines_reviewed: 1-505
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/clan.go
    line: 47
    line_end: 52
    message: >
      Concurrent access to the session map without mutex protection.
      Multiple goroutines can call Store() simultaneously, leading to a
      fatal map race detected at runtime.
    suggested_fix: >
      Wrap reads/writes in sync.RWMutex, or replace with sync.Map.
  - id: F2
    severity: major
    category: error-handling
    file: internal/act/clan.go
    line: 92
    line_end: 92
    message: >
      Error from json.Unmarshal is discarded. Malformed session data will
      silently produce a zero-value Session struct.
    suggested_fix: >
      Return wrapped error: fmt.Errorf("decode session: %w", err)
  - id: F3
    severity: major
    category: maintainability
    file: internal/act/clan.go
    line: 178
    line_end: 178
    message: >
      UnreadNotesFor function does not check if the note is read before returning the count.
    suggested_fix: >
      Add a check to see if the note is read before returning the count.
  - id: F4
    severity: major
    category: maintainability
    file: internal/act/clan.go
    line: 400
    line_end: 400
    message: >
      clanWithdrawAllowed function does not check if the character is a leader or officer before allowing withdrawal.
    suggested_fix: >
      Add a check to see if the character is a leader or officer before allowing withdrawal.
```
