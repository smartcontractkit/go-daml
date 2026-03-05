package codegen

import (
	"strings"
	"testing"

	"github.com/smartcontractkit/go-daml/codegen/model"
)

func TestBind(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"RentalProposal": {
			Name:    "RentalProposal",
			RawType: "Record",
			Fields: []*model.TmplField{
				{Name: "landlord", Type: model.Text{}},
				{Name: "tenant", Type: model.Text{}},
				{Name: "terms", Type: model.Text{}},
			},
		},
		"RentalAgreement": {
			Name:    "RentalAgreement",
			RawType: "Record",
			Fields: []*model.TmplField{
				{Name: "landlord", Type: model.Text{}},
				{Name: "tenant", Type: model.Text{}},
				{Name: "terms", Type: model.Text{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package-name",
		Structs: structs,
	}

	result, err := Bind("main", pkg, "2.0.0", true, false)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// Check that the result contains expected content
	if !strings.Contains(result, "package main") {
		t.Error("Generated code does not contain correct package declaration")
	}

	if !strings.Contains(result, "type RentalProposal struct") {
		t.Error("Generated code does not contain RentalProposal struct")
	}

	if !strings.Contains(result, "type RentalAgreement struct") {
		t.Error("Generated code does not contain RentalAgreement struct")
	}

	if !strings.Contains(result, "Landlord types.TEXT") {
		t.Error("Generated code does not contain capitalized field names")
	}

	if !strings.Contains(result, `json:"landlord"`) {
		t.Error("Generated code does not contain JSON tags with original field names")
	}

	if !strings.Contains(result, `SDKVersion  = "2.0.0"`) {
		t.Error("Generated code does not contain SDKVersion constant")
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"landlord", "Landlord"},
		{"rental_proposal", "RentalProposal"},
		{"TENANT", "TENANT"},
		{"", ""},
		{"a", "A"},
		{"FeaturedAppRightCreateActivityMarker", "FeaturedAppRightCreateActivityMarker"},
		{"FeaturedAppRight_CreateActivityMarker", "FeaturedAppRightCreateActivityMarker"},
		{"FEATUREDAPPRIGHTCREATACTIVITYMARKER", "FEATUREDAPPRIGHTCREATACTIVITYMARKER"},
	}

	for _, test := range tests {
		result := capitalize(test.input)
		if result != test.expected {
			t.Errorf("capitalize(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestDecapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Landlord", "landlord"},
		{"RentalProposal", "rentalProposal"},
		{"TENANT", "tenant"},
		{"", ""},
		{"A", "a"},
		{"FeaturedAppRightCreateActivityMarker", "featuredAppRightCreateActivityMarker"},
	}

	for _, test := range tests {
		result := decapitalize(test.input)
		if result != test.expected {
			t.Errorf("decapitalize(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestBytesHexFieldNames(t *testing.T) {
	// Test that BytesHexFieldNames contains expected entries
	if !model.BytesHexFieldNames["operationData"] {
		t.Error("BytesHexFieldNames should contain 'operationData'")
	}

	// Test that non-BytesHex fields are not in the map
	if model.BytesHexFieldNames["someOtherField"] {
		t.Error("BytesHexFieldNames should not contain 'someOtherField'")
	}
}

func TestBindWithBytesHexField(t *testing.T) {
	// Test that fields with IsBytesHex=true get hex:"bytes16" tag
	structs := map[string]*model.TmplStruct{
		"TimelockCall": {
			Name:    "TimelockCall",
			RawType: "Record",
			Fields: []*model.TmplField{
				{Name: "targetInstanceAddress", Type: model.Text{}},
				{Name: "functionName", Type: model.Text{}},
				{Name: "operationData", Type: model.Text{}, IsBytesHex: true},
			},
		},
	}

	pkg := &model.Package{
		Name:    "mcms",
		Structs: structs,
	}

	result, err := Bind("mcms", pkg, "2.0.0", true, true)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// Check that operationData field has hex:"bytes16" tag
	if !strings.Contains(result, `json:"operationData" hex:"bytes16"`) {
		t.Error("Generated code should contain hex:\"bytes16\" tag for operationData field")
	}

	// Check that other fields do NOT have hex:"bytes16" tag
	if strings.Contains(result, `json:"functionName" hex:"bytes16"`) {
		t.Error("Generated code should NOT contain hex:\"bytes16\" tag for functionName field")
	}

	if strings.Contains(result, `json:"targetInstanceAddress" hex:"bytes16"`) {
		t.Error("Generated code should NOT contain hex:\"bytes16\" tag for targetInstanceAddress field")
	}
}

func TestBindWithoutBytesHexField(t *testing.T) {
	// Test that fields without IsBytesHex do not get hex:"bytes16" tag
	structs := map[string]*model.TmplStruct{
		"SimpleRecord": {
			Name:    "SimpleRecord",
			RawType: "Record",
			Fields: []*model.TmplField{
				{Name: "name", Type: model.Text{}},
				{Name: "value", Type: model.Text{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test",
		Structs: structs,
	}

	result, err := Bind("test", pkg, "2.0.0", true, true)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// Check that no fields have hex:"bytes16" tag
	if strings.Contains(result, `hex:"bytes16"`) {
		t.Error("Generated code should NOT contain hex:\"bytes16\" tag when IsBytesHex is false")
	}
}
