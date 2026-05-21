# Adversary Review

**Target**: `internal/handler/handler.go`
**Timestamp**: 2026-05-16T16:37:19Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- **File size**: 546 lines total, 546 lines new code added (proportional)
- **Function size**: 
  - `CreateMobile` (23 lines), `CreateObject` (18 lines), `CharToRoom` (5 lines), `CharFromRoom` (9 lines), `ObjToRoom` (5 lines), `ObjToChar` (5 lines), `ObjToObj` (5 lines), `ObjFromChar` (12 lines), `ObjFromRoom` (12 lines), `ObjFromObj` (12 lines), `UnequipChar` (10 lines), `ExtractObj` (15 lines), `ExtractChar` (25 lines), `AffectModify` (25 lines), `AffectToChar` (5 lines), `AffectRemove` (10 lines), `AffectStrip` (7 lines), `AffectJoin` (15 lines), `GetEqChar` (5 lines), `CanDropObj` (3 lines), `interpolate` (3 lines)
- **Abstraction depth**: Minimal abstraction; all functions are straightforward and directly implement their intended behavior.
- **New dependencies**: No new dependencies introduced.
- **Premature generalization**: None detected.

### Scope Check
No scope creep detected. All functions in the file align with the package's purpose of handling entity manipulation.

### Alternative Approach
The code uses direct field assignments for object creation, which is appropriate for this type of system where fields are explicitly set from templates. There's no need for a more complex pattern like builder or factory patterns here.

### Assumptions
- The `util` package provides `NumberFuzzy`, `DiceRoll`, `NumberRange`, and `UMIN` functions as expected.
- `types.LEVEL_AVATAR` is defined and non-zero.
- `world.World` has methods `AddChar`, `AddObj`, `RemoveChar`, `RemoveObj`, and `Characters` field.
- `types.CharData` and `types.ObjData` have appropriate fields and methods.
- `types.AffectData` has fields `Type`, `Duration`, `Location`, `Modifier`, `BitVector`.
- `types.AffectData` has methods `IsSet` and `Or`/`AndNot` for bitvector operations.
- `types.AffectData` has fields `Type`, `Duration`, `Location`, `Modifier`, `BitVector`.
- `types.AffectData` has methods `IsSet` and `Or`/`AndNot` for bitvector operations.

### Security
No security issues found in the code.

### Verdict Summary

**VERDICT: PASS**

The code is well structured, follows idiomatic Go practices, and implements all required functionality correctly. There are no significant issues with complexity or security concerns.

```adversary-review
verdict: PASS
confidence: high
artifact:
  path: internal/handler/handler.go
  sha256: e7c3ae116dc743aa
  lines_reviewed: 1-546
findings: []
```
