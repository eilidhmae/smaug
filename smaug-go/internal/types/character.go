package types

import "time"

// CharData represents one character (PC or NPC) in the game world.
// Maps to C struct char_data (mud.h:2725).
type CharData struct {
	// Relationships
	Master   *CharData
	Leader   *CharData
	Fighting *FightData
	Reply    *CharData
	Retell   *CharData
	Switched *CharData
	Mount    *CharData

	// NPC tracking
	Hunting *HHFData
	Fearing *HHFData
	Hating  *HHFData

	// Spec fun name (resolved at runtime)
	SpecFun string

	// MUD prog state
	MpAct       []*MProgActList
	MpActNum    int
	MpScriptPos int

	// Alias-expansion recursion guard. Maps to C char_data.cmd_recurse.
	// -1 means "abort further dispatch for this chain"; otherwise the current
	// nesting depth. See alias.c:check_alias.
	CmdRecurse int

	// Template (NPC only)
	IndexData *MobIndexData

	// Connection (nil for NPCs and linkdead players)
	Desc *DescriptorData

	// Effects
	Affects []*AffectData

	// Notes
	PNote    *NoteData
	Comments *NoteData

	// Inventory (carried + worn items)
	Carrying []*ObjData

	// Location
	InRoom    *RoomIndexData
	WasInRoom *RoomIndexData

	// Player-only data (nil for NPCs)
	PCData *PCData

	// Editor state
	Editor     *EditorData
	EditorSave func(*CharData) // invoked on /s after CON_PLAYING transition

	// Timers
	Timers []*TimerData

	// Morph
	Morph *CharMorph

	// Variables
	Variables []*VariableData

	// Identity
	Name        string
	ShortDescr  string
	LongDescr   string
	Description string

	// Combat state
	NumFighting int

	// Character state
	Substate int
	Sex      int
	Class    int
	Race     int
	Level    int
	Trust    int
	Played   int
	Logon    time.Time
	SaveTime time.Time
	Timer    int
	Wait     int

	// Vitals
	Hit     int
	MaxHit  int
	Mana    int
	MaxMana int
	Move    int
	MaxMove int

	// Misc stats
	Practice   int
	NumAttacks int
	Gold       int
	Silver     int
	Copper     int
	Exp        int

	// Flags (128-bit extended bitvectors)
	Act          BitVector
	AffectedBy   BitVector
	NoAffectedBy BitVector

	// Encumbrance
	CarryWeight int
	CarryNumber int

	// Extra flags (old-style 32-bit)
	XFlags int

	// Immunities
	NoImmune          int
	NoResistant       int
	NoSusceptible     int
	Immune            int
	Resistant         int
	Susceptible       int
	StanceImmune      int
	StanceResistant   int
	StanceSusceptible int

	// Combat flags
	Attacks  BitVector
	Defenses BitVector

	// Languages
	Speaks   int
	Speaking int

	// Saving throws
	SavingPoisonDeath int
	SavingWand        int
	SavingParaPetri   int
	SavingBreath      int
	SavingSpellStaff  int

	// Stats
	Alignment   int
	BareNumDie  int
	BareSizeDie int
	MobThac0    int
	Hitroll     int
	Damroll     int
	Hitplus     int
	Damplus     int
	Position    int
	DefPosition int
	Style       int
	Height      int
	Weight      int
	Armor       int
	Wimpy       int

	// Channels deaf to
	Deaf BitVector

	// Permanent attributes
	PermStr int
	PermInt int
	PermWis int
	PermDex int
	PermCon int
	PermCha int
	PermLck int

	// Attribute modifiers
	ModStr int
	ModInt int
	ModWis int
	ModDex int
	ModCon int
	ModCha int
	ModLck int

	// Mental/emotional state
	MentalState    int
	EmotionalState int

	// Build interface
	PageLen          int
	InterPage        int
	InterType        int
	InterEditing     string
	InterEditingVnum int
	InterSubstate    int

	// Retran/regoto
	Retran int
	Regoto int

	// Mob invis level
	MobInvis int

	// Stances
	Stance  int
	Stances [MAX_STANCE]int

	// Hotboot
	HomeVnum int

	// Marriage
	Spouse string

	// Colors
	Colors []int

	// Quest (from ENABLE_QUEST)
	QuestGiver  *CharData
	QuestPoints int
	NextQuest   int
	Countdown   int
	QuestObj    int
	QuestMob    int

	// Overland
	X      int
	Y      int
	Map    int
	Sector int

	// OLC deletion confirmation (rdelete/odelete/mdelete require a repeated
	// command within a short window). LastDeleteKind is "room", "obj", or "mob".
	LastDeleteVnum int
	LastDeleteKind string
	LastDeleteTime time.Time
}

// FightData tracks active combat. Maps to C struct fighting_data.
type FightData struct {
	Who         *CharData
	Xp          int
	Align       int
	Duration    int
	TimesKilled int
}

