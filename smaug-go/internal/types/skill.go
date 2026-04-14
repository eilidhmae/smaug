package types

// SkillType defines a spell or skill.
// Maps to C struct skill_type (mud.h:3559+).
type SkillType struct {
	Name       string
	SkillLevel [MAX_CLASS]int // Level requirement per class
	SkillAdept [MAX_CLASS]int // Max learnable proficiency per class
	RaceLevel  [MAX_RACE]int  // Level requirement per race
	RaceAdept  [MAX_RACE]int  // Max proficiency per race

	// Function names (resolved to actual functions at runtime)
	SpellFunName string
	SkillFunName string

	Target     int // TAR_* targeting type
	MinimumPos int // Minimum position to use
	Slot       int // For compatibility
	MinMana    int // Minimum mana cost
	Beats      int // Wait time after use (in pulses)
	Range      int // Spell range in rooms

	// Dice formula for damage
	DiceFormula string

	// Messages
	NounDamage string // Damage noun ("fireball")
	MsgOff     string // Wear-off message
	HitChar    string // Message to caster on hit
	HitVict    string // Message to victim on hit
	HitRoom    string // Message to room on hit
	MissChar   string // Message to caster on miss
	MissVict   string // Message to victim on miss
	MissRoom   string // Message to room on miss
	DieChar    string // Message to caster on kill
	DieVict    string // Message to victim on kill
	DieRoom    string // Message to room on kill
	ImmChar    string // Immunity message to caster
	ImmVict    string // Immunity message to victim
	ImmRoom    string // Immunity message to room
	AbsChar    string // Absorb message to caster
	AbsVict    string // Absorb message to victim
	AbsRoom    string // Absorb message to room

	// Spell properties
	Type       int // SKILL_SPELL, SKILL_SKILL, etc.
	Flags      int // SF_* flags
	SaveType   int // Saving throw type
	SaveEffect int // Save effect (negate, half, etc.)
	Difficulty int
	Alignment  int
	Guild      int

	// Affects applied by this spell
	Affects []*SmaugAff

	// Components required
	Components string

	// Teachers
	Teachers string

	// Participants required
	Participants int

	// Info field — packed bitfield holding SpellDamageType/Action/Class/Power
	// (see accessor methods below). Matches C skill_type.info.
	Info int

	// Value field — used by create-object/create-mob spells as the vnum to
	// instantiate. Matches C skill_type.value.
	Value int
}

// Skill type and target type constants are defined in enums.go.

// The SMAUG spell_smaug dispatcher stores five pieces of metadata packed into
// the Info field. These accessors mirror the non-NEWSPELLS C macros
// SPELL_DAMAGE / SPELL_ACTION / SPELL_CLASS / SPELL_POWER / SPELL_SAVE from
// mud.h:4141-4145:
//
//	#define SPELL_DAMAGE(skill) ( ((skill)->info      ) & 7 )
//	#define SPELL_ACTION(skill) ( ((skill)->info >>  3) & 7 )
//	#define SPELL_CLASS(skill)  ( ((skill)->info >>  6) & 7 )
//	#define SPELL_POWER(skill)  ( ((skill)->info >>  9) & 3 )
//	#define SPELL_SAVE(skill)   ( ((skill)->info >> 11) & 7 )
//
// NEWSPELLS is compile-gated in C (#ifdef NEWSPELLS, mud.h:4129) and is OFF in
// the shipping SMAUG build — db/system/en/skills.dat was generated with that
// layout, so the Go port must use the same shifts.
//
// No separate .dat keywords are used; the Info integer is the canonical
// storage and skills.dat expresses these values packed into a single
// "Info N" line (e.g. "Info 840" for benediction == SA_CREATE|SC_LIFE|SP_MINOR,
// "Info 397" for acidmist == SD_ACID|SA_CREATE|SC_DEATH).

// SpellDamageType returns the SD_* damage category packed in Info.
func (s *SkillType) SpellDamageType() int { return s.Info & 7 }

// SpellAction returns the SA_* action category packed in Info.
func (s *SkillType) SpellAction() int { return (s.Info >> 3) & 7 }

// SpellClass returns the SC_* class category packed in Info.
func (s *SkillType) SpellClass() int { return (s.Info >> 6) & 7 }

// SpellPower returns the SP_* power tier packed in Info.
func (s *SkillType) SpellPower() int { return (s.Info >> 9) & 3 }

// SpellSave returns the SE_* save effect packed in Info bits 11-13. C reads
// this from Info; the separate SaveEffect field on SkillType is kept for
// callers that set it explicitly, and the dispatcher prefers Info-bits but
// falls back to SaveEffect when the Info bits are 0 (SE_NONE).
func (s *SkillType) SpellSave() int { return (s.Info >> 11) & 7 }

// HasFlag returns true if the given SF_* flag bit is set on skill.Flags.
// Flags are stored as int for convenient .dat parsing but tested against
// uint32 SF_* constants from constants.go.
func (s *SkillType) HasFlag(flag uint32) bool {
	return uint32(s.Flags)&flag != 0
}
