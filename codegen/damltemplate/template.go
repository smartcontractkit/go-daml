// Package damltemplate generates Daml codec source code from parsed type definitions.
package damltemplate

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/smartcontractkit/go-daml/codegen/damlparser"
)

//go:embed source.daml.tmpl
var tmplSource string

// DamlCodecConfig configures the Daml codec generator.
type DamlCodecConfig struct {
	// ModuleName is the output module name (e.g., "MyModule.Codec")
	ModuleName string
	// TypesModule is the module containing the type definitions (e.g., "MyModule.Types")
	TypesModule string
	// CustomTypeCodecs maps type names to their custom codec functions
	CustomTypeCodecs map[string]CustomCodec
	// VariantTagByteMap maps variant type names to constructor->tag byte mappings
	VariantTagByteMap map[string]map[string]int
	// TargetTypes is the list of types to generate codecs for (empty = all)
	TargetTypes []string
}

// CustomCodec defines encode/decode functions for a custom type.
type CustomCodec struct {
	EncodeFunc   string // e.g., "encodeRawInstanceAddress"
	DecodeFunc   string // e.g., "decodeRawInstanceAddressAt"
	ImportModule string // e.g., "MyModule.Codec"
	// For list types
	EncodeListFunc string // e.g., "encodeRawInstanceAddressList" (optional)
	DecodeListFunc string // e.g., "decodeRawInstanceAddressList" (optional)
}

// tmplData is the data passed to the template.
type tmplData struct {
	ModuleName       string
	TypesModule      string
	Records          []recordData
	Variants         []variantData
	MCMSCodecImports []string
	OtherImports     []importData
}

type importData struct {
	Module string
	Funcs  []string
}

type recordData struct {
	Name   string
	Fields []fieldData
}

type fieldData struct {
	Name           string
	TypeExpr       string
	EncodeExpr     string
	DecodeFunc     string
	IsList         bool   // true for [T] types that use decodeList
	IsOptional     bool   // true for Optional T types that use decodeOptional
	ElemDecodeFunc string // element decoder for list/optional types
}

type variantData struct {
	Name         string
	Constructors []constructorData
}

type constructorData struct {
	Name        string
	PayloadType string
	TagByte     int
	EncodeExpr  string // for payload encoding
	DecodeExpr  string // for payload decoding
}

