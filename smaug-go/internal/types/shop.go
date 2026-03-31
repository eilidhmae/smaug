package types

// ShopData for NPC shops.
// Maps to C struct shop_data (mud.h:1268).
type ShopData struct {
	Keeper     int
	BuyType    [MAX_TRADE]int
	ProfitBuy  int
	ProfitSell int
	OpenHour   int
	CloseHour  int
}

// RepairData for NPC repair shops.
// Maps to C struct repairshop_data (mud.h:1285).
type RepairData struct {
	Keeper    int
	FixType   [MAX_FIX]int
	ProfitFix int
	ShopType  int
	OpenHour  int
	CloseHour int
}
