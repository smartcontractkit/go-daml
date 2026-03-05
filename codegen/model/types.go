package model

import (
	"strings"
)

type ExternalPackages struct {
	// Maps packageId => package
	Packages map[string]ExternalPackage
}

type ExternalPackage struct {
	// The import to use to refer to this package, e.g. github.com/smartcontractkit/go-daml/etc
	Import string
	// The alias to import this package by
	Alias string
}

// Daml Types

type DamlType interface {
	GoType() string
	GoImport() *ExternalPackage
}

type noImport struct{}

func (noImport) GoImport() *ExternalPackage { return nil }

type List struct {
	Inner DamlType
}

func (t List) GoType() string {
	return "[]" + t.Inner.GoType()
}

func (t List) GoImport() *ExternalPackage {
	return t.Inner.GoImport()
}

type Party struct {
	noImport
}

func (t Party) GoType() string {
	return "types.PARTY"
}

type Text struct {
	noImport
}

func (t Text) GoType() string {
	return "types.TEXT"
}

type BytesHex struct {
	noImport
}

func (t BytesHex) GoType() string {
	return "types.TEXT" // BytesHex is represented as TEXT in Go
}

// IsBytesHex returns true - used by template to add hex:"bytes16" tag
func (t BytesHex) IsBytesHex() bool {
	return true
}

// BytesHexFieldNames contains field names that should use uint16 length prefix
// encoding (hex:"bytes16" tag) instead of the default uint8 length prefix.
//
// Background: The Daml BytesHex type synonym is expanded to Text by the compiler,
// so we cannot detect it from the compiled Daml LF. Instead, we use a field name
// allowlist for fields that are known to potentially exceed 255 bytes.
//
// Fields requiring bytes16 encoding (from MCMS/Codec.daml):
// - operationData: Serialized choice parameters in TimelockCall (encodeUint16)
// - predecessor: Hex hash in ScheduleBatchParams (encodeUint16)
// - salt: Hex value in ScheduleBatchParams (encodeUint16)
//
// To add a new field: add its name (case-sensitive) as a key with value true.
var BytesHexFieldNames = map[string]bool{
	"operationData": true,
	"predecessor":   true,
	"salt":          true,
}

// Uint32FieldNames contains field names where INT64 should encode as 4-byte uint32.
// This matches the Daml MCMS/Codec.daml encoding which uses encodeUint32 for these fields.
//
// From MCMS/Codec.daml encodeSignerInfo:
// - signerIndex: encodeUint32 (4 bytes)
// - signerGroup: encodeUint32 (4 bytes)
var Uint32FieldNames = map[string]bool{
	"signerIndex": true,
	"signerGroup": true,
}

// Uint32ListFieldNames contains field names where []INT64 should encode as
// length + uint32 elements (4 bytes each) instead of uint64 elements (8 bytes).
// This matches the Daml MCMS/Codec.daml encoding which uses encodeUint32 for each element.
//
// From MCMS/Codec.daml encodeSetConfigParams and encodeMultisigConfig:
// - groupQuorums: list of encodeUint32 (4 bytes each)
// - groupParents: list of encodeUint32 (4 bytes each)
// - apGroupQuorums: used in AdminParams.AP_SetConfig
// - apGroupParents: used in AdminParams.AP_SetConfig
var Uint32ListFieldNames = map[string]bool{
	"groupQuorums":   true,
	"groupParents":   true,
	"apGroupQuorums": true,
	"apGroupParents": true,
}

type Int64 struct {
	noImport
}

func (t Int64) GoType() string {
	return "types.INT64"
}

type Bool struct {
	noImport
}

func (t Bool) GoType() string {
	return "types.BOOL"
}

type Decimal struct {
	noImport
}

func (t Decimal) GoType() string {
	return "types.DECIMAL"
}

type Numeric struct {
	noImport
}

func (t Numeric) GoType() string {
	return "types.NUMERIC"
}

type Date struct {
	noImport
}

func (t Date) GoType() string {
	return "types.DATE"
}

type Timestamp struct {
	noImport
}

func (t Timestamp) GoType() string {
	return "types.TIMESTAMP"
}

type Unit struct {
	noImport
}

func (t Unit) GoType() string {
	return "types.UNIT"
}

type Map struct {
	noImport
}

func (t Map) GoType() string {
	return "types.MAP"
}

type Optional struct {
	Inner DamlType
}

func (t Optional) GoType() string {
	return "*" + t.Inner.GoType()
}

func (t Optional) GoImport() *ExternalPackage {
	return t.Inner.GoImport()
}

type ContractId struct {
	noImport
}

func (t ContractId) GoType() string {
	return "types.CONTRACT_ID"
}

type GenMap struct {
	noImport
}

func (t GenMap) GoType() string {
	return "types.GENMAP"
}

type TextMap struct {
	noImport
}

func (t TextMap) GoType() string {
	return "types.TEXTMAP"
}

type BigNumeric struct {
	noImport
}

func (t BigNumeric) GoType() string {
	return "types.BIGNUMERIC"
}

type RoundingMode struct {
	noImport
}

func (t RoundingMode) GoType() string {
	return "types.ROUNDING_MODE"
}

type Any struct {
	noImport
}

func (t Any) GoType() string {
	return "any"
}

type RelTime struct {
	noImport
}

func (t RelTime) GoType() string {
	return "types.RELTIME"
}

type Set struct {
	noImport
}

func (t Set) GoType() string {
	return "SET"
}

type Enum struct {
	noImport
}

func (t Enum) GoType() string {
	return "string"
}

type Imported struct {
	Underlying      DamlType
	ExternalPackage ExternalPackage
}

func (t Imported) GoType() string {
	return t.ExternalPackage.Alias + "." + t.Underlying.GoType()
}

func (t Imported) GoImport() *ExternalPackage {
	return &t.ExternalPackage
}

type Unknown struct {
	String string
	noImport
}

func (t Unknown) GoType() string {
	// Retain previous behavior of stripping all underscores, matched what the `capitalise` function in the template does.
	return strings.ReplaceAll(t.String, "_", "")
}
