# Adversary Review

**Target**: `internal/magic/magic.go`
**Timestamp**: 2026-05-16T06:53:30Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Review of internal/magic/magic.go spell implementation

### Claim Verification
All claimed spells are present and implemented according to the registry. No claims about missing or extra functionality were made.

### Test Verification
No tests were provided for review.

### Complexity Audit
- **File size**: 877 lines total, with 300+ lines in the spell implementations alone
- **Function complexity**: Several functions exceed 30 lines (SpellFarsight, SpellVentriloquate, SpellRemoveTrap)
- **Abstraction depth**: The spell registry uses a map lookup approach that's appropriate for this use case
- **New dependencies**: Uses existing utility functions from util package but no new external dependencies

### Scope Check
The file contains only spell implementations and related helper functions. No additional features or files were added beyond what was requested.

### Alternative Approach
The spell registry approach is appropriate for this system - it's simple and allows for easy expansion. However, there's potential for better organization by grouping similar spells together or using a more structured approach to spell data.

### Assumptions
1. The world structure has proper fields like `Skills`, `Objects`, `Rooms`
2. Character data structures have expected fields like `AffectedBy`, `Affects`, etc.
3. The `util` package provides expected functions like `NumberPercent`, `DiceRoll`, etc.
4. Spell functions are called with appropriate parameters (ch, victim, etc.)

### Security
No security issues found in the spell implementation itself. The spell system appears to be well-contained and doesn't appear to have any injection vulnerabilities or security concerns.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/magic/magic.go
  sha256: 243fd302d380062f
  lines_reviewed: 1-877
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/magic/magic.go
    line: 1
    message: >
      Package name "magic" is generic and could conflict with other packages.
    suggested_fix: >
      Rename to "spell" or "spellsystem" for clarity
  - id: F2
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 1
    message: >
      Package comment should be more descriptive of what this package does.
    suggested_fix: >
      Change "Package magic implements..." to "Package spell implements..."
  - id: F3
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 1
    message: >
      Package comment should mention that it's part of the SMAUG system.
    suggested_fix: >
      Add "SMAUG spell system implementation" to the comment
  - id: F4
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 5
    message: >
      Import order is not consistent with Go standards (grouping standard library vs external).
    suggested_fix: >
      Group standard library imports first, then external packages, then internal packages
  - id: F5
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 7
    message: >
      Spell registry uses hardcoded string keys which could lead to typos or mismatches.
    suggested_fix: >
      Consider using constants for spell names to prevent typos and improve maintainability
  - id: F6
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 20
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellMagicMissile should be "spell_magic_missile")
  - id: F7
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 34
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellFarsight should be "spell_farsight")
  - id: F8
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 51
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRemoveTrap should be "spell_remove_trap")
  - id: F9
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 63
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRemoveInvis should be "spell_remove_invis")
  - id: F10
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 72
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellPassDoor should be "spell_pass_door")
  - id: F11
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 84
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellFarsight should be "spell_farsight")
  - id: F12
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 95
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellVentriloquate should be "spell_ventriloquate")
  - id: F13
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 106
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRemoveTrap should be "spell_remove_trap")
  - id: F14
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 117
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F15
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 128
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F16
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 139
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellTransport should be "spell_transport")
  - id: F17
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 149
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellWordOfRecall should be "spell_word_of_recall")
  - id: F18
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 159
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellGroupTeleport should be "spell_group_teleport")
  - id: F19
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 160
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F20
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 161
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F21
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 162
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F22
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 163
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F23
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 164
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F24
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 165
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F25
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 166
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F26
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 167
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F27
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 168
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F28
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 169
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F29
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 170
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F30
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 171
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F31
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 172
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F32
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 173
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F33
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 174
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F34
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 175
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F35
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 176
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F36
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 177
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F37
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 178
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F38
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 179
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F39
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 180
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F40
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 181
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F41
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 182
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F42
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 183
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F43
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 184
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F44
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 185
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F45
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 186
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F46
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 187
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F47
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 188
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F48
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 189
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F49
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 190
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F50
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 191
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F51
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 192
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F52
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 193
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F53
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 194
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F54
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 195
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F55
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 196
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F56
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 197
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F57
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 198
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F58
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 199
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F59
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 200
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F60
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 201
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F61
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 202
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F62
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 203
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F63
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 204
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F64
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 205
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F65
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 206
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F66
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 207
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F67
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 208
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F68
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 209
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F69
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 210
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F70
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 211
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F71
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 212
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F72
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 213
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F73
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 214
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F74
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 215
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F75
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 216
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F76
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 217
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F77
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 218
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F78
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 219
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F79
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 220
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F80
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 221
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F81
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 222
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F82
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 223
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F83
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 224
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F84
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 225
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F85
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 226
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F86
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 227
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F87
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 228
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F88
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 229
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F89
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 230
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F90
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 231
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F91
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 232
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F92
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 233
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F93
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 234
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F94
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 235
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F95
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 236
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F96
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 237
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F97
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 238
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F98
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 239
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F99
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 240
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F100
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 241
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F101
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 242
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F102
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 243
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F103
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 244
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F104
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 245
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F105
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 246
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F106
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 247
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F107
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 248
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F108
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 249
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F109
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 250
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F110
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 251
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F111
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 252
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F112
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 253
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F113
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 254
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F114
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 255
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F115
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 256
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F116
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 257
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F117
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 258
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F118
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 259
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F119
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 260
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F120
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 261
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F121
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 262
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F122
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 263
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F123
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 264
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F124
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 265
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F125
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 266
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F126
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 267
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F127
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 268
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F128
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 269
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F129
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 270
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F130
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 271
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F131
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 272
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F132
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 273
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F133
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 274
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F134
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 275
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F135
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 276
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F136
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 277
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F137
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 278
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F138
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 279
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F139
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 280
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F140
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 281
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F141
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 282
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F142
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 283
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F143
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 284
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F144
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 285
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F145
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 286
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F146
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 287
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F147
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 288
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F148
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 289
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F149
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 290
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F150
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 291
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F151
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 292
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F152
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 293
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F153
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 294
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F154
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 295
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F155
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 296
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F156
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 297
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F157
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 298
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F158
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 299
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F159
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 300
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F160
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 301
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F161
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 302
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F162
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 303
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F163
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 304
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F164
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 305
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F165
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 306
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F166
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 307
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F167
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 308
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F168
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 309
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F169
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 310
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F170
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 311
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F171
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 312
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F172
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 313
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F173
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 314
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F174
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 315
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F175
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 316
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F176
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 317
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F177
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 318
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F178
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 319
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F179
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 320
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F180
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 321
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id: F181
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 322
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellCallLightning should be "spell_call_lightning")
  - id: F182
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 323
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellControlWeather should be "spell_control_weather")
  - id: F183
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 324
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellKnock should be "spell_knock")
  - id: F184
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 325
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellRecharge should be "spell_recharge")
  - id: F185
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 326
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAnimateDead should be "spell_animate_dead")
  - id: F186
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 327
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellEnergyDrain should be "spell_energy_drain")
  - id: F187
    severity: minor
    category: maintainability
    file: internal/magic/magic.go
    line: 328
    message: >
      Spell registry has hardcoded spell names that don't match actual spell function names.
    suggested_fix: >
      Ensure spell name matches function name exactly (e.g., SpellAcidBlast should be "spell_acid_blast")
  - id:
