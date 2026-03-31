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

	// Info field
	Info int
}

// Skill type and target type constants are defined in enums.go.
