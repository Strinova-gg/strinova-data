# Strinova Game Data

Extracted and versioned game data for [Strinova](https://store.steampowered.com/app/1289400/Strinova/), parsed directly from UE4 cooked assets.

The data in this repo is not authoritative and may have inaccurate values.

## Data

All data is versioned by game patch under `data/game/<version>/`:

### `weapons.json`
Per-weapon stats for all 46 weapons:
- Base damage (per-pellet for shotguns)
- Pellet count (`attackCount` for shotguns)
- Body-part damage multipliers (head/body/leg)
- Distance falloff curve (fullDamageRange, falloffEndRange, effectiveRange, factorEffective, factorMaximal)
- Pre-computed damage at standard distances (10m/30m/50m or 10m/20m/30m for shotguns)
- Ammo capacity, spread, ADS speed

### `agents.json`
All 23 agents with role ID, class, and primary weapon assignment:
- **Duelist** (8): Ming, Bai Mo, Fuchsia, Flavia, Eika, Mara, Chiyo, Cielle
- **Sentinel** (4): Audrey, Leona, Michele, Nobunaga
- **Support** (3): Xinyi, Fragrans, Kokona
- **Vanguard** (3): Galatea, Kanami, Lawine
- **Controller** (5): Maddelena, Meredith, Reiichi, Yugiri, Yvette

### `agents/<Name>.json`
Per-agent files with full primary weapon stats (damage, body multipliers, distance falloff table) and known gun upgrade modifiers (e.g., Cielle's single-shot mode).

### `growth.json`
Shared upgrade slot definitions (armor, stringification).

## Extraction Tool

```bash
go run ./cmd/extract-game-data -version 1.8.2.4 -content <path-to-Content/> -out data/game/
```

Reads cooked UE4 `.uasset`/`.uexp` files from an extracted game `Content/` directory and produces all JSON files. Parses:
- Weapon blueprint `ScalableFloat` properties (AttackDamage, DistanceProtected*, FactorEffective/Maximal)
- `BodyDamageFactor` MapProperty (20 body zones simplified to head/body/leg)
- `AttackCount` ByteProperty (pellet count for shotguns)
- `CT_WeaponAttribute` CurveTable (damage overrides for snipers, melee, mode variants)
- Sniper charge stages (M82A1: 105→135→165→195, M200: 100→120→140→160)

### Damage Falloff Formula

Distance damage uses a quadratic falloff curve reverse-engineered from the game:

- **Within protected zone** (d ≤ protectedEnd): `damage = baseDamage`
- **Falloff zone** (protectedEnd < d ≤ effective): `damage = floor(base × (1 - frac² × FactorEffective))` where `frac = (d - protEnd) / (effective - protEnd)`
- **Beyond effective** (d > effective): `damage = floor(base × (1 - FactorMaximal))`

Verified against in-game weapon stats UI for M82A1 (105/103/95), M200 (100/99/96), Vector (21/14/14), AA12 (56/39/28).

## Go Library

`pkg/gamedata` provides a standalone Go library for loading and querying the data:

```go
vd, _ := gamedata.LoadVersionedData("data/game", "1.8.2.4")

// Weapon damage lookup
dmg := vd.GetWeaponDamage("10108001") // SCARH → 27

// Agent lookup
agent := vd.GetAgentByRoleID(112) // Fuchsia

// Damage calculation with body part and armor
input := gamedata.DamageInput{
    WeaponID: "10108001",
    BodyPart: gamedata.BodyPartHead,
    ArmorReduction: 0.46, // Stringified 7B
}
dmg = vd.ComputeShotDamage(input) // 27 × 1.5 × 0.54 = 21.87
```

## Updating for New Game Versions

1. Extract `.pak` files from the game installation (e.g., using FModel on Windows)
2. Place the extracted `Content/` directory accessible to this tool
3. Run: `go run ./cmd/extract-game-data -version <new-version> -content <Content-path> -out data/game/`
4. Commit the new `data/game/<version>/` directory
