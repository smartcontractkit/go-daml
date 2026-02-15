package codegen_test

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/smartcontractkit/go-daml/pkg/bind"
	"github.com/smartcontractkit/go-daml/pkg/codec"
	"github.com/smartcontractkit/go-daml/pkg/model"
	"github.com/smartcontractkit/go-daml/pkg/types"
)

var (
	_ = fmt.Sprintf
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = model.Command{}
	_ bind.BoundTemplate
)

const PackageName = "all-kinds-of"
const SDKVersion = "3.3.0-snapshot.20250417.0"

type Template interface {
	CreateCommand() *model.CreateCommand
	GetTemplateID() string
}

func argsToMap(args any) map[string]any {
	if args == nil {
		return map[string]any{}
	}

	if m, ok := args.(map[string]any); ok {
		return m
	}

	type mapper interface {
		ToMap() map[string]any
	}
	if mapper, ok := args.(mapper); ok {
		return mapper.ToMap()
	}

	return map[string]any{"args": args}
}

// Accept is a Record type
type Accept struct {
}

// ToMap converts Accept to a map for DAML arguments
func (t Accept) ToMap() map[string]any {
	m := make(map[string]any)
	return m
}

func (t Accept) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *Accept) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Color is an enum type
type Color string

const (
	ColorRed Color = "Red"

	ColorGreen Color = "Green"

	ColorBlue Color = "Blue"
)

func (e Color) GetEnumConstructor() string { return string(e) }

func (e Color) GetEnumTypeID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "AllKindsOf", "Color")
}

// GetEnumTypeIDWithPackageID returns the enum type ID using the provided package ID instead of package name
func (e Color) GetEnumTypeIDWithPackageID(packageID string) string {
	return fmt.Sprintf("#%s:%s:%s", packageID, "AllKindsOf", "Color")
}

func (e Color) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(e)
}

func (e *Color) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, e)
}

var _ types.ENUM = Color("")

