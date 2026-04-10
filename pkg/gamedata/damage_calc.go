package gamedata

import "fmt"

// BodyPart represents a hit body region.
type BodyPart int

const (
	BodyPartUnknown BodyPart = iota
	BodyPartHead
	BodyPartBody
	BodyPartLeg
)

// Default body-part multipliers derived from empirical replay data cross-checked
// against game behavior. These apply when weapon-specific multipliers are not available.
//
// Empirical observations from 7 replays:
//   - Rifles (FAMAS, AKM_Galatea, AUG): head ~1.5x, leg ~0.7x
//   - DMRs (G28, SVD): head ~3.5x, leg ~0.5x
//   - Pistols (DesertEagle): head ~2.5x, leg ~0.9x
//   - Snipers (M82A1, M200): head ~2.1-3.5x, leg ~0.5-0.7x
var defaultMultipliers = map[string]map[BodyPart]float32{
	"assault_rifle": {BodyPartHead: 1.5, BodyPartBody: 1.0, BodyPartLeg: 0.7},
	"dmr":           {BodyPartHead: 3.5, BodyPartBody: 1.0, BodyPartLeg: 0.55},
	"sniper":        {BodyPartHead: 2.5, BodyPartBody: 1.0, BodyPartLeg: 0.6},
	"smg":           {BodyPartHead: 1.4, BodyPartBody: 1.0, BodyPartLeg: 0.75},
	"shotgun":       {BodyPartHead: 1.0, BodyPartBody: 1.0, BodyPartLeg: 1.0},
	"lmg":           {BodyPartHead: 1.75, BodyPartBody: 1.0, BodyPartLeg: 0.55},
	"pistol":        {BodyPartHead: 2.5, BodyPartBody: 1.0, BodyPartLeg: 0.85},
	"default":       {BodyPartHead: 1.5, BodyPartBody: 1.0, BodyPartLeg: 0.75},
}

// weaponCategory maps weapon IDs to damage categories for multiplier lookup.
var weaponCategory = map[string]string{
	"10101001": "assault_rifle", "10102001": "assault_rifle", "10104001": "assault_rifle",
	"10105001": "assault_rifle", "10106001": "assault_rifle", "10108001": "assault_rifle",
	"10111001": "assault_rifle", "10112001": "assault_rifle", "10113001": "assault_rifle",
	"10201001": "sniper", "10202001": "sniper",
	"10301001": "dmr", "10303001": "dmr", "10304001": "dmr", "10305001": "dmr",
	"10403001": "lmg", "10404001": "lmg",
	"10501001": "smg", "10502001": "smg", "10503001": "smg",
	"10602001": "shotgun", "10603001": "shotgun", "10604001": "shotgun",
	"12101001": "smg", "12201001": "pistol", "12303001": "pistol",
}

// DamageInput holds all parameters needed to compute damage for a single shot.
type DamageInput struct {
	WeaponID        string
	BodyPart        BodyPart
	Distance        float32
	ArmorReduction  float32 // 0.0 = no armor, 0.46 = stringified 7B
	UpgradeModifier float32 // multiplier from gun upgrades (1.0 = no upgrade)
}

// ComputeShotDamage calculates the damage for a single shot using game data.
// Returns 0 if the weapon is not found in the data.
func (vd *VersionedData) ComputeShotDamage(input DamageInput) float32 {
	baseDamage := vd.GetWeaponDamage(input.WeaponID)
	if baseDamage <= 0 {
		return 0
	}

	bodyMul := getBodyPartMultiplier(input.WeaponID, input.BodyPart)

	upgradeMul := input.UpgradeModifier
	if upgradeMul <= 0 {
		upgradeMul = 1.0
	}

	armorFactor := float32(1.0)
	if input.ArmorReduction > 0 {
		armorFactor = 1.0 - input.ArmorReduction
	}

	return baseDamage * bodyMul * upgradeMul * armorFactor
}

// GetWeaponCategory returns the damage category for a weapon ID.
func GetWeaponCategory(weaponID string) string {
	if cat, ok := weaponCategory[weaponID]; ok {
		return cat
	}
	return "default"
}

func getBodyPartMultiplier(weaponID string, part BodyPart) float32 {
	cat := GetWeaponCategory(weaponID)
	mults, ok := defaultMultipliers[cat]
	if !ok {
		mults = defaultMultipliers["default"]
	}
	if mul, ok := mults[part]; ok {
		return mul
	}
	return 1.0
}

// WeaponSummary returns a human-readable summary of a weapon's stats.
func (vd *VersionedData) WeaponSummary(weaponID string) string {
	if vd == nil || vd.Weapons == nil {
		return ""
	}
	w, ok := vd.Weapons.Weapons[weaponID]
	if !ok {
		return ""
	}
	cat := GetWeaponCategory(weaponID)
	return fmt.Sprintf("%s (%s) dmg=%.0f ammo=%.0f/%.0f",
		w.Name, cat, w.AttackDamage, w.AmmoPerMagazine, w.AmmoMax)
}
