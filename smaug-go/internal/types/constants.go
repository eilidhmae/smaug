package types

// String and memory management parameters.
const (
	MAX_KEY_HASH      = 2048
	MAX_STRING_LENGTH = 8192
	MAX_INPUT_LENGTH  = 1024
	MAX_INBUF_SIZE    = 1024
	MSL               = MAX_STRING_LENGTH
	MIL               = MAX_INPUT_LENGTH
)

// General game limits.
const (
	MAX_LAYERS      = 8
	MAX_NEST        = 100
	MAX_KILLTRACK   = 25
	MAX_VNUM        = 1000000000
	MAX_RGRID_ROOMS = 30000
)

// Game parameters.
const (
	MAX_EXP_WORTH = 500000
	MIN_EXP_WORTH = 20

	MAX_REXITS    = 20
	MAX_SKILL     = 600
	MAX_CLASS     = 27
	MAX_NPC_CLASS = 27
	MAX_RACE      = 26
	MAX_NPC_RACE  = 200
	MAX_MSG       = 18
)

// Level hierarchy.
const (
	MAX_LEVEL         = 65
	LEVEL_SUPREME     = MAX_LEVEL
	LEVEL_INFINITE    = MAX_LEVEL - 1
	LEVEL_ETERNAL     = MAX_LEVEL - 2
	LEVEL_IMPLEMENTOR = MAX_LEVEL - 3
	LEVEL_SUB_IMPLEM  = MAX_LEVEL - 4
	LEVEL_ASCENDANT   = MAX_LEVEL - 5
	LEVEL_GREATER     = MAX_LEVEL - 6
	LEVEL_GOD         = MAX_LEVEL - 7
	LEVEL_LESSER      = MAX_LEVEL - 8
	LEVEL_TRUEIMM     = MAX_LEVEL - 9
	LEVEL_DEMI        = MAX_LEVEL - 10
	LEVEL_SAVIOR      = MAX_LEVEL - 11
	LEVEL_CREATOR     = MAX_LEVEL - 12
	LEVEL_ACOLYTE     = MAX_LEVEL - 13
	LEVEL_NEOPHYTE    = MAX_LEVEL - 14
	LEVEL_IMMORTAL    = MAX_LEVEL - 14
	LEVEL_HERO        = MAX_LEVEL - 15
	LEVEL_AVATAR      = MAX_LEVEL - 15
	LEVEL_LOG         = LEVEL_LESSER
	LEVEL_HIGOD       = LEVEL_GOD
)

// Pulse timing.
const (
	SECONDS_PER_TICK = 70
	PULSE_PER_SECOND = 4
	PULSE_VIOLENCE   = 3 * PULSE_PER_SECOND
	PULSE_MOBILE     = 4 * PULSE_PER_SECOND
	PULSE_TICK       = SECONDS_PER_TICK * PULSE_PER_SECOND
	PULSE_AREA       = 60 * PULSE_PER_SECOND
	PULSE_AUCTION    = 9 * PULSE_PER_SECOND
	PULSE_CASINO     = 8 * PULSE_PER_SECOND
	PULSE_SAVE       = 5 * 60 * PULSE_PER_SECOND // Autosave every 5 minutes
)

// Area versions.
const (
	AREA_VERSION_WRITE = 4
	MIN_SAVE_VERSION   = 4
	HAS_SPELL_INDEX    = -1
)

// Other game limits.
const (
	MAX_CLAN             = 50
	MAX_DEITY            = 50
	MAX_CPD              = 11
	MAX_HERB             = 20
	MAX_MOB_STANCE       = 200
	MAX_PC_STANCE        = 200
	MAX_DISEASE          = 20
	MAX_PERSONAL         = 5
	MAX_WHERE_NAME       = 32
	MAX_OINVOKE_QUANTITY = 50
	MAX_NUISANCE_STAGE   = 10
)

// Duration conversion factor.
const DUR_CONV = 23.333333333333333333333333

// Item condition and weapon defaults.
const (
	INIT_WEAPON_CONDITION = 12
	MAX_ITEM_IMPACT       = 30
)

// Shop constants.
const (
	MAX_TRADE     = 5
	MAX_FIX       = 3
	SHOP_FIX      = 1
	SHOP_RECHARGE = 2
)

