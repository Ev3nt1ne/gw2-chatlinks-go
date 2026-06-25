package chatlinks

import "testing"

func TestProfessionName(t *testing.T) {
	name, ok := ProfessionName(5)
	if !ok || name != "Thief" {
		t.Errorf("ProfessionName(5) = (%q, %v), want (Thief, true)", name, ok)
	}
	if _, ok := ProfessionName(999); ok {
		t.Error("ProfessionName(999) = ok=true, want false for an unknown id")
	}
}

func TestWeaponTypeName(t *testing.T) {
	name, ok := WeaponTypeName(90)
	if !ok || name != "Sword" {
		t.Errorf("WeaponTypeName(90) = (%q, %v), want (Sword, true)", name, ok)
	}
	if _, ok := WeaponTypeName(999); ok {
		t.Error("WeaponTypeName(999) = ok=true, want false for an unknown id")
	}
}

func TestLegendName(t *testing.T) {
	name, ok := LegendName(1)
	if !ok || name != "Legendary Dragon Stance" {
		t.Errorf("LegendName(1) = (%q, %v), want (Legendary Dragon Stance, true)", name, ok)
	}
	if _, ok := LegendName(99); ok {
		t.Error("LegendName(99) = ok=true, want false for an unknown code")
	}
}
