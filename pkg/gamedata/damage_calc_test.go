package gamedata

import (
	"math"
	"os"
	"testing"
)

func TestComputeShotDamage(t *testing.T) {
	// Only run if game data exists
	dataDir := "../../data/game"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("game data not available")
	}

	vd, err := LoadVersionedData(dataDir, "1.8.2.4")
	if err != nil {
		t.Fatalf("LoadVersionedData: %v", err)
	}

	tests := []struct {
		name     string
		weapon   string
		bodyPart BodyPart
		wantMin  float32
		wantMax  float32
	}{
		{"SCARH body", "10108001", BodyPartBody, 27, 27},
		{"SCARH head", "10108001", BodyPartHead, 38, 42},
		{"SCARH leg", "10108001", BodyPartLeg, 17, 21},
		{"FAMAS body", "10106001", BodyPartBody, 25, 25},
		{"SVD body", "10303001", BodyPartBody, 76, 76},
		{"SVD head", "10303001", BodyPartHead, 240, 300},
		{"DesertEagle body", "12303001", BodyPartBody, 42, 42},
		{"DesertEagle head", "12303001", BodyPartHead, 100, 110},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := DamageInput{
				WeaponID: tc.weapon,
				BodyPart: tc.bodyPart,
			}
			dmg := vd.ComputeShotDamage(input)
			if dmg < tc.wantMin || dmg > tc.wantMax {
				t.Errorf("damage = %.1f, want [%.1f, %.1f]", dmg, tc.wantMin, tc.wantMax)
			}
		})
	}

	// Test with armor reduction
	t.Run("SCARH body with 7B armor", func(t *testing.T) {
		input := DamageInput{
			WeaponID:       "10108001",
			BodyPart:       BodyPartBody,
			ArmorReduction: 0.46,
		}
		dmg := vd.ComputeShotDamage(input)
		expected := float32(27.0 * 0.54)
		if math.Abs(float64(dmg-expected)) > 0.5 {
			t.Errorf("damage with armor = %.1f, want ~%.1f", dmg, expected)
		}
	})
}

func TestGetWeaponDamage(t *testing.T) {
	dataDir := "../../data/game"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("game data not available")
	}

	vd, err := LoadVersionedData(dataDir, "1.8.2.4")
	if err != nil {
		t.Fatalf("LoadVersionedData: %v", err)
	}

	// Verify known weapons
	scarh := vd.GetWeaponDamage("10108001")
	if scarh != 27 {
		t.Errorf("SCARH damage = %.1f, want 27", scarh)
	}

	famas := vd.GetWeaponDamage("10106001")
	if famas != 25 {
		t.Errorf("FAMAS damage = %.1f, want 25", famas)
	}

	// Unknown weapon returns 0
	unknown := vd.GetWeaponDamage("99999999")
	if unknown != 0 {
		t.Errorf("unknown weapon damage = %.1f, want 0", unknown)
	}
}

func TestGetAgentByRoleID(t *testing.T) {
	dataDir := "../../data/game"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Skip("game data not available")
	}

	vd, err := LoadVersionedData(dataDir, "1.8.2.4")
	if err != nil {
		t.Fatalf("LoadVersionedData: %v", err)
	}

	michele := vd.GetAgentByRoleID(101)
	if michele == nil || michele.Name != "Michele" {
		t.Errorf("agent 101 = %v, want Michele", michele)
	}

	fuchsia := vd.GetAgentByRoleID(112)
	if fuchsia == nil || fuchsia.Name != "Fuchsia" {
		t.Errorf("agent 112 = %v, want Fuchsia", fuchsia)
	}
}