// Mud prog defines.
const (
	MAX_IFS       = 20
	IN_IF         = 0
	IN_ELSE       = 1
	DO_IF         = 2
	DO_ELSE       = 3
	MAX_PROG_NEST = 20
)

// Echo targets for echo_to_all.
const (
	ECHOTAR_ALL = 0
	ECHOTAR_PC  = 1
	ECHOTAR_IMM = 2
	ECHOTAR_PK  = 3
)

// Who types for do_who.
const (
	WT_MORTAL   = 0
	WT_DEADLY   = 1
	WT_IMM      = 2
	WT_GROUPED  = 3
	WT_GROUPWHO = 4
)

// Ban types.
const (
	BAN_SITE  = 1
	BAN_CLASS = 2
	BAN_RACE  = 3
	BAN_WARN  = -1
)

// Extended bitvector parameters.
// IntBits (32), XBI (4), and MaxBits (128) are defined in bitvector.go.
const (
	XBM = 31 // extended bitmask (IntBits - 1)
	RSV = 5  // right-shift value (log2(IntBits))
)

// Morph restrictions.
const (
	ONLY_PKILL     = 1
	ONLY_PEACEFULL = 2
)

// Language flags.
const (
	LANG_COMMON    uint32 = 1 << 0
	LANG_ELVEN     uint32 = 1 << 1
	LANG_DWARVEN   uint32 = 1 << 2
	LANG_PIXIE     uint32 = 1 << 3
	LANG_OGRE      uint32 = 1 << 4
	LANG_ORCISH    uint32 = 1 << 5
	LANG_TROLLISH  uint32 = 1 << 6
	LANG_RODENT    uint32 = 1 << 7
	LANG_INSECTOID uint32 = 1 << 8
	LANG_MAMMAL    uint32 = 1 << 9
	LANG_REPTILE   uint32 = 1 << 10
	LANG_DRAGON    uint32 = 1 << 11
	LANG_SPIRITUAL uint32 = 1 << 12
	LANG_MAGICAL   uint32 = 1 << 13
	LANG_GOBLIN    uint32 = 1 << 14
	LANG_GOD       uint32 = 1 << 15
	LANG_ANCIENT   uint32 = 1 << 16
	LANG_HALFLING  uint32 = 1 << 17
	LANG_CLAN      uint32 = 1 << 18
	LANG_GITH      uint32 = 1 << 19
	LANG_GNOME     uint32 = 1 << 20
	LANG_UNKNOWN   uint32 = 0
)

// VALID_LANGS is the combination of all valid player languages.
const VALID_LANGS = LANG_COMMON | LANG_ELVEN | LANG_DWARVEN | LANG_PIXIE |
	LANG_OGRE | LANG_ORCISH | LANG_TROLLISH | LANG_GOBLIN |
	LANG_HALFLING | LANG_GITH | LANG_GNOME

// Resistant/Immune/Susceptible flags.
const (
	RIS_FIRE        uint32 = 1 << 0
	RIS_COLD        uint32 = 1 << 1
	RIS_ELECTRICITY uint32 = 1 << 2
	RIS_ENERGY      uint32 = 1 << 3
	RIS_BLUNT       uint32 = 1 << 4
	RIS_PIERCE      uint32 = 1 << 5
	RIS_SLASH       uint32 = 1 << 6
	RIS_ACID        uint32 = 1 << 7
	RIS_POISON      uint32 = 1 << 8
	RIS_DRAIN       uint32 = 1 << 9
	RIS_SLEEP       uint32 = 1 << 10
	RIS_CHARM       uint32 = 1 << 11
	RIS_HOLD        uint32 = 1 << 12
	RIS_NONMAGIC    uint32 = 1 << 13
	RIS_PLUS1       uint32 = 1 << 14
	RIS_PLUS2       uint32 = 1 << 15
	RIS_PLUS3       uint32 = 1 << 16
	RIS_PLUS4       uint32 = 1 << 17
	RIS_PLUS5       uint32 = 1 << 18
	RIS_PLUS6       uint32 = 1 << 19
	RIS_MAGIC       uint32 = 1 << 20
	RIS_PARALYSIS   uint32 = 1 << 21
)

