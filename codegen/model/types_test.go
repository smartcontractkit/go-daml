package model

import (
	"testing"
)

func TestSetGoType(t *testing.T) {
	s := Set{}
	got := s.GoType()
	if got != "types.SET" {
		t.Errorf("Set.GoType() = %q, want %q", got, "types.SET")
	}
}

func TestRelTimeGoType(t *testing.T) {
	r := RelTime{}
	got := r.GoType()
	if got != "types.RELTIME" {
		t.Errorf("RelTime.GoType() = %q, want %q", got, "types.RELTIME")
	}
}

func TestGenMapGoTypeTypedScalarKey(t *testing.T) {
	gm := GenMap{
		Key:   Text{},
		Value: Numeric{},
	}

	got := gm.GoType()
	if got != "map[types.TEXT]types.NUMERIC" {
		t.Errorf("GenMap.GoType() = %q, want %q", got, "map[types.TEXT]types.NUMERIC")
	}
}

func TestGenMapGoTypeUnsupportedKeyFallsBack(t *testing.T) {
	gm := GenMap{
		Key:   Date{},
		Value: Text{},
	}

	got := gm.GoType()
	if got != "types.GENMAP" {
		t.Errorf("GenMap.GoType() = %q, want %q", got, "types.GENMAP")
	}
}
