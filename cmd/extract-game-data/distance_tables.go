package main

import (
	"fmt"
	"math"
)

// computeDistanceTables generates damage-at-distance tables using the game's
// quadratic falloff formula, calibrated against in-game weapon stats UI:
//
//	Within protected zone (d <= protectedEnd):  damage = baseDamage
//	Falloff zone (protectedEnd < d <= effective): damage = floor(base * (1 - frac² * FactorEffective))
//	Beyond effective (d > effective):            damage = floor(base * (1 - FactorMaximal))
//
// Examples:
//
//	M82A1:  10m=105, 30m=103, 50m=95
//	M200:   10m=100, 30m=99,  50m=96
//	Vector: 10m=21,  30m=14,  50m=14
func computeDistanceTables(weapons map[string]*WeaponData) {
	shotgunIDs := map[string]bool{
		"10602001": true, "10603001": true, "10604001": true,
	}

	for weaponID, wd := range weapons {
		if wd.AttackDamage <= 1 || wd.DistanceFalloff == nil {
			continue
		}
		fo := wd.DistanceFalloff
		if fo.EffectiveRange <= 0 {
			continue
		}

		var distancesM []int
		if shotgunIDs[weaponID] {
			distancesM = []int{10, 20, 30}
		} else {
			distancesM = []int{10, 30, 50}
		}

		table := make(map[string]int)
		protEnd := fo.FalloffEndRange
		effective := fo.EffectiveRange

		// Hack to improve accuracy of the table due to floating point precision issues.
		fe64 := math.Round(float64(fo.FactorEffective)*10000) / 10000
		fm64 := math.Round(float64(fo.FactorMaximal)*10000) / 10000
		perPellet := float64(wd.AttackDamage)
		pellets := wd.AttackCount
		if pellets < 1 {
			pellets = 1
		}

		for _, dm := range distancesM {
			dCM := float64(dm * 100)
			var mult float64

			switch {
			case protEnd >= effective || effective <= 0:
				mult = 1.0
			case dCM <= float64(protEnd):
				mult = 1.0
			case dCM > float64(effective):
				mult = 1.0 - fm64
			default:
				fraction := (dCM - float64(protEnd)) / (float64(effective) - float64(protEnd))
				mult = 1.0 - fraction*fraction*fe64
			}

			dmg := int(math.Floor(perPellet * mult * float64(pellets)))
			key := fmt.Sprintf("%dm", dm)
			table[key] = dmg
		}

		fo.DamageAtDistance = table
	}
}