// Body part flags.
const (
	PART_HEAD        uint32 = 1 << 0
	PART_ARMS        uint32 = 1 << 1
	PART_LEGS        uint32 = 1 << 2
	PART_HEART       uint32 = 1 << 3
	PART_BRAINS      uint32 = 1 << 4
	PART_GUTS        uint32 = 1 << 5
	PART_HANDS       uint32 = 1 << 6
	PART_FEET        uint32 = 1 << 7
	PART_FINGERS     uint32 = 1 << 8
	PART_EAR         uint32 = 1 << 9
	PART_EYE         uint32 = 1 << 10
	PART_LONG_TONGUE uint32 = 1 << 11
	PART_EYESTALKS   uint32 = 1 << 12
	PART_TENTACLES   uint32 = 1 << 13
	PART_FINS        uint32 = 1 << 14
	PART_WINGS       uint32 = 1 << 15
	PART_TAIL        uint32 = 1 << 16
	PART_SCALES      uint32 = 1 << 17
	PART_CLAWS       uint32 = 1 << 18
	PART_FANGS       uint32 = 1 << 19
	PART_HORNS       uint32 = 1 << 20
	PART_TUSKS       uint32 = 1 << 21
	PART_TAILATTACK  uint32 = 1 << 22
	PART_SHARPSCALES uint32 = 1 << 23
	PART_BEAK        uint32 = 1 << 24
	PART_HAUNCH      uint32 = 1 << 25
	PART_HOOVES      uint32 = 1 << 26
	PART_PAWS        uint32 = 1 << 27
	PART_FORELEGS    uint32 = 1 << 28
	PART_FEATHERS    uint32 = 1 << 29
	PART_HUSK_SHELL  uint32 = 1 << 30
)

// Autosave flags.
const (
	SV_DEATH      uint32 = 1 << 0
	SV_KILL       uint32 = 1 << 1
	SV_PASSCHG    uint32 = 1 << 2
	SV_DROP       uint32 = 1 << 3
	SV_PUT        uint32 = 1 << 4
	SV_GIVE       uint32 = 1 << 5
	SV_AUTO       uint32 = 1 << 6
	SV_ZAPDROP    uint32 = 1 << 7
	SV_AUCTION    uint32 = 1 << 8
	SV_GET        uint32 = 1 << 9
	SV_RECEIVE    uint32 = 1 << 10
	SV_IDLE       uint32 = 1 << 11
	SV_BACKUP     uint32 = 1 << 12
	SV_QUITBACKUP uint32 = 1 << 13
	SV_FILL       uint32 = 1 << 14
	SV_EMPTY      uint32 = 1 << 15
	SV_TMPSAVE    uint32 = 1 << 16
)

// Pipe flags.
const (
	PIPE_TAMPED    uint32 = 1 << 1
	PIPE_LIT       uint32 = 1 << 2
	PIPE_HOT       uint32 = 1 << 3
	PIPE_DIRTY     uint32 = 1 << 4
	PIPE_FILTHY    uint32 = 1 << 5
	PIPE_GOINGOUT  uint32 = 1 << 6
	PIPE_BURNT     uint32 = 1 << 7
	PIPE_FULLOFASH uint32 = 1 << 8
)

// Skill/Spell flags.
const (
	SF_WATER        uint32 = 1 << 0
	SF_EARTH        uint32 = 1 << 1
	SF_AIR          uint32 = 1 << 2
	SF_ASTRAL       uint32 = 1 << 3
	SF_AREA         uint32 = 1 << 4
	SF_DISTANT      uint32 = 1 << 5
	SF_REVERSE      uint32 = 1 << 6
	SF_NOSELF       uint32 = 1 << 7
	SF_UNUSED2      uint32 = 1 << 8
	SF_ACCUMULATIVE uint32 = 1 << 9
	SF_RECASTABLE   uint32 = 1 << 10
	SF_NOSCRIBE     uint32 = 1 << 11
	SF_NOBREW       uint32 = 1 << 12
	SF_GROUPSPELL   uint32 = 1 << 13
	SF_OBJECT       uint32 = 1 << 14
	SF_CHARACTER    uint32 = 1 << 15
	SF_SECRETSKILL  uint32 = 1 << 16
	SF_PKSENSITIVE  uint32 = 1 << 17
	SF_STOPONFAIL   uint32 = 1 << 18
	SF_NOFIGHT      uint32 = 1 << 19
	SF_NODISPEL     uint32 = 1 << 20
	SF_RANDOMTARGET uint32 = 1 << 21
	SF_NOMOB        uint32 = 1 << 22
)