// HHFData tracks hunting/hating/fearing. Maps to C struct hunt_hate_fear.
type HHFData struct {
	Name string
	Who  *CharData
}

// EditorData for the string editor. Maps to C struct editor_data.
type EditorData struct {
	NumLines int
	OnLine   int
	Size     int
	Lines    [49]string
}

// TimerData for character timers. Maps to C struct timer_data.
type TimerData struct {
	DoFun string // function name instead of pointer
	Value int
	Type  int
	Count int
}

// CharMorph for morph state on a character. Maps to C struct char_morph.
type CharMorph struct {
	Morph             *MorphData
	AffectedBy        BitVector
	NoAffectedBy      BitVector
	NoImmune          int
	NoResistant       int
	NoSuscept         int
	Immune            int
	Resistant         int
	Suscept           int
	Timer             int
	AC                int
	Blood             int
	Cha               int
	Con               int
	Damroll           int
	Dex               int
	Dodge             int
	Hit               int
	Hitroll           int
	Int               int
	Lck               int
	Mana              int
	Move              int
	Parry             int
	SavingBreath      int
	SavingParaPetri   int
	SavingPoisonDeath int
	SavingSpellStaff  int
	SavingWand        int
	Str               int
	Tumble            int
	Wis               int
}

// MorphData for morph templates. Maps to C struct morph_data.
type MorphData struct {
	Name              string
	ShortDesc         string
	LongDesc          string
	Description       string
	Help              string
	KeyWords          string
	MorphOther        string
	MorphSelf         string
	UnmorphOther      string
	UnmorphSelf       string
	Skills            string
	NoSkills          string
	Deity             string
	Blood             string
	Damroll           string
	Hit               string
	Hitroll           string
	Mana              string
	Move              string
	AffectedBy        BitVector
	NoAffectedBy      BitVector
	Class             int
	DefPos            int
	NoImmune          int
	NoResistant       int
	NoSuscept         int
	Immune            int
	Resistant         int
	Suscept           int
	Obj               [3]int
	Race              int
	Timer             int
	Used              int
	Vnum              int
	AC                int
	BloodUsed         int
	Cha               int
	Con               int
	DayFrom           int
	DayTo             int
	Dex               int
	Dodge             int
	FavourUsed        int
	GloryUsed         int
	HpUsed            int
	Int               int
	Lck               int
	Level             int
	ManaUsed          int
	MoveUsed          int
	Parry             int
	PKill             int
	SavingBreath      int
	SavingParaPetri   int
	SavingPoisonDeath int
	SavingSpellStaff  int
	SavingWand        int
	Sex               int
	Str               int
	TimeFrom          int
	TimeTo            int
	Tumble            int
	Wis               int
	NoCast            bool
	ObjUse            [3]bool
}

// IsNPC returns true if this character is a non-player character.
func (ch *CharData) IsNPC() bool {
	return ch.Act.IsSet(ACT_IS_NPC)
}

// IsImmortal returns true if the character is immortal level or above.
func (ch *CharData) IsImmortal() bool {
	return ch.GetTrust() >= LEVEL_IMMORTAL
}

// GetTrust returns the effective trust level.
func (ch *CharData) GetTrust() int {
	if ch.Trust != 0 {
		return ch.Trust
	}
	if ch.IsNPC() && ch.Level >= LEVEL_IMMORTAL {
		return LEVEL_IMMORTAL
	}
	return ch.Level
}

// Send writes text to the character's descriptor output buffer.
func (ch *CharData) Send(text string) {
	if ch.Desc != nil {
		ch.Desc.WriteToBuffer(text)
	}
}

// Sendf writes formatted text to the character's descriptor.
func (ch *CharData) Sendf(format string, args ...any) {
	if ch.Desc != nil {
		ch.Desc.WriteToBufferf(format, args...)
	}
}

// GetCurrStr returns current strength (perm + mod, clamped 3-25).
func (ch *CharData) GetCurrStr() int {
	v := ch.PermStr + ch.ModStr
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}

// GetCurrInt returns current intelligence.
func (ch *CharData) GetCurrInt() int {
	v := ch.PermInt + ch.ModInt
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}

// GetCurrWis returns current wisdom.
func (ch *CharData) GetCurrWis() int {
	v := ch.PermWis + ch.ModWis
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}

// GetCurrDex returns current dexterity.
func (ch *CharData) GetCurrDex() int {
	v := ch.PermDex + ch.ModDex
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}

// GetCurrCon returns current constitution.
func (ch *CharData) GetCurrCon() int {
	v := ch.PermCon + ch.ModCon
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}

// GetCurrCha returns current charisma.
func (ch *CharData) GetCurrCha() int {
	v := ch.PermCha + ch.ModCha
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}

// GetCurrLck returns current luck.
func (ch *CharData) GetCurrLck() int {
	v := ch.PermLck + ch.ModLck
	if v < 3 {
		return 3
	}
	if v > 25 {
		return 25
	}
	return v
}
