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

func TestTextMapGoTypeTypedValue(t *testing.T) {
	tm := TextMap{
		Value: Numeric{},
	}

	got := tm.GoType()
	if got != "map[string]types.NUMERIC" {
		t.Errorf("TextMap.GoType() = %q, want %q", got, "map[string]types.NUMERIC")
	}
}

func TestTextMapGoTypeUntypedFallsBack(t *testing.T) {
	tm := TextMap{}

	got := tm.GoType()
	if got != "types.TEXTMAP" {
		t.Errorf("TextMap.GoType() = %q, want %q", got, "types.TEXTMAP")
	}
}
