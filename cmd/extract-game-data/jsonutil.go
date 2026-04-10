package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func writeJSON(path string, data interface{}) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", path, err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
	}

	keys := make([]string, 0)
	if wf, ok := data.(WeaponsFile); ok {
		for _, wd := range wf.Weapons {
			if wd.AttackDamage > 0 {
				keys = append(keys, fmt.Sprintf("%-20s dmg=%.0f", wd.Name, wd.AttackDamage))
			}
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s\n", k)
	}
}