// Trap trigger flags.
const (
	TRAP_ROOM       uint32 = 1 << 0
	TRAP_OBJ        uint32 = 1 << 1
	TRAP_ENTER_ROOM uint32 = 1 << 2
	TRAP_LEAVE_ROOM uint32 = 1 << 3
	TRAP_OPEN       uint32 = 1 << 4
	TRAP_CLOSE      uint32 = 1 << 5
	TRAP_GET        uint32 = 1 << 6
	TRAP_PUT        uint32 = 1 << 7
	TRAP_PICK       uint32 = 1 << 8
	TRAP_UNLOCK     uint32 = 1 << 9
	TRAP_N          uint32 = 1 << 10
	TRAP_S          uint32 = 1 << 11
	TRAP_E          uint32 = 1 << 12
	TRAP_W          uint32 = 1 << 13
	TRAP_U          uint32 = 1 << 14
	TRAP_D          uint32 = 1 << 15
	TRAP_EXAMINE    uint32 = 1 << 16
	TRAP_NE         uint32 = 1 << 17
	TRAP_NW         uint32 = 1 << 18
	TRAP_SE         uint32 = 1 << 19
	TRAP_SW         uint32 = 1 << 20
)

// Class constants.
const (
	CLASS_NONE        = -1
	CLASS_MAGE        = 0
	CLASS_CLERIC      = 1
	CLASS_THIEF       = 2
	CLASS_WARRIOR     = 3
	CLASS_VAMPIRE     = 4
	CLASS_DRUID       = 5
	CLASS_RANGER      = 6
	CLASS_AUGURER     = 7
	CLASS_PALADIN     = 8
	CLASS_NEPHANDI    = 9
	CLASS_SAVAGE      = 10
	CLASS_FATHOMER    = 11
	CLASS_ARCHER      = 12
	CLASS_DEMON       = 13
	CLASS_ASSASSIN    = 14
	CLASS_ANGEL       = 15
	CLASS_WEREWOLF    = 16
	CLASS_LICANTHROPE = 17
	CLASS_LICH        = 18
	CLASS_MONGER      = 19
	CLASS_PIRATE      = 20
)

// NPC race constants.
const RACE_DRAGON = 31

// Well known mob virtual numbers.
const (
	MOB_VNUM_CITYGUARD       = 3060
	MOB_VNUM_VAMPIRE         = 80
	MOB_VNUM_ANIMATED_CORPSE = 5
	MOB_VNUM_POLY_WOLF       = 10
	MOB_VNUM_POLY_MIST       = 11
	MOB_VNUM_POLY_BAT        = 12
	MOB_VNUM_POLY_HAWK       = 13
	MOB_VNUM_POLY_CAT        = 14
	MOB_VNUM_POLY_DOVE       = 15
	MOB_VNUM_POLY_FISH       = 16
	MOB_VNUM_DEITY           = 17
)

// Well known object virtual numbers.
const (
	OBJ_VNUM_MONEY_ONE       = 2
	OBJ_VNUM_MONEY_SOME      = 3
	OBJ_VNUM_CORPSE_NPC      = 10
	OBJ_VNUM_CORPSE_PC       = 11
	OBJ_VNUM_SEVERED_HEAD    = 12
	OBJ_VNUM_TORN_HEART      = 13
	OBJ_VNUM_SLICED_ARM      = 14
	OBJ_VNUM_SLICED_LEG      = 15
	OBJ_VNUM_SPILLED_GUTS    = 16
	OBJ_VNUM_BLOOD           = 17
	OBJ_VNUM_BLOODSTAIN      = 18
	OBJ_VNUM_SCRAPS          = 19
	OBJ_VNUM_MUSHROOM        = 20
	OBJ_VNUM_LIGHT_BALL      = 21
	OBJ_VNUM_SPRING          = 22
	OBJ_VNUM_SKIN            = 23
	OBJ_VNUM_SLICE           = 24
	OBJ_VNUM_SHOPPING_BAG    = 25
	OBJ_VNUM_BLOODLET        = 26
	OBJ_VNUM_FIRE            = 30
	OBJ_VNUM_TRAP            = 31
	OBJ_VNUM_PORTAL          = 32
	OBJ_VNUM_BLACK_POWDER    = 33
	OBJ_VNUM_SCROLL_SCRIBING = 34
	OBJ_VNUM_FLASK_BREWING   = 35
	OBJ_VNUM_NOTE            = 36
	OBJ_VNUM_DEITY           = 64
	OBJ_VNUM_BLOOD_SPLATTER  = 94
	OBJ_VNUM_PUDDLE          = 95
)

