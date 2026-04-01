package net

import "fmt"

// MSDP (Mud Server Data Protocol) constants.
const (
	MSDP_VAR       byte = 1
	MSDP_VAL       byte = 2
	MSDP_TABLE_OPEN  byte = 3
	MSDP_TABLE_CLOSE byte = 4
	MSDP_ARRAY_OPEN  byte = 5
	MSDP_ARRAY_CLOSE byte = 6
)

// MSDPVariable holds one reportable variable.
type MSDPVariable struct {
	Name  string
	Value string
}

// BuildMSDPReport builds an MSDP report for the given variables.
func BuildMSDPReport(vars []MSDPVariable) []byte {
	var data []byte
	for _, v := range vars {
		data = append(data, MSDP_VAR)
		data = append(data, []byte(v.Name)...)
		data = append(data, MSDP_VAL)
		data = append(data, []byte(v.Value)...)
	}
	return TelnetSubneg(TELOPT_MSDP, data)
}

// MSDPCharReport builds standard character status variables.
func MSDPCharReport(name string, hp, maxHP, mana, maxMana, move, maxMove int) []byte {
	return BuildMSDPReport([]MSDPVariable{
		{"CHARACTER_NAME", name},
		{"HEALTH", fmt.Sprintf("%d", hp)},
		{"HEALTH_MAX", fmt.Sprintf("%d", maxHP)},
		{"MANA", fmt.Sprintf("%d", mana)},
		{"MANA_MAX", fmt.Sprintf("%d", maxMana)},
		{"MOVEMENT", fmt.Sprintf("%d", move)},
		{"MOVEMENT_MAX", fmt.Sprintf("%d", maxMove)},
	})
}
