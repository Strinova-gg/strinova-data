// Package gamedata provides versioned game data loading and damage calculation
// from extracted Strinova game assets.
package gamedata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VersionedData holds all game data for a specific game version.
type VersionedData struct {
	Version string
	Weapons *WeaponsFile
	Agents  *AgentsFile
	Growth  *GrowthFile
}

// WeaponsFile is the top-level JSON structure for weapons.json.
type WeaponsFile struct {
	Version   string                `json:"version"`
	Extracted string                `json:"extracted"`
	Weapons   map[string]WeaponData `json:"weapons"`
}

// WeaponData holds stats for a single weapon.
type WeaponData struct {
	Name               string  `json:"name"`
	WeaponID           string  `json:"weaponId"`
	AttackDamage       float32 `json:"attackDamage,omitempty"`
	AttackRange        float32 `json:"attackRange,omitempty"`
	AttackKeepTime     float32 `json:"attackKeepTime,omitempty"`
	DamageScaleFor2D   float32 `json:"damageScaleFor2D,omitempty"`
	AmmoMax            float32 `json:"ammoMax,omitempty"`
	AmmoPerMagazine    float32 `json:"ammoPerMagazine,omitempty"`
	SpreadModifierBase float32 `json:"spreadModifierBase,omitempty"`
	MobileDamage       float32 `json:"mobileDamage,omitempty"`
}

// AgentsFile is the top-level JSON structure for agents.json.
type AgentsFile struct {
	Version          string                        `json:"version"`
	Extracted        string                        `json:"extracted"`
	Agents           map[string]AgentData          `json:"agents"`
	SecondaryWeapons map[string]SecondaryWeaponData `json:"secondaryWeapons,omitempty"`
}

// AgentData holds profile data for a single agent.
// Secondary weapons are player-selectable (not agent-locked) and live in SecondaryWeapons.
type AgentData struct {
	Name              string  `json:"name"`
	Class             string  `json:"class"`
	PrimaryWeapon     string  `json:"primaryWeapon,omitempty"`
	PrimaryWeaponName string  `json:"primaryWeaponName,omitempty"`
	PrimaryWeaponDamage float32 `json:"primaryWeaponDamage,omitempty"`
}

// SecondaryWeaponData holds data for a shared sidearm that any player can equip.
type SecondaryWeaponData struct {
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	AttackDamage float32 `json:"attackDamage,omitempty"`
}

// GrowthFile is the top-level JSON structure for growth.json.
type GrowthFile struct {
	Version     string                       `json:"version"`
	Extracted   string                       `json:"extracted"`
	Mode        string                       `json:"mode"`
	SharedSlots map[string]map[string]SlotOption `json:"sharedSlots"`
	GunSlots    json.RawMessage              `json:"gunSlots"`
}

// SlotOption represents one upgrade choice for a slot.
type SlotOption struct {
	Name            string  `json:"name"`
	Description     string  `json:"description,omitempty"`
	Cost            int     `json:"cost,omitempty"`
	Armor           int     `json:"armor,omitempty"`
	SpeedBonus      float32 `json:"speedBonus,omitempty"`
	DamageReduction float32 `json:"damageReduction,omitempty"`
}

// LoadVersionedData loads all game data for a specific version from the data directory.
func LoadVersionedData(dataDir string, version string) (*VersionedData, error) {
	versionDir := filepath.Join(dataDir, version)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no game data for version %s at %s", version, versionDir)
	}

	vd := &VersionedData{Version: version}

	if weapons, err := loadWeapons(filepath.Join(versionDir, "weapons.json")); err == nil {
		vd.Weapons = weapons
	}
	if agents, err := loadAgents(filepath.Join(versionDir, "agents.json")); err == nil {
		vd.Agents = agents
	}
	if growth, err := loadGrowth(filepath.Join(versionDir, "growth.json")); err == nil {
		vd.Growth = growth
	}

	return vd, nil
}

// GetWeaponDamage returns the base damage for a weapon ID.
func (vd *VersionedData) GetWeaponDamage(weaponID string) float32 {
	if vd == nil || vd.Weapons == nil {
		return 0
	}
	if w, ok := vd.Weapons.Weapons[weaponID]; ok {
		return w.AttackDamage
	}
	return 0
}

// GetAgentByRoleID returns agent data by role ID.
func (vd *VersionedData) GetAgentByRoleID(roleID int32) *AgentData {
	if vd == nil || vd.Agents == nil {
		return nil
	}
	key := fmt.Sprintf("%d", roleID)
	if a, ok := vd.Agents.Agents[key]; ok {
		return &a
	}
	return nil
}

func loadWeapons(path string) (*WeaponsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wf WeaponsFile
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, err
	}
	return &wf, nil
}

func loadAgents(path string) (*AgentsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var af AgentsFile
	if err := json.Unmarshal(data, &af); err != nil {
		return nil, err
	}
	return &af, nil
}

func loadGrowth(path string) (*GrowthFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gf GrowthFile
	if err := json.Unmarshal(data, &gf); err != nil {
		return nil, err
	}
	return &gf, nil
}