// Academy equipment vnums.
const (
	OBJ_VNUM_SCHOOL_MACE   = 10315
	OBJ_VNUM_SCHOOL_DAGGER = 10312
	OBJ_VNUM_SCHOOL_SWORD  = 10313
	OBJ_VNUM_SCHOOL_VEST   = 10308
	OBJ_VNUM_SCHOOL_SHIELD = 10310
	OBJ_VNUM_SCHOOL_BANNER = 10311
)

// Well known room virtual numbers.
const (
	ROOM_VNUM_LIMBO        = 2
	ROOM_VNUM_POLY         = 3
	ROOM_VNUM_CHAT         = 1200
	ROOM_VNUM_TEMPLE       = 21001
	ROOM_VNUM_ALTAR        = 21194
	ROOM_VNUM_SCHOOL       = 10300
	ROOM_AUTH_START        = 100
	ROOM_VNUM_HALLOFFALLEN = 21195
)

// ACT bits for mobs (extended bitvector bit numbers).
const (
	ACT_IS_NPC        = 0
	ACT_SENTINEL      = 1
	ACT_SCAVENGER     = 2
	ACT_NOLOCATE      = 3
	ACT_AGGRESSIVE    = 5
	ACT_STAY_AREA     = 6
	ACT_WIMPY         = 7
	ACT_PET           = 8
	ACT_TRAIN         = 9
	ACT_PRACTICE      = 10
	ACT_IMMORTAL      = 11
	ACT_DEADLY        = 12
	ACT_POLYSELF      = 13
	ACT_META_AGGR     = 14
	ACT_GUARDIAN      = 15
	ACT_RUNNING       = 16
	ACT_NOWANDER      = 17
	ACT_MOUNTABLE     = 18
	ACT_MOUNTED       = 19
	ACT_SCHOLAR       = 20
	ACT_SECRETIVE     = 21
	ACT_HARDHAT       = 22
	ACT_MOBINVIS      = 23
	ACT_NOASSIST      = 24
	ACT_AUTONOMOUS    = 25
	ACT_PACIFIST      = 26
	ACT_NOATTACK      = 27
	ACT_ANNOYING      = 28
	ACT_STATSHIELD    = 29
	ACT_PROTOTYPE     = 30
	ACT_NOSUMMON      = 31
	ACT_NOSTEAL       = 32
	ACT_INFEST        = 34
	ACT_BLOCKING      = 36
	ACT_IS_CLONE      = 37
	ACT_IS_DREAMFORM  = 38
	ACT_IS_SPIRITFORM = 39
	ACT_IS_PROJECTION = 40
	ACT_STOP_SCRIPT   = 41
	ACT_BANKER        = 42
	ACT_UNDERTAKER    = 43
	ACT_CHALLENGED    = 44
	ACT_CHALLENGER    = 45
	ACT_QUESTMASTER   = 46
	ACT_ONMAP         = 47
	MAX_ACT_FLAGS     = 48
)

// Color type constants (raw ANSI codes).
const (
	AT_COLORIZE = -1
	AT_BLACK    = 0
	AT_BLOOD    = 1
	AT_DGREEN   = 2
	AT_ORANGE   = 3
	AT_DBLUE    = 4
	AT_PURPLE   = 5
	AT_CYAN     = 6
	AT_GREY     = 7
	AT_DGREY    = 8
	AT_RED      = 9
	AT_GREEN    = 10
	AT_YELLOW   = 11
	AT_BLUE     = 12
	AT_PINK     = 13
	AT_LBLUE    = 14
	AT_WHITE    = 15
	AT_BLINK    = 16
)

