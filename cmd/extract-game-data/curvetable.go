package main

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// curveTableRow holds a parsed row from CT_WeaponAttribute.
type curveTableRow struct {
	Name string
	Keys []curveKey
}

type curveKey struct {
	Time  float32
	Value float32
}

// extractCurveTableOverrides reads CT_WeaponAttribute and returns all
// Attribute.AttackDamage.* rows with their curve values.
func extractCurveTableOverrides(contentDir string) []curveTableRow {
	uassetPath := filepath.Join(contentDir, "CT_WeaponAttribute.uasset")
	uexpPath := filepath.Join(contentDir, "CT_WeaponAttribute.uexp")

	uasset, err := os.ReadFile(uassetPath)
	if err != nil {
		return nil
	}
	uexp, err := os.ReadFile(uexpPath)
	if err != nil {
		return nil
	}

	names := parseNameTable(uasset)
	if len(names) == 0 {
		return nil
	}

	if len(uexp) < 17 {
		return nil
	}
	numRows := int(binary.LittleEndian.Uint32(uexp[12:16]))

	if numRows <= 0 || numRows > 10000 {
		return nil
	}

	const rowSize = 310
	if 17+numRows*rowSize > len(uexp) {
		return nil
	}

	keysIdx := int32(-1)
	arrPropIdx := int32(-1)
	for i, n := range names {
		switch n {
		case "Keys":
			keysIdx = int32(i)
		case "ArrayProperty":
			arrPropIdx = int32(i)
		}
	}

	var results []curveTableRow
	for row := 0; row < numRows; row++ {
		rowOff := 17 + row*rowSize
		nameIdx := int(binary.LittleEndian.Uint32(uexp[rowOff : rowOff+4]))
		if nameIdx < 0 || nameIdx >= len(names) {
			continue
		}
		rowName := names[nameIdx]
		if !strings.Contains(rowName, "AttackDamage") {
			continue
		}

		keys := parseCurveRowKeys(uexp[rowOff:rowOff+rowSize], keysIdx, arrPropIdx)
		if len(keys) > 0 {
			results = append(results, curveTableRow{Name: rowName, Keys: keys})
		}
	}

	return results
}

// parseCurveRowKeys extracts SimpleCurveKey pairs from a 310-byte CurveTable row.
// The row structure: FName(8) + tagged properties (InterpMode, Keys array, DefaultValue, etc.) + None(8)
// The Keys array uses a 49-byte inner FPropertyTag for the struct elements.
func parseCurveRowKeys(rowData []byte, keysIdx, arrPropIdx int32) []curveKey {
	if keysIdx < 0 || arrPropIdx < 0 {
		return nil
	}

	for probe := 8; probe < 200; probe++ {
		if probe+20 > len(rowData) {
			break
		}
		ni := int32(binary.LittleEndian.Uint32(rowData[probe : probe+4]))
		nn := int32(binary.LittleEndian.Uint32(rowData[probe+4 : probe+8]))
		if ni != keysIdx || nn != 0 {
			continue
		}
		ti := int32(binary.LittleEndian.Uint32(rowData[probe+8 : probe+12]))
		if ti != arrPropIdx {
			continue
		}

		dataOff := probe + 33
		if dataOff+4 > len(rowData) {
			break
		}
		count := int(binary.LittleEndian.Uint32(rowData[dataOff : dataOff+4]))
		if count <= 0 || count > 100 {
			break
		}

		const innerTagSize = 49
		keyStart := dataOff + 4 + innerTagSize
		if keyStart+count*8 > len(rowData) {
			break
		}

		var keys []curveKey
		for k := 0; k < count; k++ {
			off := keyStart + k*8
			t := math.Float32frombits(binary.LittleEndian.Uint32(rowData[off : off+4]))
			v := math.Float32frombits(binary.LittleEndian.Uint32(rowData[off+4 : off+8]))
			keys = append(keys, curveKey{Time: t, Value: v})
		}
		return keys
	}

	return nil
}

// findDamageOverride searches CurveTable rows for a damage value matching the weapon.
func findDamageOverride(rows []curveTableRow, weaponName, weaponID string) (float32, bool) {
	upperName := strings.ToUpper(weaponName)

	weaponCTPatterns := map[string][]string{
		"10102001": {"Auto.AKM"},
		"10502001": {"Auto.VECTOR"},
		"10111001": {"Auto.M4A1_Fragrans"},
		"10604001": {"Single.M1887"},
		"10201001": {"Stage1.M82A1"},
		"10202001": {"Stage1.M200A"},
	}

	patterns, hasPattern := weaponCTPatterns[weaponID]
	if !hasPattern {
		for _, row := range rows {
			if strings.Contains(strings.ToUpper(row.Name), upperName) &&
				strings.Contains(row.Name, "AttackDamage") {
				if val := curveYValue(row.Keys); val > 1 {
					return val, true
				}
			}
		}
		return 0, false
	}

	for _, pattern := range patterns {
		for _, row := range rows {
			if strings.Contains(row.Name, pattern) {
				if val := curveYValue(row.Keys); val > 1 {
					return val, true
				}
			}
		}
	}
	return 0, false
}

// curveYValue returns the Y value from a flat curve (all keys have the same Y).
func curveYValue(keys []curveKey) float32 {
	if len(keys) == 0 {
		return 0
	}
	return keys[0].Value
}

// findChargeStages extracts Stage1-4 damage values for charge weapons (snipers).
func findChargeStages(rows []curveTableRow, weaponID string) []float32 {
	stagePatterns := map[string][]string{
		"10201001": {"Stage1.M82A1", "Stage2.M82A1", "Stage3.M82A1", "Stage4.M82A1"},
		"10202001": {"Stage1.M200A", "Stage2.M200A", "Stage3.M200A", "Stage4.M200A"},
	}

	patterns, ok := stagePatterns[weaponID]
	if !ok {
		return nil
	}

	var stages []float32
	for _, pattern := range patterns {
		found := false
		for _, row := range rows {
			if strings.Contains(row.Name, pattern) {
				val := curveYValue(row.Keys)
				if val > 0 {
					stages = append(stages, val)
					found = true
					break
				}
			}
		}
		if !found {
			break
		}
	}

	if len(stages) < 2 {
		return nil
	}
	return stages
}