// MappyContract is a Template type
type MappyContract struct {
	Operator types.PARTY   `json:"operator"`
	Value    types.TEXTMAP `json:"value"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t MappyContract) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "AllKindsOf", "MappyContract")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t MappyContract) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "AllKindsOf", "MappyContract")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t MappyContract) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["operator"] = t.Operator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["value"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Value).(mapper); ok {
			return m.toMap()
		}
		return t.Value
	}()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t MappyContract) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["operator"] = t.Operator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["value"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Value).(mapper); ok {
			return m.toMap()
		}
		return t.Value
	}()

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t MappyContract) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *MappyContract) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for MappyContract

// Archive exercises the Archive choice on this MappyContract contract
// This method uses the package name in the template ID
func (t MappyContract) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "AllKindsOf", "MappyContract"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t MappyContract) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "AllKindsOf", "MappyContract"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// MyPair is a Record type
type MyPair struct {
	Left  any `json:"left"`
	Right any `json:"right"`
}

// ToMap converts MyPair to a map for DAML arguments
func (t MyPair) ToMap() map[string]any {
	m := make(map[string]any)

	m["left"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Left).(mapper); ok {
			return m.toMap()
		}
		return t.Left
	}()

	m["right"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.Right).(mapper); ok {
			return m.toMap()
		}
		return t.Right
	}()

	return m
}

func (t MyPair) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *MyPair) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// OneOfEverything is a Template type
type OneOfEverything struct {
	Operator        types.PARTY     `json:"operator"`
	SomeBoolean     types.BOOL      `json:"someBoolean"`
	SomeInteger     types.INT64     `json:"someInteger"`
	SomeDecimal     types.NUMERIC   `json:"someDecimal"`
	SomeMaybe       *types.INT64    `json:"someMaybe"`
	SomeMaybeNot    *types.INT64    `json:"someMaybeNot"`
	SomeText        types.TEXT      `json:"someText"`
	SomeDate        types.DATE      `json:"someDate"`
	SomeDatetime    types.TIMESTAMP `json:"someDatetime"`
	SomeSimpleList  []types.INT64   `json:"someSimpleList"`
	SomeSimplePair  MyPair          `json:"someSimplePair"`
	SomeNestedPair  MyPair          `json:"someNestedPair"`
	SomeUglyNesting VPair           `json:"someUglyNesting"`
	SomeMeasurement types.NUMERIC   `json:"someMeasurement"`
	SomeEnum        Color           `json:"someEnum"`
	TheUnit         types.UNIT      `json:"theUnit"`
}

// GetTemplateID returns the template ID for this template using the package name
func (t OneOfEverything) GetTemplateID() string {
	return fmt.Sprintf("#%s:%s:%s", PackageName, "AllKindsOf", "OneOfEverything")
}

// GetTemplateIDWithPackageID returns the template ID using the provided package ID instead of package name
func (t OneOfEverything) GetTemplateIDWithPackageID(packageID string) string {
	return fmt.Sprintf("%s:%s:%s", packageID, "AllKindsOf", "OneOfEverything")
}

// CreateCommand returns a CreateCommand for this template using the package name
func (t OneOfEverything) CreateCommand() *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["operator"] = t.Operator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someBoolean"] = bool(t.SomeBoolean)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someInteger"] = int64(t.SomeInteger)

	if t.SomeDecimal != "" {
		args["someDecimal"] = t.SomeDecimal
	}

	if t.SomeMaybe != nil {
		args["someMaybe"] = map[string]any{
			"_type": "optional",
			"value": int64(*t.SomeMaybe),
		}
	} else {
		args["someMaybe"] = map[string]any{
			"_type": "optional",
		}
	}

	if t.SomeMaybeNot != nil {
		args["someMaybeNot"] = map[string]any{
			"_type": "optional",
			"value": int64(*t.SomeMaybeNot),
		}
	} else {
		args["someMaybeNot"] = map[string]any{
			"_type": "optional",
		}
	}

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someText"] = string(t.SomeText)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someDate"] = t.SomeDate

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someDatetime"] = t.SomeDatetime

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someSimpleList"] = func() []any {
		res := make([]any, 0, len(t.SomeSimpleList))
		for _, e := range t.SomeSimpleList {
			res = append(res, int64(e))
		}
		return res
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someSimplePair"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.SomeSimplePair).(mapper); ok {
			return m.toMap()
		}
		return t.SomeSimplePair
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someNestedPair"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.SomeNestedPair).(mapper); ok {
			return m.toMap()
		}
		return t.SomeNestedPair
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someUglyNesting"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.SomeUglyNesting).(mapper); ok {
			return m.toMap()
		}
		return t.SomeUglyNesting
	}()

	if t.SomeMeasurement != "" {
		args["someMeasurement"] = t.SomeMeasurement
	}

	if t.SomeEnum != "" {
		args["someEnum"] = func() any {
			type mapper interface{ toMap() map[string]any }
			if m, ok := any(t.SomeEnum).(mapper); ok {
				return m.toMap()
			}
			return t.SomeEnum
		}()
	}

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["theUnit"] = map[string]any{"_type": "unit"}

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// CreateCommandWithPackageID returns a CreateCommand using the provided package ID instead of package name
func (t OneOfEverything) CreateCommandWithPackageID(packageID string) *model.CreateCommand {
	args := make(map[string]any)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["operator"] = t.Operator.ToMap()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someBoolean"] = bool(t.SomeBoolean)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someInteger"] = int64(t.SomeInteger)

	if t.SomeDecimal != "" {
		args["someDecimal"] = t.SomeDecimal
	}

	if t.SomeMaybe != nil {
		args["someMaybe"] = map[string]any{
			"_type": "optional",
			"value": int64(*t.SomeMaybe),
		}
	} else {
		args["someMaybe"] = map[string]any{
			"_type": "optional",
		}
	}

	if t.SomeMaybeNot != nil {
		args["someMaybeNot"] = map[string]any{
			"_type": "optional",
			"value": int64(*t.SomeMaybeNot),
		}
	} else {
		args["someMaybeNot"] = map[string]any{
			"_type": "optional",
		}
	}

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someText"] = string(t.SomeText)

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someDate"] = t.SomeDate

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someDatetime"] = t.SomeDatetime

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someSimpleList"] = func() []any {
		res := make([]any, 0, len(t.SomeSimpleList))
		for _, e := range t.SomeSimpleList {
			res = append(res, int64(e))
		}
		return res
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someSimplePair"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.SomeSimplePair).(mapper); ok {
			return m.toMap()
		}
		return t.SomeSimplePair
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someNestedPair"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.SomeNestedPair).(mapper); ok {
			return m.toMap()
		}
		return t.SomeNestedPair
	}()

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["someUglyNesting"] = func() any {
		type mapper interface{ toMap() map[string]any }
		if m, ok := any(t.SomeUglyNesting).(mapper); ok {
			return m.toMap()
		}
		return t.SomeUglyNesting
	}()

	if t.SomeMeasurement != "" {
		args["someMeasurement"] = t.SomeMeasurement
	}

	if t.SomeEnum != "" {
		args["someEnum"] = func() any {
			type mapper interface{ toMap() map[string]any }
			if m, ok := any(t.SomeEnum).(mapper); ok {
				return m.toMap()
			}
			return t.SomeEnum
		}()
	}

	// IMPORTANT: always include non-optional fields (GENMAP/MAP/LIST/[] etc), even if empty
	args["theUnit"] = map[string]any{"_type": "unit"}

	return &model.CreateCommand{
		TemplateID: t.GetTemplateIDWithPackageID(packageID),
		Arguments:  args,
	}
}

func (t OneOfEverything) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(t)
}

func (t *OneOfEverything) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, t)
}

// Choice methods for OneOfEverything

// Archive exercises the Archive choice on this OneOfEverything contract
// This method uses the package name in the template ID
func (t OneOfEverything) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "AllKindsOf", "OneOfEverything"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// ArchiveWithPackageID exercises the Archive choice using the provided package ID instead of package name
func (t OneOfEverything) ArchiveWithPackageID(contractID string, packageID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "AllKindsOf", "OneOfEverything"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]any{},
	}
}

// Accept exercises the Accept choice on this OneOfEverything contract
// This method uses the package name in the template ID
func (t OneOfEverything) Accept(contractID string, args Accept) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", PackageName, "AllKindsOf", "OneOfEverything"),
		ContractID: contractID,
		Choice:     "Accept",
		Arguments:  argsToMap(args),
	}
}

// AcceptWithPackageID exercises the Accept choice using the provided package ID instead of package name
func (t OneOfEverything) AcceptWithPackageID(contractID string, packageID string, args Accept) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("#%s:%s:%s", packageID, "AllKindsOf", "OneOfEverything"),
		ContractID: contractID,
		Choice:     "Accept",
		Arguments:  argsToMap(args),
	}
}

// VPair is a variant/union type
type VPair struct {
	Left  *any   `json:"Left,omitempty"`
	Right *any   `json:"Right,omitempty"`
	Both  *VPair `json:"Both,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for VPair
func (v VPair) MarshalJSON() ([]byte, error) {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Marshal(v)
}

// UnmarshalJSON implements custom JSON unmarshalling for VPair
func (v *VPair) UnmarshalJSON(data []byte) error {
	jsonCodec := codec.NewJsonCodec()
	return jsonCodec.Unmarshal(data, v)
}

// GetVariantTag implements types.VARIANT interface
func (v VPair) GetVariantTag() string {

	if v.Left != nil {
		return "Left"
	}

	if v.Right != nil {
		return "Right"
	}

	if v.Both != nil {
		return "Both"
	}

	return ""
}

// GetVariantValue implements types.VARIANT interface
func (v VPair) GetVariantValue() any {

	if v.Left != nil {
		return v.Left
	}

	if v.Right != nil {
		return v.Right
	}

	if v.Both != nil {
		return v.Both
	}

	return nil
}

var _ types.VARIANT = (*VPair)(nil)
