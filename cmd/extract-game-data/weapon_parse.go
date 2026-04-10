package main

import (
	"encoding/binary"
	"math"
)

// parseAttackCount extracts the AttackCount ByteProperty (number of pellets per shot).
// Returns 1 for single-bullet weapons, >1 for shotguns.
func parseAttackCount(uexp []byte, names []string, nameIdx map[string]int32) int {
	acIdx, ok := nameIdx["AttackCount"]
	if !ok {
		return 1
	}
	bpIdx, ok := nameIdx["ByteProperty"]
	if !ok {
		return 1
	}

	for off := 0; off+34 < len(uexp); off++ {
		ni := int32(binary.LittleEndian.Uint32(uexp[off : off+4]))
		nn := int32(binary.LittleEndian.Uint32(uexp[off+4 : off+8]))
		ti := int32(binary.LittleEndian.Uint32(uexp[off+8 : off+12]))
		if ni != acIdx || nn != 0 || ti != bpIdx {
			continue
		}
		size := int(binary.LittleEndian.Uint32(uexp[off+16 : off+20]))
		if size == 1 {
			valOff := off + 24 + 8 + 1
			if valOff < len(uexp) {
				val := int(uexp[valOff])
				if val > 0 {
					return val
				}
			}
		}
		break
	}
	return 1
}

// parseBodyDamageMultipliers extracts head/body/leg damage multipliers from the
// BodyDamageFactor MapProperty<EnumProperty, StructProperty> in a weapon blueprint.
// Uses Head for head multiplier, Chest for body, LeftThigh for legs.
func parseBodyDamageMultipliers(uexp []byte, names []string, nameIdx map[string]int32) *BodyMultipliers {
	bdfIdx, ok := nameIdx["BodyDamageFactor"]
	if !ok {
		return nil
	}
	mapPropIdx, hasMap := nameIdx["MapProperty"]
	if !hasMap {
		return nil
	}
	valIdx, hasVal := nameIdx["Value"]
	fpIdx, hasFP := nameIdx["FloatProperty"]
	if !hasVal || !hasFP {
		return nil
	}

	type bodyPartTarget struct {
		nameKey string
		dest    string
	}
	targets := []bodyPartTarget{
		{"EWFBaseHumanBodyPartsType::Head", "head"},
		{"EWFBaseHumanBodyPartsType::Chest", "body"},
		{"EWFBaseHumanBodyPartsType::LeftThigh", "leg"},
	}

	partIndices := make(map[int32]string)
	for _, t := range targets {
		if idx, ok := nameIdx[t.nameKey]; ok {
			partIndices[idx] = t.dest
		}
	}
	if len(partIndices) == 0 {
		return nil
	}

	for off := 0; off+42 < len(uexp); off++ {
		ni := int32(binary.LittleEndian.Uint32(uexp[off : off+4]))
		nn := int32(binary.LittleEndian.Uint32(uexp[off+4 : off+8]))
		ti := int32(binary.LittleEndian.Uint32(uexp[off+8 : off+12]))
		if ni != bdfIdx || nn != 0 || ti != mapPropIdx {
			continue
		}

		size := int(binary.LittleEndian.Uint32(uexp[off+16 : off+20]))
		dataStart := off + 41
		if dataStart+8+size > len(uexp) {
			break
		}

		mapData := uexp[dataStart+8 : dataStart+8+size-8]
		result := &BodyMultipliers{Body: 1.0}
		found := 0

		for partIdx, region := range partIndices {
			for j := 0; j+30 < len(mapData); j++ {
				if int32(binary.LittleEndian.Uint32(mapData[j:j+4])) != partIdx {
					continue
				}
				if int32(binary.LittleEndian.Uint32(mapData[j+4:j+8])) != 0 {
					continue
				}
				for k := j + 8; k+29 < len(mapData) && k < j+200; k++ {
					fni := int32(binary.LittleEndian.Uint32(mapData[k : k+4]))
					fnn := int32(binary.LittleEndian.Uint32(mapData[k+4 : k+8]))
					fti := int32(binary.LittleEndian.Uint32(mapData[k+8 : k+12]))
					if fni == valIdx && fnn == 0 && fti == fpIdx {
						val := math.Float32frombits(binary.LittleEndian.Uint32(mapData[k+25 : k+29]))
						if val > 0.01 && val < 100 {
							switch region {
							case "head":
								result.Head = val
							case "body":
								result.Body = val
							case "leg":
								result.Leg = val
							}
							found++
						}
						break
					}
				}
				break
			}
		}

		if found > 0 {
			return result
		}
		break
	}

	return nil
}

func parseDistanceFalloff(uexp []byte, names []string, nameIdx map[string]int32) *FalloffData {
	result := &FalloffData{}
	found := false

	for _, prop := range []struct {
		name string
		dest *float32
	}{
		{"DistanceProtectedStart", &result.FullDamageRange},
		{"DistanceProtectedEnd", &result.FalloffEndRange},
		{"DistanceEffective", &result.EffectiveRange},
		{"FactorEffective", &result.FactorEffective},
		{"FactorMaximal", &result.FactorMaximal},
	} {
		val, ok := findPropertyValue(uexp, names, nameIdx, prop.name)
		if ok {
			*prop.dest = val
			if prop.name == "DistanceProtectedEnd" || prop.name == "DistanceEffective" {
				found = true
			}
		}
	}

	if !found {
		return nil
	}
	return result
}
