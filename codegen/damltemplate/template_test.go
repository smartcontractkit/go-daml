package damltemplate

import (
	"strings"
	"testing"

	"github.com/smartcontractkit/go-daml/codegen/damlparser"
)

func TestGenerateRecord(t *testing.T) {
	module := &damlparser.ParsedModule{
		ModuleName: "Test.TestTypes",
		Records: []damlparser.ParsedRecord{
			{
				Name: "ChainUpdate",
				Fields: []damlparser.ParsedField{
					{Name: "remoteChainSelector", TypeExpr: "Numeric 0"},
					{Name: "allowed", TypeExpr: "Bool"},
					{Name: "minBlockConfirmations", TypeExpr: "Int"},
				},
			},
		},
	}

	config := DamlCodecConfig{
		ModuleName:  "Test.TestCodec",
		TypesModule: "Test.TestTypes",
	}

	output, err := Generate(module, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check module declaration
	if !strings.Contains(output, "module Test.TestCodec where") {
		t.Error("Missing module declaration")
	}

	// Check import
	if !strings.Contains(output, "import Test.TestTypes (ChainUpdate(..))") {
		t.Error("Missing types import")
	}

	// Check encode function
	if !strings.Contains(output, "encodeChainUpdate : ChainUpdate -> BytesHex") {
		t.Error("Missing encode function signature")
	}

	// Check field encoding
	if !strings.Contains(output, "encodeNumeric0 params.remoteChainSelector") {
		t.Error("Missing Numeric 0 encoding")
	}
	if !strings.Contains(output, "encodeBool params.allowed") {
		t.Error("Missing Bool encoding")
	}
	if !strings.Contains(output, "fromSome (encodeInt64 params.minBlockConfirmations)") {
		t.Error("Missing Int encoding with fromSome wrapper")
	}

	// Check decode function
	if !strings.Contains(output, "decodeChainUpdateAt : BytesHex -> Int -> Optional (ChainUpdate, Int)") {
		t.Error("Missing decode function signature")
	}
}

func TestGenerateVariant(t *testing.T) {
	module := &damlparser.ParsedModule{
		ModuleName: "Test.TestTypes",
		Variants: []damlparser.ParsedVariant{
			{
				Name: "TransferTimeout",
				Constructors: []damlparser.ParsedConstructor{
					{Name: "Indefinite", PayloadType: ""},
					{Name: "RelativeHours", PayloadType: "Int"},
				},
			},
		},
	}

	config := DamlCodecConfig{
		ModuleName:  "Test.TestCodec",
		TypesModule: "Test.TestTypes",
		VariantTagByteMap: map[string]map[string]int{
			"TransferTimeout": {
				"Indefinite":    0,
				"RelativeHours": 1,
			},
		},
	}

	output, err := Generate(module, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check encode function
	if !strings.Contains(output, "encodeTransferTimeout : TransferTimeout -> BytesHex") {
		t.Error("Missing variant encode signature")
	}

	// Check case expression
	if !strings.Contains(output, "case variant of") {
		t.Error("Missing case expression")
	}

	// Check constructors
	if !strings.Contains(output, "Indefinite -> fromSome (encodeUint8 0)") {
		t.Error("Missing Indefinite encoding")
	}
	if !strings.Contains(output, "RelativeHours payload -> fromSome (encodeUint8 1)") {
		t.Error("Missing RelativeHours encoding")
	}
}

func TestGenerateWithCustomTypes(t *testing.T) {
	module := &damlparser.ParsedModule{
		ModuleName: "Test.TestTypes",
		Records: []damlparser.ParsedRecord{
			{
				Name: "PoolConfig",
				Fields: []damlparser.ParsedField{
					{Name: "rateLimiter", TypeExpr: "RawInstanceAddress"},
					{Name: "ccvs", TypeExpr: "[RawInstanceAddress]"},
				},
			},
		},
	}

	config := DamlCodecConfig{
		ModuleName:  "Test.TestCodec",
		TypesModule: "Test.TestTypes",
		CustomTypeCodecs: map[string]CustomCodec{
			"RawInstanceAddress": {
				EncodeFunc:     "encodeRawInstanceAddress",
				DecodeFunc:     "decodeRawInstanceAddressAt",
				ImportModule:   "Test.Codec",
				EncodeListFunc: "encodeRawInstanceAddressList",
				DecodeListFunc: "decodeRawInstanceAddressList",
			},
		},
	}

	output, err := Generate(module, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check Test.Codec import
	if !strings.Contains(output, "import Test.Codec") {
		t.Error("Missing Test.Codec import")
	}

	// Check custom type encoding
	if !strings.Contains(output, "encodeRawInstanceAddress params.rateLimiter") {
		t.Error("Missing custom type encoding")
	}

	// Check list encoding with custom list codec
	if !strings.Contains(output, "encodeRawInstanceAddressList params.ccvs") {
		t.Error("Missing custom list encoding")
	}
}

// TestGenerateGenericListAndCustomDecode covers the three bug fixes:
// 1. Custom DecodeFunc without "At" suffix (InstrumentId uses decodeInstrumentId, not decodeInstrumentIdAt)
// 2. encodeList/decodeList imports are present for generic [T] list fields
// 3. List fields use "decodeList encoded offset decodeXAt" not "decodeXAt encoded offset"
func TestGenerateGenericListAndCustomDecode(t *testing.T) {
	module := &damlparser.ParsedModule{
		ModuleName: "Test.PricingTypes",
		Records: []damlparser.ParsedRecord{
			{
				Name: "PriceUpdates",
				Fields: []damlparser.ParsedField{
					{Name: "tokenPriceUpdates", TypeExpr: "[TokenPriceUpdate]"},
					{Name: "gasPriceUpdates", TypeExpr: "[GasPriceUpdate]"},
				},
			},
			{
				Name: "TokenPriceUpdate",
				Fields: []damlparser.ParsedField{
					{Name: "instrumentId", TypeExpr: "Splice.Api.Token.HoldingV1.InstrumentId"},
					{Name: "usdPerToken", TypeExpr: "Numeric 0"},
				},
			},
			{
				Name: "GasPriceUpdate",
				Fields: []damlparser.ParsedField{
					{Name: "routeSelector", TypeExpr: "Numeric 0"},
					{Name: "usdPerUnitGas", TypeExpr: "Numeric 0"},
				},
			},
		},
	}

	config := DamlCodecConfig{
		ModuleName:  "Test.PricingCodecGen",
		TypesModule: "Test.PricingTypes",
		CustomTypeCodecs: map[string]CustomCodec{
			"InstrumentId": {
				EncodeFunc:     "encodeInstrumentId",
				DecodeFunc:     "decodeInstrumentId", // NO "At" suffix
				ImportModule:   "Test.Codec",
				EncodeListFunc: "encodeInstrumentIdList",
				DecodeListFunc: "decodeInstrumentIdList",
			},
		},
	}

	output, err := Generate(module, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Bug 1: Custom DecodeFunc without "At" suffix
	if !strings.Contains(output, "decodeInstrumentId encoded offset") {
		t.Error("Bug 1 regression: should use decodeInstrumentId (no At suffix) from CustomCodec.DecodeFunc")
	}
	if strings.Contains(output, "decodeInstrumentIdAt") {
		t.Error("Bug 1 regression: decodeInstrumentIdAt should NOT appear (custom codec specifies no At suffix)")
	}

	// Bug 2: encodeList/decodeList imports
	if !strings.Contains(output, "encodeList") {
		t.Error("Bug 2 regression: encodeList must be imported for generic list fields")
	}
	if !strings.Contains(output, "decodeList") {
		t.Error("Bug 2 regression: decodeList must be imported for generic list fields")
	}

	// Bug 3: List fields use decodeList wrapper
	if !strings.Contains(output, "decodeList encoded offset decodeTokenPriceUpdateAt") {
		t.Error("Bug 3 regression: [TokenPriceUpdate] should decode as 'decodeList encoded offset decodeTokenPriceUpdateAt'")
	}
	if !strings.Contains(output, "decodeList encoded offset decodeGasPriceUpdateAt") {
		t.Error("Bug 3 regression: [GasPriceUpdate] should decode as 'decodeList encoded offset decodeGasPriceUpdateAt'")
	}

	// Encode side: generic list uses encodeList with element encoder
	if !strings.Contains(output, "encodeList params.tokenPriceUpdates encodeTokenPriceUpdate") {
		t.Error("List encode should use 'encodeList params.tokenPriceUpdates encodeTokenPriceUpdate'")
	}

	// InstrumentId import comes from Test.Codec
	if !strings.Contains(output, "import Test.Codec") {
		t.Error("Missing Test.Codec import for InstrumentId custom codec")
	}
}

func TestGenerateWithOptional(t *testing.T) {
	module := &damlparser.ParsedModule{
		ModuleName: "Test.TestTypes",
		Records: []damlparser.ParsedRecord{
			{
				Name: "Config",
				Fields: []damlparser.ParsedField{
					{Name: "admin", TypeExpr: "Optional Party"},
				},
			},
		},
	}

	config := DamlCodecConfig{
		ModuleName:  "Test.TestCodec",
		TypesModule: "Test.TestTypes",
	}

	output, err := Generate(module, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check Optional encoding
	if !strings.Contains(output, "encodeOptional params.admin encodeParty") {
		t.Error("Missing Optional encoding")
	}
}
