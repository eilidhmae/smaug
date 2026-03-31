package types

// MobIndexData is the prototype/template for a non-player character.
// Maps to C struct mob_index_data (mud.h:2615).
type MobIndexData struct {
	// Spec fun name (resolved at runtime)
	SpecFun string

	Shop      *ShopData
	RShop     *RepairData
	MudProgs  []*MProgData
	ProgTypes BitVector

	PlayerName  string
	ShortDescr  string
	LongDescr   string
	Description string

	Vnum   int
	Count  int
	Killed int

	Sex        int
	Level      int
	Act        BitVector
	AffectedBy BitVector
	Alignment  int
	MobThac0   int
	AC         int

	// Hit dice: hitnodice d hitsizedice + hitplus
	HitNoDice   int
	HitSizeDice int
	HitPlus     int

	// Damage dice
	DamNoDice   int
	DamSizeDice int
	DamPlus     int

	NumAttacks int
	Gold       int
	Silver     int
	Copper     int
	Exp        int
	XFlags     int

	Immune      int
	Resistant   int
	Susceptible int

	Attacks  BitVector
	Defenses BitVector

	Speaks   int
	Speaking int

	Position    int
	DefPosition int
	Height      int
	Weight      int
	Race        int
	Class       int
	Hitroll     int
	Damroll     int

	PermStr int
	PermInt int
	PermWis int
	PermDex int
	PermCon int
	PermCha int
	PermLck int

	SavingPoisonDeath int
	SavingWand        int
	SavingParaPetri   int
	SavingBreath      int
	SavingSpellStaff  int

	Stances [MAX_STANCE]int
}
