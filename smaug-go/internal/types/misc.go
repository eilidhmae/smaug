package types

// VariableData for mob/player variables.
// Maps to C struct variable_data (mud.h:674).
type VariableData struct {
	Type    int // vtNONE, vtINT, vtXBIT, vtSTR
	Flags   int
	Vnum    int
	CTime   int64
	MTime   int64
	RTime   int64
	Expires int64
	Timer   int
	Tag     string
	Data    any
}

// HelpData for the help system.
// Maps to C struct help_data (mud.h:1176).
type HelpData struct {
	Level   int
	Keyword string
	Text    string
}

// BoardData for boards.
// Maps to C struct board_data (mud.h:1604).
type BoardData struct {
	Notes          []*NoteData
	NoteFile       string
	ReadGroup      string
	PostGroup      string
	ExtraReaders   string
	ExtraRemovers  string
	ExtraBallots   string
	OTakeMessg     string
	OPostMessg     string
	ORemoveMessg   string
	OCopyMessg     string
	OListMessg     string
	PostMessg      string
	OReadMessg     string
	BoardObj       int
	NumPosts       int
	MinReadLevel   int
	MinPostLevel   int
	MinRemoveLevel int
	MinBallotLevel int
	MaxPosts       int
	Type           int
}

// BanData for site bans.
// Maps to C struct ban_data (mud.h:748).
type BanData struct {
	Name      string
	User      string
	Note      string
	BanBy     string
	BanTime   string
	Flag      int
	UnbanDate int
	Duration  int
	Level     int
	Warn      bool
	Prefix    bool
	Suffix    bool
}

// SocialType for social commands.
type SocialType struct {
	Name        string
	CharNoArg   string
	OthersNoArg string
	CharFound   string
	OthersFound string
	VictFound   string
	CharAuto    string
	OthersAuto  string
}

// CmdType for command table entries.
type CmdType struct {
	Name     string
	DoFun    string // function name, resolved at runtime
	Position int
	Level    int
	Log      int
	Flags    int
}

// AuctionData for the auction system.
type AuctionData struct {
	Item     *ObjData
	Seller   *CharData
	Buyer    *CharData
	Bet      int
	Going    int
	Pulse    int
	Starting int
}

// LiqType for liquid definitions.
type LiqType struct {
	Name   string
	Color  string
	Affect [3]int
}

// StrAppType for strength attribute bonuses.
type StrAppType struct {
	ToHit int
	ToDam int
	Carry int
	Wield int
}

// IntAppType for intelligence attribute bonuses.
type IntAppType struct {
	Learn int
}

// WisAppType for wisdom attribute bonuses.
type WisAppType struct {
	Practice int
}

// DexAppType for dexterity attribute bonuses.
type DexAppType struct {
	Defensive int
}

// ConAppType for constitution attribute bonuses.
type ConAppType struct {
	Hitp  int
	Shock int
}

// ChaAppType for charisma attribute bonuses.
type ChaAppType struct {
	Charm int
}

// LckAppType for luck attribute bonuses.
type LckAppType struct {
	Luck int
}