// Magic item flags.
const (
	ITEM_RETURNING       uint32 = 1 << 0
	ITEM_BACKSTABBER     uint32 = 1 << 1
	ITEM_BANE            uint32 = 1 << 2
	ITEM_MAGIC_LOYAL     uint32 = 1 << 3
	ITEM_HASTE           uint32 = 1 << 4
	ITEM_DRAIN           uint32 = 1 << 5
	ITEM_LIGHTNING_BLADE uint32 = 1 << 6
	ITEM_PKDISARMED      uint32 = 1 << 7
)

// Lever/dial/switch/button/pullchain trigger flags.
const (
	TRIG_UP           uint32 = 1 << 0
	TRIG_UNLOCK       uint32 = 1 << 1
	TRIG_LOCK         uint32 = 1 << 2
	TRIG_D_NORTH      uint32 = 1 << 3
	TRIG_D_SOUTH      uint32 = 1 << 4
	TRIG_D_EAST       uint32 = 1 << 5
	TRIG_D_WEST       uint32 = 1 << 6
	TRIG_D_UP         uint32 = 1 << 7
	TRIG_D_DOWN       uint32 = 1 << 8
	TRIG_DOOR         uint32 = 1 << 9
	TRIG_CONTAINER    uint32 = 1 << 10
	TRIG_OPEN         uint32 = 1 << 11
	TRIG_CLOSE        uint32 = 1 << 12
	TRIG_PASSAGE      uint32 = 1 << 13
	TRIG_OLOAD        uint32 = 1 << 14
	TRIG_MLOAD        uint32 = 1 << 15
	TRIG_TELEPORT     uint32 = 1 << 16
	TRIG_TELEPORTALL  uint32 = 1 << 17
	TRIG_TELEPORTPLUS uint32 = 1 << 18
	TRIG_DEATH        uint32 = 1 << 19
	TRIG_CAST         uint32 = 1 << 20
	TRIG_FAKEBLADE    uint32 = 1 << 21
	TRIG_RAND4        uint32 = 1 << 22
	TRIG_RAND6        uint32 = 1 << 23
	TRIG_TRAPDOOR     uint32 = 1 << 24
	TRIG_ANOTHEROOM   uint32 = 1 << 25
	TRIG_USEDIAL      uint32 = 1 << 26
	TRIG_ABSOLUTEVNUM uint32 = 1 << 27
	TRIG_SHOWROOMDESC uint32 = 1 << 28
	TRIG_AUTORETURN   uint32 = 1 << 29
)

// Teleport flags.
const (
	TELE_SHOWDESC     uint32 = 1 << 0
	TELE_TRANSALL     uint32 = 1 << 1
	TELE_TRANSALLPLUS uint32 = 1 << 2
)

// Wear flags (on objects).
const (
	ITEM_TAKE          uint32 = 1 << 0
	ITEM_WEAR_FINGER   uint32 = 1 << 1
	ITEM_WEAR_NECK     uint32 = 1 << 2
	ITEM_WEAR_BODY     uint32 = 1 << 3
	ITEM_WEAR_HEAD     uint32 = 1 << 4
	ITEM_WEAR_LEGS     uint32 = 1 << 5
	ITEM_WEAR_FEET     uint32 = 1 << 6
	ITEM_WEAR_HANDS    uint32 = 1 << 7
	ITEM_WEAR_ARMS     uint32 = 1 << 8
	ITEM_WEAR_SHIELD   uint32 = 1 << 9
	ITEM_WEAR_ABOUT    uint32 = 1 << 10
	ITEM_WEAR_WAIST    uint32 = 1 << 11
	ITEM_WEAR_WRIST    uint32 = 1 << 12
	ITEM_WIELD         uint32 = 1 << 13
	ITEM_HOLD          uint32 = 1 << 14
	ITEM_DUAL_WIELD    uint32 = 1 << 15
	ITEM_WEAR_EARS     uint32 = 1 << 16
	ITEM_WEAR_EYES     uint32 = 1 << 17
	ITEM_MISSILE_WIELD uint32 = 1 << 18
	ITEM_WEAR_BACK     uint32 = 1 << 19
	ITEM_WEAR_FACE     uint32 = 1 << 20
	ITEM_WEAR_ANKLE    uint32 = 1 << 21
	ITEM_WEAR_MAX             = 21
)

