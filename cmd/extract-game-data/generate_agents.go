package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// AgentDef defines an agent's static data.
type AgentDef struct {
	RoleID        int32
	Name          string
	Class         string
	PrimaryWeapon string // weapon ID
}

// Agents that exist in the game with id, role, class, and primary weapon.
var agentRoster = []AgentDef{
	// Duelist
	{132, "Ming", "Duelist", "10102001"},
	{110, "Bai Mo", "Duelist", "10602001"},
	{112, "Fuchsia", "Duelist", "10108001"},
	{115, "Flavia", "Duelist", "10502001"},
	{119, "Eika", "Duelist", "10603001"},
	{122, "Mara", "Duelist", "10503001"},
	{125, "Chiyo", "Duelist", "10305001"},
	{130, "Cielle", "Duelist", "10604001"},
	// Sentinel
	{105, "Audrey", "Sentinel", "10403001"},
	{123, "Leona", "Sentinel", "10404001"},
	{101, "Michele", "Sentinel", "10101001"},
	{108, "Nobunaga", "Sentinel", "10304001"},
	// Support
	{146, "Xinyi", "Support", "10106001"},
	{120, "Fragrans", "Support", "10111001"},
	{124, "Kokona", "Support", "10202001"},
	// Vanguard
	{205, "Galatea", "Vanguard", "10112001"},
	{137, "Kanami", "Vanguard", "10201001"},
	{128, "Lawine", "Vanguard", "10104001"},
	// Controller
	{107, "Maddelena", "Controller", "10301001"},
	{133, "Meredith", "Controller", "10105001"},
	{109, "Reiichi", "Controller", "10303001"},
	{121, "Yugiri", "Controller", "10113001"},
	{131, "Yvette", "Controller", "10501001"},
}

// Known gun upgrade overrides (from GrowthPartAttributeModifier_Bomb DataTable
// analysis + user-confirmed in-game values).
var knownUpgrades = map[int32][]interface{}{
	130: { // Cielle (M1887)
		map[string]interface{}{
			"slot": "part3",
			"optionA": map[string]interface{}{
				"name":        "SingleBulletShootingMode",
				"description": "Switches M1887 to single-shot DMR mode",
				"modifiedWeaponStats": map[string]interface{}{
					"attackDamage":          100,
					"attackCount":           1,
					"damagePerShot":         100,
					"bodyDamageMultipliers": map[string]interface{}{"head": 1.5, "body": 1.0, "leg": 0.7},
					"distanceFalloff": map[string]interface{}{
						"fullDamageRange": 0, "falloffEndRange": 2200,
						"effectiveRange": 6000, "factorEffective": 0.06, "factorMaximal": 0.09,
						"damageAtDistance": map[string]int{"10m": 100, "30m": 99, "50m": 96},
					},
				},
			},
		},
	},
	109: { // Reiichi (SVD)
		map[string]interface{}{
			"slot": "part3",
			"optionA": map[string]interface{}{
				"name":        "BurstMode",
				"description": "Adds 3 extra shots for 4-shot burst",
				"rawModifiers": map[string]interface{}{
					"attackCountAdditive": 3,
				},
			},
		},
	},
}

type SecondaryWeaponDef struct {
	WeaponID string
	Name     string
	Category string
}

var secondaryWeapons = []SecondaryWeaponDef{
	{"12303001", "DesertEagle", "pistol"},
	{"12101001", "MicroUzi", "machine_pistol"},
	{"12201001", "Chiappa", "revolver"},
	{"12307001", "FlamePistol", "special_pistol"},
}

func generateAgentsJSON(versionDir string, weapons map[string]*WeaponData, version string) {
	agents := make(map[string]interface{})
	for _, a := range agentRoster {
		entry := map[string]interface{}{
			"name":  a.Name,
			"class": a.Class,
		}
		if a.PrimaryWeapon != "" {
			entry["primaryWeapon"] = a.PrimaryWeapon
			if w, ok := weapons[a.PrimaryWeapon]; ok {
				entry["primaryWeaponName"] = w.Name
				entry["primaryWeaponDamage"] = w.AttackDamage
			}
		}
		agents[fmt.Sprintf("%d", a.RoleID)] = entry
	}

	secs := make(map[string]interface{})
	for _, sw := range secondaryWeapons {
		sec := map[string]interface{}{
			"name":     sw.Name,
			"category": sw.Category,
		}
		if w, ok := weapons[sw.WeaponID]; ok {
			sec["attackDamage"] = w.AttackDamage
		}
		secs[sw.WeaponID] = sec
	}

	output := map[string]interface{}{
		"version":          version,
		"extracted":        "auto",
		"agents":           agents,
		"secondaryWeapons": secs,
	}

	path := filepath.Join(versionDir, "agents.json")
	writeJSON(path, output)
	fmt.Printf("Generated agents.json with %d agents\n", len(agents))
}

