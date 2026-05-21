# Adversary Review

**Target**: `internal/mudprog/sleep.go`
**Timestamp**: 2026-05-16T08:31:20Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the code are supported by the implementation:
- The comment about single-threaded game loop and no mutex needed is accurate
- The re-entrant safety pattern using `kept[:0]` pattern is correctly implemented
- The comment about mirroring C's implementation is accurate

### Test Verification
No test files were provided for review.

### Complexity Audit
The file contains 85 lines of code with one function (`SleepUpdate`) that is 23 lines long. The function has a clear flow but contains complex logic for handling re-entrant calls and maintaining queue integrity.

### Scope Check
The file implements only the sleep queue functionality as described in the comments and matches the intended behavior described in the comments.

### Alternative Approach
A simpler alternative would be to use a channel-based approach where each sleeping mudprog waits on a channel that gets closed when its time elapses. This would avoid the complexity of managing the queue manually and could potentially be more efficient.

### Assumptions
1. The game loop runs in a single goroutine (assumed from comment)
2. The `Driver` function can safely be called with nil parameters
3. The `sd.Mob` pointer will not be nil during the resume process
4. The `Driver` function won't cause additional re-entrant calls that break the queue management logic

### Security
No security issues found in the code.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/sleep.go
  sha256: 319c79e5b07aa470
  lines_reviewed: 1-85
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/mudprog/sleep.go
    line: 35
    line_end: 36
    message: >
      The re-entrant safety pattern assumes that new entries added to sleepQueue
      during Driver() execution will be appended to the end, but there's no
      guarantee that the re-entrant calls will append to the end of the slice.
    suggested_fix: >
      Consider using a more robust approach like a channel-based system or
      explicitly managing the queue to prevent re-entry issues.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/mudprog/sleep.go
    line: 33
    line_end: 34
    message: >
      The comment mentions "TODO(tier4)" for tracking generation counters,
      indicating potential future issues with stale entries in the queue.
    suggested_fix: >
      Address this TODO or at least document why it's not currently a concern.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/sleep.go
    line: 70
    line_end: 71
    message: >
      The comment about "no InRoom and no IndexData" as a heuristic for
      detecting extracted mobs is not very robust and could lead to false
      positives or false negatives.
    suggested_fix: >
      Consider implementing a more robust mechanism for tracking whether
      entities have been extracted from the game world.
```