// Exit flags.
const (
	EX_ISDOOR      uint32 = 1 << 0
	EX_CLOSED      uint32 = 1 << 1
	EX_LOCKED      uint32 = 1 << 2
	EX_SECRET      uint32 = 1 << 3
	EX_SWIM        uint32 = 1 << 4
	EX_PICKPROOF   uint32 = 1 << 5
	EX_FLY         uint32 = 1 << 6
	EX_CLIMB       uint32 = 1 << 7
	EX_DIG         uint32 = 1 << 8
	EX_EATKEY      uint32 = 1 << 9
	EX_NOPASSDOOR  uint32 = 1 << 10
	EX_HIDDEN      uint32 = 1 << 11
	EX_PASSAGE     uint32 = 1 << 12
	EX_PORTAL      uint32 = 1 << 13
	EX_RES1        uint32 = 1 << 14
	EX_RES2        uint32 = 1 << 15
	EX_xCLIMB      uint32 = 1 << 16
	EX_xENTER      uint32 = 1 << 17
	EX_xLEAVE      uint32 = 1 << 18
	EX_xAUTO       uint32 = 1 << 19
	EX_NOFLEE      uint32 = 1 << 20
	EX_xSEARCHABLE uint32 = 1 << 21
	EX_BASHED      uint32 = 1 << 22
	EX_BASHPROOF   uint32 = 1 << 23
	EX_NOMOB       uint32 = 1 << 24
	EX_WINDOW      uint32 = 1 << 25
	EX_xLOOK       uint32 = 1 << 26
	EX_ISBOLT      uint32 = 1 << 27
	EX_BOLTED      uint32 = 1 << 28
	MAX_EXFLAG            = 28
)

// Container flags.
const (
	CONT_CLOSEABLE uint32 = 1 << 0
	CONT_PICKPROOF uint32 = 1 << 1
	CONT_CLOSED    uint32 = 1 << 2
	CONT_LOCKED    uint32 = 1 << 3
	CONT_EATKEY    uint32 = 1 << 4
)

// Reverse apply constant.
const REVERSE_APPLY = 1000

// Pull/push direction type bases.
const (
	PT_WATER = 100
	PT_AIR   = 200
	PT_EARTH = 300
	PT_FIRE  = 400
)

// PCFLAG bits.
const (
	PCFLAG_R1        uint32 = 1 << 0
	PCFLAG_DEADLY    uint32 = 1 << 1
	PCFLAG_UNAUTHED  uint32 = 1 << 2
	PCFLAG_NORECALL  uint32 = 1 << 3
	PCFLAG_NOINTRO   uint32 = 1 << 4
	PCFLAG_GAG       uint32 = 1 << 5
	PCFLAG_RETIRED   uint32 = 1 << 6
	PCFLAG_GUEST     uint32 = 1 << 7
	PCFLAG_NOSUMMON  uint32 = 1 << 8
	PCFLAG_PAGERON   uint32 = 1 << 9
	PCFLAG_NOTITLE   uint32 = 1 << 10
	PCFLAG_GROUPWHO  uint32 = 1 << 11
	PCFLAG_DIAGNOSE  uint32 = 1 << 12
	PCFLAG_HIGHGAG   uint32 = 1 << 13
	PCFLAG_WATCH     uint32 = 1 << 14
	PCFLAG_HELPSTART uint32 = 1 << 15
	PCFLAG_DND       uint32 = 1 << 16
	PCFLAG_IDLE      uint32 = 1 << 17
	PCFLAG_NOBIO     uint32 = 1 << 18
	PCFLAG_NODESC    uint32 = 1 << 19
	PCFLAG_BECKON    uint32 = 1 << 20
	PCFLAG_NOEXP     uint32 = 1 << 21
	PCFLAG_NOBECKON  uint32 = 1 << 22
	PCFLAG_HINTS     uint32 = 1 << 23
	PCFLAG_NOHTTP    uint32 = 1 << 24
	PCFLAG_FREEKILL  uint32 = 1 << 25
)

// Area status flags.
const (
	AREA_DELETED uint32 = 1 << 0
	AREA_LOADED  uint32 = 1 << 1
)

