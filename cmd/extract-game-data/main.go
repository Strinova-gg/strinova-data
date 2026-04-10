// Command extract-game-data reads cooked UE4 weapon blueprint assets from the
// Content/ directory and produces versioned JSON files with weapon stats, agent
// profiles, and growth data.
//
// Usage: go run ./cmd/extract-game-data -version 1.8.2.4 -content Content/ -out data/game/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	version := flag.String("version", "", "Game version string (required)")
	contentDir := flag.String("content", "Content/", "Path to extracted Content directory")
	outDir := flag.String("out", "data/game/", "Output directory for versioned JSON")
	flag.Parse()

	if strings.TrimSpace(*version) == "" {
		fmt.Fprintln(os.Stderr, "error: -version is required (e.g. -version 1.8.2.4)")
		flag.Usage()
		os.Exit(1)
	}

	versionDir := filepath.Join(*outDir, *version)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	weapons := extractWeapons(*contentDir)

	overrides := extractCurveTableOverrides(*contentDir)
	merged := 0
	for weaponID, wd := range weapons {
		if wd.AttackDamage <= 1 {
			if override, ok := findDamageOverride(overrides, wd.Name, weaponID); ok {
				wd.AttackDamage = override
				merged++
			}
		}
		if stages := findChargeStages(overrides, weaponID); len(stages) > 0 {
			wd.ChargeStages = stages
		}
	}
	if merged > 0 {
		fmt.Printf("Merged %d CurveTable damage overrides\n", merged)
	}

	for _, wd := range weapons {
		if wd.AttackCount > 1 {
			wd.DamagePerShot = wd.AttackDamage * float32(wd.AttackCount)
		}
	}

	computeDistanceTables(weapons)

	writeJSON(filepath.Join(versionDir, "weapons.json"), WeaponsFile{
		Version:   *version,
		Extracted: time.Now().Format("2006-01-02"),
		Weapons:   weapons,
	})
	fmt.Printf("Extracted %d weapons to %s/weapons.json\n", len(weapons), versionDir)

	generateAgentsJSON(versionDir, weapons, *version)
	generatePerAgentJSONs(versionDir, weapons, *version)
	generateGrowthJSON(versionDir, *version)
}
