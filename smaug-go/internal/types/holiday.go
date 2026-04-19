package types

// HolidayData maps to C HOLIDAY_DATA at src/holidays.h:48-56.
//
// Month and Day are 1-indexed to match the C file format. The 0-indexed
// TimeInfoData.Month / Day is translated at creation time — DoSetHoliday
// create stores TimeInfo.Month+1 / TimeInfo.Day+1 (fixing a latent C bug
// where the unshifted 0-indexed value was stored in a 1-indexed slot).
type HolidayData struct {
	Month    int // 1-indexed. Valid: 1..SysData.MonthsPerYear.
	Day      int // 1-indexed. Valid: 1..SysData.DaysPerMonth.
	Name     string
	Announce string
}
