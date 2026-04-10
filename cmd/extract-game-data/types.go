package main

// JSON output types for extracted weapon data.

type WeaponsFile struct {
	Version   string                 `json:"version"`
	Extracted string                 `json:"extracted"`
	Weapons   map[string]*WeaponData `json:"weapons"`
}

type WeaponData struct {
	Name                  string           `json:"name"`
	WeaponID              string           `json:"weaponId"`
	AttackDamage          float32          `json:"attackDamage,omitempty"`
	AttackCount           int              `json:"attackCount,omitempty"`
	DamagePerShot         float32          `json:"damagePerShot,omitempty"`
	ChargeStages          []float32        `json:"chargeStages,omitempty"`
	BodyDamageMultipliers *BodyMultipliers `json:"bodyDamageMultipliers,omitempty"`
	DistanceFalloff       *FalloffData     `json:"distanceFalloff,omitempty"`
	AttackRange           float32          `json:"attackRange,omitempty"`
	AttackKeepTime        float32          `json:"attackKeepTime,omitempty"`
	DamageScaleFor2D      float32          `json:"damageScaleFor2D,omitempty"`
	AmmoMax               float32          `json:"ammoMax,omitempty"`
	AmmoPerMagazine       float32          `json:"ammoPerMagazine,omitempty"`
	SpreadModifierBase    float32          `json:"spreadModifierBase,omitempty"`
	MobileDamage          float32          `json:"mobileDamage,omitempty"`
}

type BodyMultipliers struct {
	Head float32 `json:"head"`
	Body float32 `json:"body"`
	Leg  float32 `json:"leg"`
}

type FalloffData struct {
	FullDamageRange  float32        `json:"fullDamageRange"`
	FalloffEndRange  float32        `json:"falloffEndRange"`
	EffectiveRange   float32        `json:"effectiveRange"`
	FactorEffective  float32        `json:"factorEffective"`
	FactorMaximal    float32        `json:"factorMaximal"`
	DamageAtDistance map[string]int `json:"damageAtDistance,omitempty"`
}
