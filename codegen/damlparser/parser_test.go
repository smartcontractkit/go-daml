package damlparser

import (
	"testing"
)

func TestParseRecord(t *testing.T) {
	source := `-- | Types for LockReleaseTokenPool
module Test.LockReleaseTokenPoolTypes where

import DA.Crypto.Text (BytesHex)
import MCMS.RawInstanceAddress (RawInstanceAddress)

data ChainUpdate = ChainUpdate
    with
        remoteChainSelector : Numeric 0
        remotePools : [BytesHex]
        remoteTokenAddress : BytesHex
        inboundCCVs : [RawInstanceAddress]
        outboundCCVs : [RawInstanceAddress]
        minBlockConfirmations : Int
        inboundRateLimiter : RawInstanceAddress
        inboundCustomBlockConfirmationsRateLimiter : RawInstanceAddress
        outboundRateLimiter : RawInstanceAddress
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if module.ModuleName != "Test.LockReleaseTokenPoolTypes" {
		t.Errorf("ModuleName = %q, want %q", module.ModuleName, "Test.LockReleaseTokenPoolTypes")
	}

	if len(module.Imports) != 2 {
		t.Errorf("len(Imports) = %d, want 2", len(module.Imports))
	}

	if len(module.Records) != 1 {
		t.Fatalf("len(Records) = %d, want 1", len(module.Records))
	}

	rec := module.Records[0]
	if rec.Name != "ChainUpdate" {
		t.Errorf("Record.Name = %q, want %q", rec.Name, "ChainUpdate")
	}

	expectedFields := []struct {
		name     string
		typeExpr string
	}{
		{"remoteChainSelector", "Numeric 0"},
		{"remotePools", "[BytesHex]"},
		{"remoteTokenAddress", "BytesHex"},
		{"inboundCCVs", "[RawInstanceAddress]"},
		{"outboundCCVs", "[RawInstanceAddress]"},
		{"minBlockConfirmations", "Int"},
		{"inboundRateLimiter", "RawInstanceAddress"},
		{"inboundCustomBlockConfirmationsRateLimiter", "RawInstanceAddress"},
		{"outboundRateLimiter", "RawInstanceAddress"},
	}

	if len(rec.Fields) != len(expectedFields) {
		t.Fatalf("len(Fields) = %d, want %d", len(rec.Fields), len(expectedFields))
	}

	for i, expected := range expectedFields {
		if rec.Fields[i].Name != expected.name {
			t.Errorf("Field[%d].Name = %q, want %q", i, rec.Fields[i].Name, expected.name)
		}
		if rec.Fields[i].TypeExpr != expected.typeExpr {
			t.Errorf("Field[%d].TypeExpr = %q, want %q", i, rec.Fields[i].TypeExpr, expected.typeExpr)
		}
	}
}

func TestParseVariant(t *testing.T) {
	source := `module Test.LockReleaseTokenPoolTypes where

-- | Transfer timeout configuration
data TransferTimeout
    = Indefinite          -- Use maxTime (year 9999)
    | RelativeHours Int   -- Use now + N hours
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Variants) != 1 {
		t.Fatalf("len(Variants) = %d, want 1", len(module.Variants))
	}

	v := module.Variants[0]
	if v.Name != "TransferTimeout" {
		t.Errorf("Variant.Name = %q, want %q", v.Name, "TransferTimeout")
	}

	if len(v.Constructors) != 2 {
		t.Fatalf("len(Constructors) = %d, want 2", len(v.Constructors))
	}

	if v.Constructors[0].Name != "Indefinite" {
		t.Errorf("Constructors[0].Name = %q, want %q", v.Constructors[0].Name, "Indefinite")
	}
	if v.Constructors[0].PayloadType != "" {
		t.Errorf("Constructors[0].PayloadType = %q, want empty", v.Constructors[0].PayloadType)
	}

	if v.Constructors[1].Name != "RelativeHours" {
		t.Errorf("Constructors[1].Name = %q, want %q", v.Constructors[1].Name, "RelativeHours")
	}
	if v.Constructors[1].PayloadType != "Int" {
		t.Errorf("Constructors[1].PayloadType = %q, want %q", v.Constructors[1].PayloadType, "Int")
	}
}

