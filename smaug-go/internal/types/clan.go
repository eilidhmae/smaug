package types

// ClanData represents a clan/guild/order.
// Maps to C struct clan_data (mud.h:1469).
type ClanData struct {
	Filename    string
	Name        string
	Abbrev      string
	Motto       string
	Description string
	Deity       string
	Leader      string
	Number1     string
	Number2     string
	Badge       string
	LeadRank    string
	OneRank     string
	TwoRank     string

	PKills    [7]int
	PDeaths   [7]int
	MKills    int
	MDeaths   int
	IllegalPK int
	Score     int

	ClanType  int
	Favour    int
	Strikes   int
	Members   int
	MemLimit  int
	Alignment int

	Board     int
	ClanObj1  int
	ClanObj2  int
	ClanObj3  int
	ClanObj4  int
	ClanObj5  int
	Recall    int
	Storeroom int
	Guard1    int
	Guard2    int
	Class     int
}

// CouncilData represents a council.
// Maps to C struct council_data (mud.h:1511).
type CouncilData struct {
	Filename    string
	Name        string
	Description string
	Head        string
	Head2       string
	Powers      string
	Abbrev      string
	Members     int
	Board       int
	Meeting     int
	Storeroom   int
}

// DeityData represents a deity.
// Maps to C struct deity_data (mud.h:1528).
type DeityData struct {
	Filename    string
	Name        string
	Description string

	Alignment   int
	Worshippers int
	SCorpse     int
	SDeityObj   int
	SAvatar     int
	SRecall     int

	Flee        int
	FleeNpcRace int
	FleeNpcFoe  int
	Kill        int
	KillMagic   int
	KillNpcRace int
	KillNpcFoe  int
	Sac         int
	BuryCorpse  int
	AidSpell    int
	Aid         int
	Backstab    int
	Steal       int
	Die         int
	DieNpcRace  int
	DieNpcFoe   int
	SpellAid    int
	DigCorpse   int

	Race        int
	Race2       int
	Class       int
	Sex         int
	NpcRace     int
	NpcFoe      int
	Suscept     int
	Element     int
	Affected    BitVector
	SusceptNum  int
	ElementNum  int
	AffectedNum int
	ObjStat     int
}