func generatePerAgentJSONs(versionDir string, weapons map[string]*WeaponData, version string) {
	agentsDir := filepath.Join(versionDir, "agents")
	os.MkdirAll(agentsDir, 0o755)

	count := 0
	for _, a := range agentRoster {
		pw := weapons[a.PrimaryWeapon]

		agentData := map[string]interface{}{
			"version": version,
			"roleId":  a.RoleID,
			"name":    a.Name,
			"class":   a.Class,
		}

		if pw != nil {
			weaponInfo := map[string]interface{}{
				"weaponId":     a.PrimaryWeapon,
				"name":         pw.Name,
				"attackDamage": pw.AttackDamage,
				"attackCount":  pw.AttackCount,
			}
			if pw.DamagePerShot > 0 {
				weaponInfo["damagePerShot"] = pw.DamagePerShot
			}
			if pw.ChargeStages != nil {
				weaponInfo["chargeStages"] = pw.ChargeStages
			}
			if pw.BodyDamageMultipliers != nil {
				weaponInfo["bodyDamageMultipliers"] = pw.BodyDamageMultipliers
			}
			if pw.DistanceFalloff != nil {
				weaponInfo["distanceFalloff"] = pw.DistanceFalloff
			}
			agentData["primaryWeapon"] = weaponInfo
		}

		if upgrades, ok := knownUpgrades[a.RoleID]; ok {
			agentData["gunUpgrades"] = upgrades
		}

		filename := strings.ReplaceAll(a.Name, " ", "_")
		filename = strings.ReplaceAll(filename, "—", "-")
		path := filepath.Join(agentsDir, filename+".json")
		writeJSON(path, agentData)
		count++
	}

	fmt.Printf("Generated %d per-agent JSON files in %s/agents/\n", count, versionDir)
}

func generateGrowthJSON(versionDir string, version string) {
	growth := map[string]interface{}{
		"version":   version,
		"extracted": "auto",
		"mode":      "bomb",
		"sharedSlots": map[string]interface{}{
			"6": map[string]interface{}{
				"A": map[string]interface{}{"name": "LightRegenArmor", "description": "Light armor with HP regeneration", "cost": 250},
				"B": map[string]interface{}{"name": "HeavyArmor", "description": "Heavy armor (+50 armor)", "armor": 50, "cost": 250},
			},
			"7": map[string]interface{}{
				"A": map[string]interface{}{"name": "StringifiedMoveSPD", "description": "+15% movement speed", "speedBonus": 0.15, "cost": 250},
				"B": map[string]interface{}{"name": "StringifiedDMGReduction", "description": "46% damage reduction when stringified", "damageReduction": 0.46, "cost": 250},
			},
		},
		"gunSlots": map[string]interface{}{
			"description": "Slots 0-5 are per-agent gun/ability upgrades (90 credits each)",
			"cost":        90,
		},
	}

	path := filepath.Join(versionDir, "growth.json")
	writeJSON(path, growth)
	fmt.Printf("Generated growth.json\n")
}

// computeFalloffDamage calculates damage at a distance using the game's quadratic formula.
func computeFalloffDamage(baseDmg float32, pellets int, distM int, protEnd, effective, factorEff, factorMax float32) int {
	dCM := float64(distM * 100)
	base := float64(baseDmg)
	p := float64(pellets)
	if p < 1 {
		p = 1
	}

	var mult float64
	switch {
	case protEnd >= effective || effective <= 0:
		mult = 1.0
	case dCM <= float64(protEnd):
		mult = 1.0
	case dCM > float64(effective):
		mult = 1.0 - float64(factorMax)
	default:
		frac := (dCM - float64(protEnd)) / (float64(effective) - float64(protEnd))
		mult = 1.0 - frac*frac*float64(factorEff)
	}

	return int(math.Floor(base * mult * p))
}