func TestParseMultipleTypes(t *testing.T) {
	source := `module Test.PricingTypes where

import Splice.Api.Token.HoldingV1

data PriceUpdates = PriceUpdates
    with
        tokenPriceUpdates : [TokenPriceUpdate]
        gasPriceUpdates : [GasPriceUpdate]
    deriving (Eq, Show)

data TokenPriceUpdate = TokenPriceUpdate
    with
        instrumentId : Splice.Api.Token.HoldingV1.InstrumentId
        usdPerToken : Numeric 0
    deriving (Eq, Show)

data GasPriceUpdate = GasPriceUpdate
    with
        routeSelector : Numeric 0
        usdPerUnitGas : Numeric 0
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Records) != 3 {
		t.Errorf("len(Records) = %d, want 3", len(module.Records))
	}

	names := []string{"PriceUpdates", "TokenPriceUpdate", "GasPriceUpdate"}
	for i, name := range names {
		if module.Records[i].Name != name {
			t.Errorf("Records[%d].Name = %q, want %q", i, module.Records[i].Name, name)
		}
	}
}

func TestParseOptionalFields(t *testing.T) {
	source := `module Test.RegistryTypes where

data RegistrationConfig = RegistrationConfig
    with
        instrumentId : Splice.Api.Token.HoldingV1.InstrumentId
        owner : Optional Party
        pendingOwner : Optional Party
        serviceEndpoint : Optional ServiceEndpoint
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Records) != 1 {
		t.Fatalf("len(Records) = %d, want 1", len(module.Records))
	}

	rec := module.Records[0]
	expectedFields := []struct {
		name     string
		typeExpr string
	}{
		{"instrumentId", "Splice.Api.Token.HoldingV1.InstrumentId"},
		{"owner", "Optional Party"},
		{"pendingOwner", "Optional Party"},
		{"serviceEndpoint", "Optional ServiceEndpoint"},
	}

	if len(rec.Fields) != len(expectedFields) {
		t.Fatalf("len(Fields) = %d, want %d", len(rec.Fields), len(expectedFields))
	}

	for i, expected := range expectedFields {
		if rec.Fields[i].Name != expected.name {
			t.Errorf("Field[%d].Name = %q, want %q", i, rec.Fields[i].Name, expected.name)
		}
		if rec.Fields[i].TypeExpr != expected.typeExpr {
			t.Errorf("Field[%d].TypeExpr = %q, want %q", i, rec.Fields[i].TypeExpr, expected.typeExpr)
		}
	}
}

func TestParseQualifiedImports(t *testing.T) {
	source := `module Test.CommitteeVerifierTypes where

import DA.Crypto.Text (BytesHex)
import MCMS.RawInstanceAddress qualified as RawInstanceAddress

data SetDepsParams = SetDepsParams
    with
        rmnRemote : Optional RawInstanceAddress.RawInstanceAddress
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Imports) != 2 {
		t.Errorf("len(Imports) = %d, want 2", len(module.Imports))
	}

	// The parser should capture the module names
	if module.Imports[0] != "DA.Crypto.Text" {
		t.Errorf("Imports[0] = %q, want %q", module.Imports[0], "DA.Crypto.Text")
	}
	if module.Imports[1] != "MCMS.RawInstanceAddress" {
		t.Errorf("Imports[1] = %q, want %q", module.Imports[1], "MCMS.RawInstanceAddress")
	}
}

func TestParseWithInlineComment(t *testing.T) {
	source := `module Test.GlobalConfigTypes where

data RouteConfigArgs = RouteConfigArgs
    with
        routeSelector : Numeric 0               -- Route selector
        isEnabled : Bool                        -- Flag whether enabled
        addressBytesLength : Int                -- Length in bytes
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Records) != 1 {
		t.Fatalf("len(Records) = %d, want 1", len(module.Records))
	}

	rec := module.Records[0]
	if len(rec.Fields) != 3 {
		t.Fatalf("len(Fields) = %d, want 3", len(rec.Fields))
	}

	// Comments should be stripped from type expressions
	if rec.Fields[0].TypeExpr != "Numeric 0" {
		t.Errorf("Fields[0].TypeExpr = %q, want %q", rec.Fields[0].TypeExpr, "Numeric 0")
	}
	if rec.Fields[1].TypeExpr != "Bool" {
		t.Errorf("Fields[1].TypeExpr = %q, want %q", rec.Fields[1].TypeExpr, "Bool")
	}
	if rec.Fields[2].TypeExpr != "Int" {
		t.Errorf("Fields[2].TypeExpr = %q, want %q", rec.Fields[2].TypeExpr, "Int")
	}
}

func TestParseNestedRecordType(t *testing.T) {
	source := `module Test.ConfigurationTypes where

data RouteConfigArgs = RouteConfigArgs
    with
        routeSelector : Numeric 0
        routeConfig : RouteConfig
    deriving (Eq, Show)

data RouteConfig = RouteConfig
    with
        isEnabled : Bool
        maxDataBytes : Int
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Records) != 2 {
		t.Fatalf("len(Records) = %d, want 2", len(module.Records))
	}

	// First record should have RouteConfig as a field type
	if module.Records[0].Fields[1].TypeExpr != "RouteConfig" {
		t.Errorf("Fields[1].TypeExpr = %q, want %q", module.Records[0].Fields[1].TypeExpr, "RouteConfig")
	}
}

func TestParseListTypes(t *testing.T) {
	source := `module Test where

data Config = Config
    with
        parties : [Party]
        items : [SomeType]
        addresses : [RawInstanceAddress]
    deriving (Eq, Show)
`

	module, err := ParseString(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(module.Records) != 1 {
		t.Fatalf("len(Records) = %d, want 1", len(module.Records))
	}

	expectedTypes := []string{"[Party]", "[SomeType]", "[RawInstanceAddress]"}
	for i, expected := range expectedTypes {
		if module.Records[0].Fields[i].TypeExpr != expected {
			t.Errorf("Fields[%d].TypeExpr = %q, want %q", i, module.Records[0].Fields[i].TypeExpr, expected)
		}
	}
}
