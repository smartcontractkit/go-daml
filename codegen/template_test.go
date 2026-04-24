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

func TestBindWithSetField(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"MessageStore": {
			Name:    "MessageStore",
			RawType: "Record",
			Fields: []*model.TmplField{
				{Name: "owner", Type: model.Text{}},
				{Name: "executedMessages", Type: model.Set{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("main", pkg, "2.0.0", true, false)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !strings.Contains(result, "ExecutedMessages types.SET") {
		t.Errorf("Generated code should contain 'ExecutedMessages types.SET', got:\n%s", result)
	}

	if !strings.Contains(result, `json:"executedMessages"`) {
		t.Error("Generated code should contain JSON tag for executedMessages")
	}
}

func TestBindUsesSharedNestedToDAMLValueHelper(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"DocumentVerifier": {
			Name:       "DocumentVerifier",
			ModuleName: "Example",
			RawType:    "Template",
			IsTemplate: true,
			Fields: []*model.TmplField{
				{Name: "operator", Type: model.Party{}},
			},
		},
		"DeployDocumentVerifier": {
			Name:       "DeployDocumentVerifier",
			ModuleName: "Example",
			RawType:    "Record",
			Fields: []*model.TmplField{
				{Name: "contract", Type: model.Unknown{String: "DocumentVerifier"}},
			},
		},
		"ExampleFactory": {
			Name:       "ExampleFactory",
			ModuleName: "Example",
			RawType:    "Template",
			IsTemplate: true,
			Fields: []*model.TmplField{
				{Name: "operator", Type: model.Party{}},
			},
			Choices: []*model.TmplChoice{
				{Name: "deployDocumentVerifier", ArgType: model.Unknown{String: "DeployDocumentVerifier"}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("main", pkg, "2.0.0", true, false)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if strings.Contains(result, "func nestedToDAMLValue(v any) any") {
		t.Error("Generated code should not emit a per-package nested helper")
	}

	if !strings.Contains(result, "model.NestedToDAMLValue") {
		t.Error("Generated code should use model.NestedToDAMLValue for nested values")
	}

	if !strings.Contains(result, `m["contract"] = model.NestedToDAMLValue(t.Contract)`) {
		t.Error("Generated code should use shared helper for nested template fields")
	}
}

func TestBindWithRelTimeField(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"Schedule": {
			Name:    "Schedule",
			RawType: "Record",
			Fields: []*model.TmplField{
				{Name: "duration", Type: model.RelTime{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("main", pkg, "2.0.0", true, false)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !strings.Contains(result, "Duration types.RELTIME") {
		t.Errorf("Generated code should contain 'Duration types.RELTIME', got:\n%s", result)
	}
}

func TestBindWithTypedGenMapField(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"MapHolder": {
			Name:    "MapHolder",
			RawType: "Record",
			Fields: []*model.TmplField{
				{
					Name: "values",
					Type: model.GenMap{
						Key:   model.Text{},
						Value: model.Numeric{},
					},
				},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("main", pkg, "2.0.0", true, false)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !strings.Contains(result, "Values map[types.TEXT]types.NUMERIC") {
		t.Fatalf("generated code should contain typed GENMAP field, got:\n%s", result)
	}

	if !strings.Contains(result, `"genmap"`) || !strings.Contains(result, `"value": t.Values`) {
		t.Fatalf("generated code should wrap typed GENMAP fields for DAML encoding, got:\n%s", result)
	}
}

func TestBindWithTypedTextMapField(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"MapHolder": {
			Name:    "MapHolder",
			RawType: "Record",
			Fields: []*model.TmplField{
				{
					Name: "values",
					Type: model.TextMap{
						Value: model.Numeric{},
					},
				},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("main", pkg, "2.0.0", true, false)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !strings.Contains(result, "Values map[string]types.NUMERIC") {
		t.Fatalf("generated code should contain typed TEXTMAP field, got:\n%s", result)
	}

	if !strings.Contains(result, `"textmap"`) || !strings.Contains(result, `"value": t.Values`) {
		t.Fatalf("generated code should wrap typed TEXTMAP fields for DAML encoding, got:\n%s", result)
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

func TestFieldHints(t *testing.T) {
	// An empty FieldHints should not match any field name
	empty := model.FieldHints{}
	if empty.BytesHexFields["operationData"] {
		t.Error("empty FieldHints.BytesHexFields should not match any field")
	}
	if empty.BytesFields["signerAddress"] {
		t.Error("empty FieldHints.BytesFields should not match any field")
	}

	// A populated FieldHints should match exactly the configured fields
	hints := model.FieldHints{
		BytesHexFields: map[string]bool{"operationData": true},
		BytesFields:    map[string]bool{"signerAddress": true},
	}
	if !hints.BytesHexFields["operationData"] {
		t.Error("hints.BytesHexFields should contain 'operationData'")
	}
	if hints.BytesHexFields["someOtherField"] {
		t.Error("hints.BytesHexFields should not contain 'someOtherField'")
	}
	if !hints.BytesFields["signerAddress"] {
		t.Error("hints.BytesFields should contain 'signerAddress'")
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

func TestBindChoiceEncoderUsesDAMLChoiceNamesForDedupedTypes(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"Workflow": {
			Name:       "Workflow",
			DAMLName:   "Workflow",
			RawType:    "Template",
			IsTemplate: true,
			Choices: []*model.TmplChoice{
				{Name: "ApproveTransfer", ArgType: model.Unknown{String: "ApproveTransfer2"}},
				{Name: "ApplyConfiguration", ArgType: model.Unknown{String: "ApplyConfiguration2"}},
				{Name: "AssignHandler", ArgType: model.Unknown{String: "AssignHandler"}},
			},
		},
		"ApproveTransfer2": {
			Name:     "ApproveTransfer2",
			DAMLName: "ApproveTransfer",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "caller", Type: model.Party{}},
			},
		},
		"ApplyConfiguration2": {
			Name:     "ApplyConfiguration2",
			DAMLName: "ApplyConfiguration",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "caller", Type: model.Party{}},
			},
		},
		"ApplyConfigurationParams2": {
			Name:     "ApplyConfigurationParams2",
			DAMLName: "ApplyConfigurationParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "value", Type: model.Text{}},
			},
		},
		"AssignHandler": {
			Name:     "AssignHandler",
			DAMLName: "AssignHandler",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "caller", Type: model.Party{}},
			},
		},
		"AssignHandlerParams": {
			Name:     "AssignHandlerParams",
			DAMLName: "AssignHandlerParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "value", Type: model.Text{}},
			},
		},
		"UnusedActionParams": {
			Name:     "UnusedActionParams",
			DAMLName: "UnusedActionParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "value", Type: model.Text{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-workflow",
		Structs: structs,
	}

	result, err := Bind("workflow", pkg, "3.4.10", true, true)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !strings.Contains(result, `func (e *encoder) ApproveTransfer(args ApproveTransfer2) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should prefer the original DAML choice name for the encoder method")
	}
	if !strings.Contains(result, `func (e *encoder) ApproveTransferMCMSParams(args ApproveTransfer2MCMSParams) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should use the original DAML choice name for MCMSParams helpers")
	}
	if !strings.Contains(result, `return e.EncodeChoiceArgs("ApproveTransfer", args)`) {
		t.Error("Generated code should encode the original DAML choice name")
	}
	if strings.Contains(result, `return e.EncodeChoiceArgs("ApproveTransfer2", args)`) {
		t.Error("Generated code should not encode the deduped Go method name")
	}
	if !strings.Contains(result, `func (e *encoder) ApplyConfiguration(args ApplyConfiguration2) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should use the original DAML choice name for deduped direct choice args")
	}
	if !strings.Contains(result, `func (e *encoder) ApplyConfigurationParams(args ApplyConfigurationParams2) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should keep Params in the method name when the clean choice method is already used")
	}
	if strings.Contains(result, `return e.EncodeChoiceArgs("ApplyConfiguration2", args)`) {
		t.Error("Generated code should not encode deduped choice names")
	}
	if !strings.Contains(result, `func (e *encoder) AssignHandlerParams(args AssignHandlerParams) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should keep Params in the method name when a direct choice-arg method exists")
	}
	if strings.Contains(result, `func (e *encoder) UnusedAction(`) {
		t.Error("Generated code should not create Params encoders for non-choice names")
	}
}

func TestBindChoiceEncoderKeepsHintedParamHelpers(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"Dispatcher": {
			Name:       "Dispatcher",
			DAMLName:   "Dispatcher",
			RawType:    "Template",
			IsTemplate: true,
			Choices: []*model.TmplChoice{
				{Name: "ExecuteOperation", ArgType: model.Unknown{String: "ExecuteOperation"}},
			},
		},
		"ExecuteOperation": {
			Name:     "ExecuteOperation",
			DAMLName: "ExecuteOperation",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "submitter", Type: model.Party{}},
			},
		},
		"ProcessBatchParams": {
			Name:     "ProcessBatchParams",
			DAMLName: "ProcessBatchParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "salt", Type: model.Text{}},
			},
		},
		"CancelOperationParams": {
			Name:     "CancelOperationParams",
			DAMLName: "CancelOperationParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "opId", Type: model.Text{}},
			},
		},
		"ExecuteQueuedOperationsParams": {
			Name:     "ExecuteQueuedOperationsParams",
			DAMLName: "ExecuteQueuedOperationsParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "calls", Type: model.List{Inner: model.Text{}}},
			},
		},
		"NotAFunctionParams": {
			Name:     "NotAFunctionParams",
			DAMLName: "NotAFunctionParams",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "value", Type: model.Text{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("dispatcher", pkg, "3.4.10", true, true, model.FieldHints{
		ChoiceParamEncoderNames: map[string]bool{
			"ProcessBatch":             true,
			"CancelOperation":          true,
			"ExecuteQueuedOperations":  true,
			"DisabledDispatcherAction": false,
		},
	})
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	for _, name := range []string{"ProcessBatch", "CancelOperation", "ExecuteQueuedOperations"} {
		if !strings.Contains(result, `func (e *encoder) `+name+`(`) {
			t.Errorf("Generated code should keep %s dispatcher params encoder", name)
		}
		if !strings.Contains(result, `return e.EncodeChoiceArgs("`+name+`", args)`) {
			t.Errorf("Generated code should encode dispatcher function %s", name)
		}
	}
	if strings.Contains(result, `func (e *encoder) NotAFunction(`) {
		t.Error("Generated code should not create arbitrary Params encoders for unhinted records")
	}
	if strings.Contains(result, `func (e *encoder) DisabledDispatcherAction(`) {
		t.Error("Generated code should ignore disabled ChoiceParamEncoderNames hints")
	}
}

func TestBindChoiceEncoderFallsBackForDuplicateChoiceNames(t *testing.T) {
	structs := map[string]*model.TmplStruct{
		"FirstTemplate": {
			Name:       "FirstTemplate",
			DAMLName:   "FirstTemplate",
			RawType:    "Template",
			IsTemplate: true,
			Choices: []*model.TmplChoice{
				{Name: "Get", ArgType: model.Unknown{String: "Get"}},
			},
		},
		"SecondTemplate": {
			Name:       "SecondTemplate",
			DAMLName:   "SecondTemplate",
			RawType:    "Template",
			IsTemplate: true,
			Choices: []*model.TmplChoice{
				{Name: "Get", ArgType: model.Unknown{String: "Get2"}},
			},
		},
		"Get": {
			Name:     "Get",
			DAMLName: "Get",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "caller", Type: model.Party{}},
			},
		},
		"Get2": {
			Name:     "Get2",
			DAMLName: "Get",
			RawType:  "Record",
			Fields: []*model.TmplField{
				{Name: "caller", Type: model.Party{}},
			},
		},
	}

	pkg := &model.Package{
		Name:    "test-package",
		Structs: structs,
	}

	result, err := Bind("test", pkg, "3.4.10", true, true)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !strings.Contains(result, `func (e *encoder) Get(args Get) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should use the non-deduped Go method when available")
	}
	if !strings.Contains(result, `func (e *encoder) Get2(args Get2) (*bind.EncodedChoice, error)`) {
		t.Error("Generated code should fall back to the deduped Go method name for duplicate Daml choices")
	}
	if strings.Count(result, `return e.EncodeChoiceArgs("Get", args)`) != 4 {
		t.Error("Generated code should still encode the original duplicate Daml choice name")
	}
}
