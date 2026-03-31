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
