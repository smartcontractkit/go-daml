package model

import (
	"testing"

	"github.com/smartcontractkit/go-daml/pkg/types"
)

func TestNestedToDAMLValue_PreservesSET(t *testing.T) {
	input := types.SET{
		types.TEXT("alice"),
		types.TEXT("bob"),
	}

	got := NestedToDAMLValue(input)

	set, ok := got.(types.SET)
	if !ok {
		t.Fatalf("NestedToDAMLValue() type = %T, want types.SET", got)
	}
	if len(set) != 2 {
		t.Fatalf("NestedToDAMLValue() len = %d, want 2", len(set))
	}
	if set[0] != types.TEXT("alice") || set[1] != types.TEXT("bob") {
		t.Fatalf("NestedToDAMLValue() = %#v, want original set elements preserved", set)
	}
}