// Generate produces Daml codec source code.
func Generate(module *damlparser.ParsedModule, config DamlCodecConfig) (string, error) {
	data := buildTemplateData(module, config)

	funcs := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("daml").Funcs(funcs).Parse(tmplSource)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

func buildTemplateData(module *damlparser.ParsedModule, config DamlCodecConfig) *tmplData {
	data := &tmplData{
		ModuleName:  config.ModuleName,
		TypesModule: config.TypesModule,
	}

	// Track which MCMS.Codec functions we need (primitives)
	mcmsImports := make(map[string]bool)
	otherImports := make(map[string]map[string]bool)

	// Filter to target types if specified
	targetSet := make(map[string]bool)
	for _, t := range config.TargetTypes {
		targetSet[t] = true
	}

	// Process records
	for _, rec := range module.Records {
		if len(targetSet) > 0 && !targetSet[rec.Name] {
			continue
		}

		rd := recordData{Name: rec.Name}
		for _, field := range rec.Fields {
			encFunc, decFunc, importModule := getCodecFuncs(field.TypeExpr, config.CustomTypeCodecs)
			fd := fieldData{
				Name:       field.Name,
				TypeExpr:   field.TypeExpr,
				EncodeExpr: buildEncodeExpr(field.Name, field.TypeExpr, encFunc, config.CustomTypeCodecs),
				DecodeFunc: decFunc,
			}

			// Check if this is a generic list type that needs decodeList wrapper
			if isGenericListType(field.TypeExpr, config.CustomTypeCodecs) {
				fd.IsList = true
				fd.ElemDecodeFunc = getElemDecodeFunc(field.TypeExpr, config.CustomTypeCodecs)
				fd.DecodeFunc = "decodeList"
				// Always import encodeList/decodeList for generic lists
				mcmsImports["encodeList"] = true
				mcmsImports["decodeList"] = true
				// Also track the element type's encoder/decoder
				elemType := strings.TrimSpace(field.TypeExpr[1 : len(field.TypeExpr)-1])
				elemEnc, elemDec, elemMod := getCodecFuncs(elemType, config.CustomTypeCodecs)
				trackImport(elemMod, elemEnc, elemDec, mcmsImports, otherImports)
			}

			// Check if this is an Optional type that needs decodeOptional wrapper
			if isOptionalType(field.TypeExpr) {
				fd.IsOptional = true
				fd.ElemDecodeFunc = getOptionalElemDecodeFunc(field.TypeExpr, config.CustomTypeCodecs)
				fd.DecodeFunc = "decodeOptional"
				// Always import encodeOptional/decodeOptional
				mcmsImports["encodeOptional"] = true
				mcmsImports["decodeOptional"] = true
				// Also track the inner type's encoder/decoder
				innerType := strings.TrimPrefix(strings.TrimSpace(field.TypeExpr), "Optional ")
				innerEnc, innerDec, innerMod := getCodecFuncs(innerType, config.CustomTypeCodecs)
				trackImport(innerMod, innerEnc, innerDec, mcmsImports, otherImports)
			}

			rd.Fields = append(rd.Fields, fd)
			trackImport(importModule, encFunc, decFunc, mcmsImports, otherImports)
		}
		data.Records = append(data.Records, rd)
	}

	// Process variants
	for _, v := range module.Variants {
		if len(targetSet) > 0 && !targetSet[v.Name] {
			continue
		}

		vd := variantData{Name: v.Name}
		tagMap := config.VariantTagByteMap[v.Name]

		// Variants always need encodeUint8 and extractBytes
		mcmsImports["encodeUint8"] = true
		mcmsImports["extractBytes"] = true

		for i, ctor := range v.Constructors {
			cd := constructorData{
				Name:        ctor.Name,
				PayloadType: ctor.PayloadType,
			}

			// Use explicit tag byte if provided, else use constructor index
			if tagMap != nil {
				if tag, ok := tagMap[ctor.Name]; ok {
					cd.TagByte = tag
				} else {
					cd.TagByte = i
				}
			} else {
				cd.TagByte = i
			}

			if ctor.PayloadType != "" {
				encFunc, decFunc, importModule := getCodecFuncs(ctor.PayloadType, config.CustomTypeCodecs)
				cd.EncodeExpr = buildPayloadEncodeExpr(ctor.PayloadType, encFunc, config.CustomTypeCodecs)
				cd.DecodeExpr = decFunc
				trackImport(importModule, encFunc, decFunc, mcmsImports, otherImports)
			}
			vd.Constructors = append(vd.Constructors, cd)
		}
		data.Variants = append(data.Variants, vd)
	}

	// Build import lists
	data.MCMSCodecImports = mapKeys(mcmsImports)
	for mod, funcs := range otherImports {
		data.OtherImports = append(data.OtherImports, importData{
			Module: mod,
			Funcs:  mapKeys(funcs),
		})
	}

	return data
}

func getCodecFuncs(typeExpr string, customCodecs map[string]CustomCodec) (encodeFunc, decodeFunc, importModule string) {
	typeExpr = strings.TrimSpace(typeExpr)

	// Check custom codecs first
	if codec, ok := customCodecs[typeExpr]; ok {
		return codec.EncodeFunc, codec.DecodeFunc, codec.ImportModule
	}

	// Handle list types
	if strings.HasPrefix(typeExpr, "[") && strings.HasSuffix(typeExpr, "]") {
		elemType := strings.TrimSpace(typeExpr[1 : len(typeExpr)-1])

		// Check for custom list codec
		if codec, ok := customCodecs[elemType]; ok && codec.EncodeListFunc != "" {
			return codec.EncodeListFunc, codec.DecodeListFunc, codec.ImportModule
		}

		// Special cases for built-in list types
		switch elemType {
		case "BytesHex":
			return "encodeBytesHexList", "decodeBytesHexList", "MCMS.Codec"
		case "Party":
			return "encodePartyList", "decodePartyList", "MCMS.Codec"
		}

		// Generic list encoding
		_, elemDec, mod := getCodecFuncs(elemType, customCodecs)
		return "encodeList", elemDec, mod // encodeList needs the element encoder passed separately
	}

	// Handle Optional types
	if strings.HasPrefix(typeExpr, "Optional ") {
		return "encodeOptional", "decodeOptional", "MCMS.Codec"
	}

	// Primitive types
	switch typeExpr {
	case "Int":
		return "encodeInt64", "decodeInt64At", "MCMS.Codec"
	case "Bool":
		return "encodeBool", "decodeBoolAt", "MCMS.Codec"
	case "Text":
		return "encodeText", "decodeTextAt", "MCMS.Codec"
	case "Party":
		return "encodeParty", "decodePartyAt", "MCMS.Codec"
	case "Numeric 0":
		return "encodeNumeric0", "decodeNumeric0At", "MCMS.Codec"
	case "Decimal":
		return "encodeDecimal", "decodeDecimalAt", "MCMS.Codec"
	case "BytesHex":
		return "encodeBytesHex", "decodeBytesHexAt", "MCMS.Codec"
	}

	// Qualified type (e.g., Splice.Api.Token.HoldingV1.InstrumentId)
	if strings.Contains(typeExpr, ".") {
		// Try to find in custom codecs by short name
		parts := strings.Split(typeExpr, ".")
		shortName := parts[len(parts)-1]
		if codec, ok := customCodecs[shortName]; ok {
			return codec.EncodeFunc, codec.DecodeFunc, codec.ImportModule
		}
	}

	// Unknown type - assume it's a record type defined in same module
	return "encode" + typeExpr, "decode" + typeExpr + "At", ""
}

// isGenericListType returns true for [T] types that need decodeList wrapper
// (i.e., not special-cased like [BytesHex], [Party], or custom list codecs)
func isGenericListType(typeExpr string, customCodecs map[string]CustomCodec) bool {
	typeExpr = strings.TrimSpace(typeExpr)
	if !strings.HasPrefix(typeExpr, "[") || !strings.HasSuffix(typeExpr, "]") {
		return false
	}
	elemType := strings.TrimSpace(typeExpr[1 : len(typeExpr)-1])

	// Check for custom list codec
	if codec, ok := customCodecs[elemType]; ok && codec.EncodeListFunc != "" {
		return false // has dedicated list codec
	}

	// Special built-in list types don't need generic wrapper
	switch elemType {
	case "BytesHex", "Party":
		return false
	}

	return true
}

// getElemDecodeFunc returns the element decoder function for a list type
func getElemDecodeFunc(typeExpr string, customCodecs map[string]CustomCodec) string {
	typeExpr = strings.TrimSpace(typeExpr)
	if !strings.HasPrefix(typeExpr, "[") || !strings.HasSuffix(typeExpr, "]") {
		return ""
	}
	elemType := strings.TrimSpace(typeExpr[1 : len(typeExpr)-1])

	// Check custom codecs
	if codec, ok := customCodecs[elemType]; ok {
		return codec.DecodeFunc
	}

	// Check qualified type
	if strings.Contains(elemType, ".") {
		parts := strings.Split(elemType, ".")
		shortName := parts[len(parts)-1]
		if codec, ok := customCodecs[shortName]; ok {
			return codec.DecodeFunc
		}
	}

	// Default: assume record type in same module
	return "decode" + elemType + "At"
}

// isOptionalType returns true for Optional T types
func isOptionalType(typeExpr string) bool {
	return strings.HasPrefix(strings.TrimSpace(typeExpr), "Optional ")
}

// getOptionalElemDecodeFunc returns the element decoder function for an Optional type
func getOptionalElemDecodeFunc(typeExpr string, customCodecs map[string]CustomCodec) string {
	typeExpr = strings.TrimSpace(typeExpr)
	if !strings.HasPrefix(typeExpr, "Optional ") {
		return ""
	}
	innerType := strings.TrimSpace(strings.TrimPrefix(typeExpr, "Optional "))

	// Check custom codecs
	if codec, ok := customCodecs[innerType]; ok {
		return codec.DecodeFunc
	}

	// Check qualified type
	if strings.Contains(innerType, ".") {
		parts := strings.Split(innerType, ".")
		shortName := parts[len(parts)-1]
		if codec, ok := customCodecs[shortName]; ok {
			return codec.DecodeFunc
		}
	}

	// Primitive types
	switch innerType {
	case "Int":
		return "decodeInt64At"
	case "Bool":
		return "decodeBoolAt"
	case "Text":
		return "decodeTextAt"
	case "Party":
		return "decodePartyAt"
	case "Numeric 0":
		return "decodeNumeric0At"
	case "BytesHex":
		return "decodeBytesHexAt"
	}

	// Default: assume record type in same module
	return "decode" + innerType + "At"
}

func buildEncodeExpr(fieldName, typeExpr, encodeFunc string, customCodecs map[string]CustomCodec) string {
	typeExpr = strings.TrimSpace(typeExpr)

	// Int requires fromSome wrapper
	if typeExpr == "Int" {
		return fmt.Sprintf("fromSome (%s params.%s)", encodeFunc, fieldName)
	}

	// List types with generic encodeList
	if strings.HasPrefix(typeExpr, "[") && strings.HasSuffix(typeExpr, "]") {
		elemType := strings.TrimSpace(typeExpr[1 : len(typeExpr)-1])

		// Check for custom list codec
		if codec, ok := customCodecs[elemType]; ok && codec.EncodeListFunc != "" {
			return fmt.Sprintf("%s params.%s", codec.EncodeListFunc, fieldName)
		}

		// Special built-in list types
		switch elemType {
		case "BytesHex":
			return fmt.Sprintf("encodeBytesHexList params.%s", fieldName)
		case "Party":
			return fmt.Sprintf("encodePartyList params.%s", fieldName)
		}

		// Generic list - need element encoder
		elemEnc, _, _ := getCodecFuncs(elemType, customCodecs)
		return fmt.Sprintf("encodeList params.%s %s", fieldName, elemEnc)
	}

	// Optional types
	if strings.HasPrefix(typeExpr, "Optional ") {
		innerType := strings.TrimPrefix(typeExpr, "Optional ")
		innerEnc, _, _ := getCodecFuncs(innerType, customCodecs)
		return fmt.Sprintf("encodeOptional params.%s %s", fieldName, innerEnc)
	}

	// Standard case
	return fmt.Sprintf("%s params.%s", encodeFunc, fieldName)
}

func buildPayloadEncodeExpr(typeExpr, encodeFunc string, customCodecs map[string]CustomCodec) string {
	typeExpr = strings.TrimSpace(typeExpr)

	// Int requires fromSome wrapper
	if typeExpr == "Int" {
		return fmt.Sprintf("fromSome (%s payload)", encodeFunc)
	}

	return fmt.Sprintf("%s payload", encodeFunc)
}

func trackImport(module, encFunc, decFunc string, mcms map[string]bool, other map[string]map[string]bool) {
	if module == "" {
		return
	}

	switch module {
	case "MCMS.Codec":
		if encFunc != "" && encFunc != "encodeList" {
			mcms[encFunc] = true
		}
		if decFunc != "" && decFunc != "decodeList" {
			mcms[decFunc] = true
		}
		// Always need these for lists/optionals
		if encFunc == "encodeList" {
			mcms["encodeList"] = true
		}
		if decFunc == "decodeList" {
			mcms["decodeList"] = true
		}
	default:
		if other[module] == nil {
			other[module] = make(map[string]bool)
		}
		if encFunc != "" {
			other[module][encFunc] = true
		}
		if decFunc != "" {
			other[module][decFunc] = true
		}
	}
}

func mapKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
