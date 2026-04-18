# Changelog

## 2026-04-17

- Post-Phase-5 audit: 4 parallel adversary reviews (phase completeness, Go idiom, security + usability, C↔Go capability comparison). Overall verdict CONCERNS — no regressions, but documented combat-depth, config-command, and documentation-accuracy gaps. Findings in `smaug-go/doc/audit-2026-04-17.md`; follow-ups in `TODO.md`. Commit `720278d`.
- Post-audit batch 1 — 6 independent fixes (each test-first, adversary-verified): `mpForce` trust cap, `DoSaveArea` atomic write, `MaxAliases=50`, boot fail-loud on foundational subsystems, brute-force disconnect message wording, `TestSpellFarsight_Success` deterministic. Added positive-direction tests with mutation verification (`TestMpForce_PositiveDispatch`, `TestSpellFarsight_NPCTargetSucceeds`) and a `savesSpellStaffFn` seam in `magic/magic.go`. Files affected: `internal/act/{cmds2,olc}.go`, `internal/boot/boot.go`, `internal/game/loop.go`, `internal/magic/magic.go`, `internal/mudprog/commands.go` (+ tests). Commit `9b789a7`.
