package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

func extractWeapons(contentDir string) map[string]*WeaponData {
	weaponsDir := filepath.Join(contentDir, "CyWeapons")
	results := make(map[string]*WeaponData)

	entries, err := os.ReadDir(weaponsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading CyWeapons dir: %v\n", err)
		return results
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(weaponsDir, entry.Name())
		files, _ := os.ReadDir(dirPath)
		for _, f := range files {
			if !strings.HasPrefix(f.Name(), "Weapon") || !strings.HasSuffix(f.Name(), ".uasset") {
				continue
			}
			if strings.HasSuffix(f.Name(), "_C.uasset") {
				continue
			}

			weaponID := strings.TrimPrefix(f.Name(), "Weapon")
			weaponID = strings.TrimSuffix(weaponID, ".uasset")

			uassetPath := filepath.Join(dirPath, f.Name())
			uexpPath := strings.TrimSuffix(uassetPath, ".uasset") + ".uexp"

			wd := parseWeaponBlueprint(uassetPath, uexpPath, entry.Name(), weaponID)
			if wd != nil {
				results[weaponID] = wd
			}
		}
	}

	return results
}

func parseWeaponBlueprint(uassetPath, uexpPath, dirName, weaponID string) *WeaponData {
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

	nameIdx := make(map[string]int32)
	for i, n := range names {
		nameIdx[n] = int32(i)
	}

	wd := &WeaponData{
		Name:     dirName,
		WeaponID: weaponID,
	}

	targets := map[string]*float32{
		"AttackDamage":       &wd.AttackDamage,
		"AttackRange":        &wd.AttackRange,
		"AttackKeepTime":     &wd.AttackKeepTime,
		"DamageScaleFor2D":   &wd.DamageScaleFor2D,
		"AmmoMax":            &wd.AmmoMax,
		"AmmoPerMagazine":    &wd.AmmoPerMagazine,
		"SpreadModifierBase": &wd.SpreadModifierBase,
	}

	for propName, dest := range targets {
		val, ok := findPropertyValue(uexp, names, nameIdx, propName)
		if ok {
			*dest = val
		}
	}

	mv, ok := findScalableFloatField(uexp, names, nameIdx, "AttackDamage", "MobileValue")
	if ok {
		wd.MobileDamage = mv
	}

	wd.AttackCount = parseAttackCount(uexp, names, nameIdx)
	if wd.AttackCount > 1 {
		wd.DamagePerShot = wd.AttackDamage * float32(wd.AttackCount)
	}

	wd.BodyDamageMultipliers = parseBodyDamageMultipliers(uexp, names, nameIdx)
	wd.DistanceFalloff = parseDistanceFalloff(uexp, names, nameIdx)

	if wd.AttackDamage == 0 && wd.AmmoMax == 0 && wd.AttackRange == 0 {
		return nil
	}

	return wd
}

func parseNameTable(uasset []byte) []string {
	if len(uasset) < 24 {
		return nil
	}

	magic := binary.LittleEndian.Uint32(uasset[0:4])
	if magic != 0x9E2A83C1 {
		return nil
	}

	off := 20
	numCustom := int(binary.LittleEndian.Uint32(uasset[off : off+4]))
	off += 4 + numCustom*20

	off += 4 // TotalHeaderSize

	folderLen := int(int32(binary.LittleEndian.Uint32(uasset[off : off+4])))
	off += 4
	if folderLen > 0 {
		off += folderLen
	} else if folderLen < 0 {
		off += (-folderLen) * 2
	}

	off += 4 // PackageFlags

	nameCount := int(binary.LittleEndian.Uint32(uasset[off : off+4]))
	off += 4
	nameOffset := int(binary.LittleEndian.Uint32(uasset[off : off+4]))

	names := make([]string, 0, nameCount)
	pos := nameOffset
	for i := 0; i < nameCount; i++ {
		if pos+4 > len(uasset) {
			break
		}
		slen := int(int32(binary.LittleEndian.Uint32(uasset[pos : pos+4])))
		pos += 4

		var name string
		if slen > 0 && slen < 10000 && pos+slen <= len(uasset) {
			name = string(uasset[pos : pos+slen-1])
			pos += slen
		} else if slen < 0 && slen > -10000 && pos+(-slen)*2 <= len(uasset) {
			pos += (-slen) * 2
			name = "(utf16)"
		}

		pos += 4 // hash
		names = append(names, name)
	}

	return names
}

func findPropertyValue(uexp []byte, names []string, nameIdx map[string]int32, propName string) (float32, bool) {
	targetIdx, ok := nameIdx[propName]
	if !ok {
		return 0, false
	}
	floatIdx, hasFloat := nameIdx["FloatProperty"]
	intIdx, hasInt := nameIdx["IntProperty"]
	structIdx, hasStruct := nameIdx["StructProperty"]

	for off := 0; off+28 < len(uexp); off++ {
		ni := int32(binary.LittleEndian.Uint32(uexp[off : off+4]))
		nn := int32(binary.LittleEndian.Uint32(uexp[off+4 : off+8]))
		if ni != targetIdx || nn != 0 {
			continue
		}

		ti := int32(binary.LittleEndian.Uint32(uexp[off+8 : off+12]))

		if hasFloat && ti == floatIdx {
			if off+29 < len(uexp) {
				val := math.Float32frombits(binary.LittleEndian.Uint32(uexp[off+25 : off+29]))
				return val, true
			}
		}

		if hasInt && ti == intIdx {
			if off+29 < len(uexp) {
				val := int32(binary.LittleEndian.Uint32(uexp[off+25 : off+29]))
				return float32(val), true
			}
		}

		if hasStruct && ti == structIdx {
			return findScalableFloatField(uexp, names, nameIdx, propName, "Value")
		}
	}

	return 0, false
}

func findScalableFloatField(uexp []byte, names []string, nameIdx map[string]int32, propName string, fieldName string) (float32, bool) {
	targetIdx, ok := nameIdx[propName]
	if !ok {
		return 0, false
	}
	structIdx, hasStruct := nameIdx["StructProperty"]
	fieldIdx, hasField := nameIdx[fieldName]
	floatIdx, hasFloat := nameIdx["FloatProperty"]

	if !hasStruct || !hasField || !hasFloat {
		return 0, false
	}

	for off := 0; off+49 < len(uexp); off++ {
		ni := int32(binary.LittleEndian.Uint32(uexp[off : off+4]))
		nn := int32(binary.LittleEndian.Uint32(uexp[off+4 : off+8]))
		ti := int32(binary.LittleEndian.Uint32(uexp[off+8 : off+12]))

		if ni != targetIdx || nn != 0 || ti != structIdx {
			continue
		}

		size := int(binary.LittleEndian.Uint32(uexp[off+16 : off+20]))
		dataStart := off + 49
		if dataStart+size > len(uexp) {
			continue
		}

		for sp := 0; sp+29 < size; sp++ {
			fni := int32(binary.LittleEndian.Uint32(uexp[dataStart+sp : dataStart+sp+4]))
			fnn := int32(binary.LittleEndian.Uint32(uexp[dataStart+sp+4 : dataStart+sp+8]))
			fti := int32(binary.LittleEndian.Uint32(uexp[dataStart+sp+8 : dataStart+sp+12]))

			if fni == fieldIdx && fnn == 0 && fti == floatIdx {
				val := math.Float32frombits(binary.LittleEndian.Uint32(uexp[dataStart+sp+25 : dataStart+sp+29]))
				return val, true
			}
		}
		break
	}

	return 0, false
}