// Area feature flags.
const (
	AFLAG_NOPKILL     uint32 = 1 << 0
	AFLAG_FREEKILL    uint32 = 1 << 1
	AFLAG_NOTELEPORT  uint32 = 1 << 2
	AFLAG_SPELLLIMIT  uint32 = 1 << 3
	AFLAG_SILENCE     uint32 = 1 << 4
	AFLAG_NOSUMMON    uint32 = 1 << 5
	AFLAG_SCRAP       uint32 = 1 << 6
	AFLAG_HIDDEN      uint32 = 1 << 7
	AFLAG_NOWHERE     uint32 = 1 << 8
	AFLAG_NEWBIE      uint32 = 1 << 9
	AFLAG_NOHOVER     uint32 = 1 << 10
	AFLAG_NOLOGOUT    uint32 = 1 << 11
	AFLAG_NOPORTALIN  uint32 = 1 << 12
	AFLAG_NOPORTALOUT uint32 = 1 << 13
	AFLAG_NOASTRAL    uint32 = 1 << 14
	AFLAG_NOMAGIC     uint32 = 1 << 15
)

// Bit reset constants.
const (
	BIT_RESET_DOOR           = 0
	BIT_RESET_OBJECT         = 1
	BIT_RESET_MOBILE         = 2
	BIT_RESET_ROOM           = 3
	BIT_RESET_TYPE_MASK      = 0xFF
	BIT_RESET_DOOR_THRESHOLD = 8
	BIT_RESET_DOOR_MASK      = 0xFF00
	BIT_RESET_SET            = 1 << 30
	BIT_RESET_TOGGLE         = 1 << 31
)

// Skill number type ranges.
const (
	TYPE_UNDEFINED = -1
	TYPE_HIT       = 1000
	TYPE_HERB      = 2000
	TYPE_PERSONAL  = 3000
	TYPE_RACIAL    = 4000
	TYPE_DISEASE   = 5000
)

// Act string flags.
const (
	STRING_NONE uint32 = 0
	STRING_IMM  uint32 = 1 << 1
)

// Auth flags.
const (
	FLAG_WRAUTH = 1
	FLAG_AUTH   = 2
)

// Sector bitvector flags.
const (
	BVSECT_INSIDE       uint32 = 1 << 0
	BVSECT_CITY         uint32 = 1 << 1
	BVSECT_FIELD        uint32 = 1 << 2
	BVSECT_FOREST       uint32 = 1 << 3
	BVSECT_HILLS        uint32 = 1 << 4
	BVSECT_MOUNTAIN     uint32 = 1 << 5
	BVSECT_WATER_SWIM   uint32 = 1 << 6
	BVSECT_WATER_NOSWIM uint32 = 1 << 7
	BVSECT_UNDERWATER   uint32 = 1 << 8
	BVSECT_AIR          uint32 = 1 << 9
	BVSECT_DESERT       uint32 = 1 << 10
	BVSECT_DUNNO        uint32 = 1 << 11
	BVSECT_OCEANFLOOR   uint32 = 1 << 12
	BVSECT_UNDERGROUND  uint32 = 1 << 13
	BVSECT_LAVA         uint32 = 1 << 14
	BVSECT_SWAMP        uint32 = 1 << 15
	MAX_SECFLAG                = 15
)

// Command flags.
const (
	CMD_FLAG_POSSESS     uint32 = 1 << 0
	CMD_FLAG_POLYMORPHED uint32 = 1 << 1
	CMD_WATCH            uint32 = 1 << 2
	CMD_FLAG_RETIRED     uint32 = 1 << 3
	CMD_FLAG_NO_ABORT    uint32 = 1 << 4
)

// Liquids.
const (
	LIQ_WATER = 0
	LIQ_MAX   = 18
)

// Max ignore count.
const MAX_IGN = 6

// Auction history size.
const AUCTION_MEM = 3

// Max climate settings.
const MAX_CLIMATE = 5

// Stance grand master.
const STANCE_GRAND_MASTER = 200

// Mud prog error defines.
const (
	ERROR_PROG   = -1
	IN_FILE_PROG = -2
)

// Spell silent marker.
const SPELL_SILENT_MARKER = "silent"

// Hidden tilde character.
const HIDDEN_TILDE = '*'
