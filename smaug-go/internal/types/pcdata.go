package types

// PCData holds player-only data. Maps to C struct pc_data (mud.h:2914).
type PCData struct {
	Pet     *CharData
	Clan    *ClanData
	Council *CouncilData
	Area    *AreaData
	Deity   *DeityData

	// Strings
	Lang        string
	Homepage    string
	Email       string
	ICQ         string
	ClanName    string
	CouncilName string
	DeityName   string
	Pwd         string
	BamfIn      string
	BamfOut     string
	Filename    string
	Rank        string
	Title       string
	Bestowments string
	BettedOn    string
	Prompt      string
	FPrompt     string
	SubPrompt   string
	Bio         string
	AuthedBy    string
	HelledBy    string
	SeeMe       string
	RecentSite  string
	PrevSite    string

	// Flags
	Flags int

	// Stats
	PKills    int
	PDeaths   int
	MKills    int
	MDeaths   int
	IllegalPK int
	BetAmt    int
	Honour    int

	// Times
	OutcastTime int64
	RestoreTime int64
	ReleaseDate int64
	// Hell is the Unix timestamp at which a player was sent to hell.
	// Zero means not in hell. Maps to C pcdata->release_date/hell combo;
	// the Go port keeps it as a simple int64 timestamp for "hell" command.
	Hell int64

	// Building ranges
	RRangeLo int
	RRangeHi int
	MRangeLo int
	MRangeHi int
	ORangeLo int
	ORangeHi int

	// Imm settings
	WizInvis int
	MinSnoop int

	// Conditions
	Condition [MAX_CONDS]int

	// Skills (proficiency 0-100 per skill slot)
	Learned [MAX_SKILL]int

	// Kill tracking
	Killed [MAX_KILLTRACK]KilledData

	// Quest
	QuestNumber int
	QuestCurr   int
	QuestAccum  int

	// Deity
	Favor    int
	Charmies int

	// Auth
	AuthState int

	// Nuisance
	Nuisance *NuisanceData

	// Pager
	PagerLen int

	// Stances
	Stances [MAX_STANCE]int

	// Tourney
	OpenedTourney bool

	// Ignored players
	Ignored []string

	// Tell history (imm only)
	TellHistory []string
	LTIndex     int

	// Color settings
	Colorize []int

	// Aliases
	Aliases []*AliasData

	// Hotboot
	Hotboot bool

	// Timezone
	Timezone int
	Day      int
	Month    int

	// Bank
	GBalance int
	SBalance int
	CBalance int

	// Marriage
	Spouse string

	// Arena
	AKills  int
	ADeaths int

	// Overland
	SecEdit int
}

// KilledData tracks mob kills. Maps to C struct killed_data.
type KilledData struct {
	Vnum  int
	Count int
}

// NuisanceData for nuisance tracking. Maps to C struct nuisance_data.
type NuisanceData struct {
	Time    int64
	MaxTime int64
	Flags   int
	Power   int
}

// AliasData for player aliases.
type AliasData struct {
	Name string
	Cmd  string
}

// NoteData for board notes. Maps to C struct note_data.
type NoteData struct {
	Sender       string
	Date         string
	ToList       string
	Subject      string
	Text         string
	Voting       int
	YesVotes     string
	NoVotes      string
	Abstentions  string
	YesTally     int
	NoTally      int
	AbstainTally int
	NoRemove     int
}
