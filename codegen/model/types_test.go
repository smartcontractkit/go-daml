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

func TestGenMapGoImportsIncludesKeyAndValueImports(t *testing.T) {
	keyPkg := ExternalPackage{Import: "example.com/keypkg", Alias: "keypkg"}
	valPkg := ExternalPackage{Import: "example.com/valpkg", Alias: "valpkg"}

	gm := GenMap{
		Key: Imported{
			Underlying:      Unknown{String: "KeyType"},
			ExternalPackage: keyPkg,
		},
		Value: Imported{
			Underlying:      Unknown{String: "ValueType"},
			ExternalPackage: valPkg,
		},
	}

	got := gm.GoImports()
	if len(got) != 2 {
		t.Fatalf("GenMap.GoImports() len = %d, want 2", len(got))
	}
	if got[0] != keyPkg {
		t.Fatalf("GenMap.GoImports()[0] = %+v, want %+v", got[0], keyPkg)
	}
	if got[1] != valPkg {
		t.Fatalf("GenMap.GoImports()[1] = %+v, want %+v", got[1], valPkg)
	}
}
