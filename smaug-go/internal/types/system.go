package types

// SystemData holds game-wide settings and statistics.
// Maps to C struct system_data (mud.h:3353).
type SystemData struct {
	MaxPlayers int
	AllTimeMax int
	TimeOfMax  string
	MudName    string
	PortName   string

	// Global economy
	GlobalGoldLooted   int
	GlobalSilverLooted int
	GlobalCopperLooted int

	// Used consumables
	UPillVal    int
	UPotionVal  int
	BrewedUsed  int
	ScribedUsed int

	// Restrictions
	NoNameResolving bool
	DenyNewPlayers  bool
	WaitForAuth     bool

	// Mail levels
	ReadAllMail    int
	ReadMailFree   int
	WriteMailFree  int
	TakeOthersMail int

	// Channel levels
	MuseLevel  int
	ThinkLevel int
	BuildLevel int
	LogLevel   int

	// Permission levels
	LevelModifyProto     int
	LevelOverridePrivate int
	LevelMsetPlayer      int

	// Combat modifiers
	BashPlrVsPlr  int
	BashNonTank   int
	GougePlrVsPlr int
	GougeNonTank  int
	StunPlrVsPlr  int
	StunRegular   int
	DodgeMod      int
	ParryMod      int
	TumbleMod     int
	TumblePK      int
	DamNonavVsMob int
	DamMobVsNonav int
	DamPlrVsPlr   int
	DamPlrVsMob   int
	DamMobVsPlr   int
	DamMobVsMob   int

	// Misc levels
	LevelGetObjNoTake int
	LevelForcePC      int
	BestowDif         int

	MaxSN          int
	PeacefulExpMod int
	DeadlyExpMod   int
	GuildOverseer  string
	GuildAdvisor   string
	SaveFlags      int
	SaveFrequency  int
	CheckImmHost   int
	MorphOpt       int
	SavePets       int
	PKChannels     int
	PKSilence      int
	BanSiteLevel   int
	BanClassLevel  int
	BanRaceLevel   int
	IdentRetries   int
	PKLoot         int
	NewsHTMLPath   string
	MaxHTMLNews    int
	SaveVersion    int
	Wizlock        bool
	MagicHell      bool

	// Hotboot
	HotbootInProgress bool

	// Timezone overrides
	SecPerTick     int
	PulsePerSec    int
	PulseTick      int
	PulseViolence  int
	PulseMobile    int
	PulseCalendar  int
	HoursPerDay    int
	DaysPerWeek    int
	DaysPerMonth   int
	DaysPerYear    int
	MonthsPerYear  int
	HourSunrise    int
	HourDayBegin   int
	HourNoon       int
	HourSunset     int
	HourNightBegin int
	HourMidnight   int
	MaxHoliday     int
}

// TimeInfoData for the game calendar.
// Maps to C struct time_info_data (mud.h:799).
type TimeInfoData struct {
	Hour     int
	Day      int
	Month    int
	Year     int
	Season   int
	Sunlight int
}

// WeatherInfoData for global weather state.
// Maps to C struct weather_data (global, not per-area).
type WeatherInfoData struct {
	Mmhg   int // barometric pressure
	Change int // pressure change
	Sky    int // SKY_CLOUDLESS, etc.
	Temp   int
}

// ClassType for class definitions.
// Maps to C struct class_type (mud.h:1368).
type ClassType struct {
	WhoName        string
	Login          string
	LoginOther     string
	Logout         string
	LogoutOther    string
	Reconnect      string
	ReconnectOther string
	Affected       BitVector
	AttrPrime      int
	AttrSecond     int
	AttrDeficient  int
	Resist         int
	Suscept        int
	Weapon         int
	Guild          int
	SkillAdept     int
	Thac0_00       int
	Thac0_32       int
	HPMin          int
	HPMax          int
	FMana          bool
	ExpBase        int
}

// RaceData for race definitions.
// Maps to C struct race_type (mud.h:1395).
type RaceData struct {
	Name              string
	Affected          BitVector
	StrPlus           int
	DexPlus           int
	WisPlus           int
	IntPlus           int
	ConPlus           int
	ChaPlus           int
	LckPlus           int
	Hit               int
	Mana              int
	Resist            int
	Suscept           int
	ClassRestriction  int
	Language          int
	ACPlus            int
	Alignment         int
	Attacks           BitVector
	Defenses          BitVector
	MinAlign          int
	MaxAlign          int
	ExpMultiplier     int
	Height            int
	Weight            int
	HungerMod         int
	ThirstMod         int
	SavingPoisonDeath int
	SavingWand        int
	SavingParaPetri   int
	SavingBreath      int
	SavingSpellStaff  int
	WhereName         [MAX_WHERE_NAME]string
	ManaRegen         int
	HPRegen           int
	RaceRecall        int
}
