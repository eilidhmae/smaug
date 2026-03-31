package types

// AreaData represents a zone/area in the game world.
// Maps to C struct area_data (mud.h:3274).
type AreaData struct {
	Resets []*ResetData

	Name     string
	Filename string
	Author   string
	Credits  string
	ResetMsg string

	Flags          int
	Status         int
	Age            int
	NPlayer        int
	ResetFrequency int

	// Vnum ranges
	LowRVnum int
	HiRVnum  int
	LowOVnum int
	HiOVnum  int
	LowMVnum int
	HiMVnum  int

	// Level ranges
	LowSoftRange int
	HiSoftRange  int
	LowHardRange int
	HiHardRange  int

	// Spell economy
	SpellLimit     int
	CurrSpellCount int

	MaxPlayers int

	// Statistics
	MKills       int
	MDeaths      int
	PKills       int
	PDeaths      int
	GoldLooted   int
	SilverLooted int
	CopperLooted int
	IllegalPK    int
	HighEconomy  int
	LowEconomy   int

	// Weather
	Weather  *WeatherData
	WeatherX int
	WeatherY int

	// Overland
	Continent int

	// Version tracking for file format
	Version int
}

// ResetData for area resets.
// Maps to C struct reset_data (mud.h:3246).
type ResetData struct {
	Command byte // 'M', 'O', 'P', 'G', 'E', 'H', 'B', 'T', 'D', 'R', 'S'
	Extra   int
	Arg1    int
	Arg2    int
	Arg3    int
}

// WeatherData for area climate.
// Maps to C struct weather_data (mud.h:822).
type WeatherData struct {
	Temp          int
	Precip        int
	Wind          int
	TempVector    int
	PrecipVector  int
	WindVector    int
	ClimateTemp   int
	ClimatePrecip int
	ClimateWind   int
	Neighbors     []*NeighborData
	Echo          string
	EchoColor     int
}

// NeighborData for weather system.
type NeighborData struct {
	Name    string
	Address *AreaData
}
