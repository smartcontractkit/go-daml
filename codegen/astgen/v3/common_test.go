package v3

import (
	"testing"

	daml "github.com/digital-asset/dazl-client/v8/go/api/com/digitalasset/daml/lf/archive/daml_lf_2"
	"github.com/smartcontractkit/go-daml/codegen/model"
	"github.com/stretchr/testify/require"
)

func TestHandleConType_StdlibTypes(t *testing.T) {
	codeGen := &codeGenAst{
		externalPackages: model.ExternalPackages{
			Packages: map[string]model.ExternalPackage{},
		},
		importedPackages: map[string]model.ExternalPackage{},
	}

	tests := []struct {
		name         string
		typeName     string
		expectedType model.DamlType
	}{
		{"Set via InternedStr", "Set", model.Set{}},
		{"RelTime via InternedStr", "RelTime", model.RelTime{}},
		{"Unknown via InternedStr", "SomeOtherType", model.Unknown{String: "SomeOtherType"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &daml.Package{
				InternedStrings: []string{"", "da-set-pkg-id", tt.typeName},
				InternedDottedNames: []*daml.InternedDottedName{
					{SegmentsInternedStr: []int32{2}},
				},
			}

			conType := &daml.Type_Con{
				Tycon: &daml.TypeConId{
					Module: &daml.ModuleId{
						PackageId: &daml.SelfOrImportedPackageId{
							Sum: &daml.SelfOrImportedPackageId_ImportedPackageIdInternedStr{
								ImportedPackageIdInternedStr: 1,
							},
						},
					},
					NameInternedDname: 0,
				},
			}

			result := codeGen.handleConType(pkg, conType)
			require.Equal(t, tt.expectedType, result)
		})
	}
}

func TestHandleConType_PackageImportId_StdlibTypes(t *testing.T) {
	codeGen := &codeGenAst{
		externalPackages: model.ExternalPackages{
			Packages: map[string]model.ExternalPackage{},
		},
		importedPackages: map[string]model.ExternalPackage{},
	}

	tests := []struct {
		name         string
		typeName     string
		expectedType model.DamlType
	}{
		{"Set via PackageImportId", "Set", model.Set{}},
		{"RelTime via PackageImportId", "RelTime", model.RelTime{}},
		{"Unknown via PackageImportId", "SomeOtherType", model.Unknown{String: "SomeOtherType"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &daml.Package{
				InternedStrings: []string{"", tt.typeName},
				InternedDottedNames: []*daml.InternedDottedName{
					{SegmentsInternedStr: []int32{1}},
				},
				ImportsSum: &daml.Package_PackageImports{
					PackageImports: &daml.PackageImports{
						ImportedPackages: []string{"some-stdlib-package-id"},
					},
				},
			}

			conType := &daml.Type_Con{
				Tycon: &daml.TypeConId{
					Module: &daml.ModuleId{
						PackageId: &daml.SelfOrImportedPackageId{
							Sum: &daml.SelfOrImportedPackageId_PackageImportId{
								PackageImportId: 0,
							},
						},
					},
					NameInternedDname: 0,
				},
			}

			result := codeGen.handleConType(pkg, conType)
			require.Equal(t, tt.expectedType, result)
		})
	}
}

func TestParseKeyExpressionV3(t *testing.T) {
	codeGen := &codeGenAst{}

	// Create a mock package with interned strings
	pkg := &daml.Package{
		InternedStrings: []string{
			"", // index 0 is usually empty
			"owner",
			"amount",
			"orderId",
			"customer",
		},
	}

	t.Run("Record projection key", func(t *testing.T) {
		// Create a key expression with a record projection (e.g., this.owner)
		key := &daml.DefTemplate_DefKey{
			KeyExpr: &daml.Expr{
				Sum: &daml.Expr_RecProj_{
					RecProj: &daml.Expr_RecProj{
						FieldInternedStr: 1, // "owner"
						Record: &daml.Expr{
							Sum: &daml.Expr_VarInternedStr{
								VarInternedStr: 1, // template variable
							},
						},
					},
				},
			},
		}

		fieldNames := codeGen.parseKeyExpression(pkg, key)
		require.Len(t, fieldNames, 2) // Both the field and the variable
		require.Contains(t, fieldNames, "owner")
	})

	t.Run("Record construction key", func(t *testing.T) {
		// Create a key expression with a record construction (composite key)
		key := &daml.DefTemplate_DefKey{
			KeyExpr: &daml.Expr{
				Sum: &daml.Expr_RecCon_{
					RecCon: &daml.Expr_RecCon{
						Fields: []*daml.FieldWithExpr{
							{
								FieldInternedStr: 1, // "owner"
								Expr: &daml.Expr{
									Sum: &daml.Expr_VarInternedStr{
										VarInternedStr: 1,
									},
								},
							},
							{
								FieldInternedStr: 3, // "orderId"
								Expr: &daml.Expr{
									Sum: &daml.Expr_VarInternedStr{
										VarInternedStr: 3,
									},
								},
							},
						},
					},
				},
			},
		}

		fieldNames := codeGen.parseKeyExpression(pkg, key)
		require.Len(t, fieldNames, 2)
		require.Contains(t, fieldNames, "owner")
		require.Contains(t, fieldNames, "orderId")
	})

	t.Run("Variable reference key", func(t *testing.T) {
		// Create a key expression with a simple variable reference
		key := &daml.DefTemplate_DefKey{
			KeyExpr: &daml.Expr{
				Sum: &daml.Expr_VarInternedStr{
					VarInternedStr: 4, // "customer"
				},
			},
		}

		fieldNames := codeGen.parseKeyExpression(pkg, key)
		require.Len(t, fieldNames, 1)
		require.Equal(t, "customer", fieldNames[0])
	})

	t.Run("Empty key expression", func(t *testing.T) {
		// Test with nil key
		key := &daml.DefTemplate_DefKey{
			KeyExpr: nil,
		}

		fieldNames := codeGen.parseKeyExpression(pkg, key)
		require.Len(t, fieldNames, 0)
	})
}

func TestHandleBuiltinType_GenMapPreservesTypeArgs(t *testing.T) {
	codeGen := &codeGenAst{}

	got := codeGen.handleBuiltinType(nil, &daml.Type_Builtin{
		Builtin: daml.BuiltinType_GENMAP,
		Args: []*daml.Type{
			{
				Sum: &daml.Type_Builtin_{
					Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_TEXT},
				},
			},
			{
				Sum: &daml.Type_Builtin_{
					Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_NUMERIC},
				},
			},
		},
	})

	require.Equal(t, model.GenMap{
		Key:   model.Text{},
		Value: model.Numeric{},
	}, got)
}

func TestExtractTapp_CurriedGenMapPreservesTypeArgs(t *testing.T) {
	codeGen := &codeGenAst{}

	curried := &daml.Type{
		Sum: &daml.Type_Tapp{
			Tapp: &daml.Type_TApp{
				Lhs: &daml.Type{
					Sum: &daml.Type_Tapp{
						Tapp: &daml.Type_TApp{
							Lhs: &daml.Type{
								Sum: &daml.Type_Builtin_{
									Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_GENMAP},
								},
							},
							Rhs: &daml.Type{
								Sum: &daml.Type_Builtin_{
									Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_TEXT},
								},
							},
						},
					},
				},
				Rhs: &daml.Type{
					Sum: &daml.Type_Builtin_{
						Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_BOOL},
					},
				},
			},
		},
	}

	got := codeGen.extractType(nil, curried)

	require.Equal(t, model.GenMap{
		Key:   model.Text{},
		Value: model.Bool{},
	}, got)
}

func TestHandleBuiltinType_TextMapPreservesTypeArg(t *testing.T) {
	codeGen := &codeGenAst{}

	got := codeGen.handleBuiltinType(nil, &daml.Type_Builtin{
		Builtin: daml.BuiltinType_TEXTMAP,
		Args: []*daml.Type{
			{
				Sum: &daml.Type_Builtin_{
					Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_NUMERIC},
				},
			},
		},
	})

	require.Equal(t, model.TextMap{
		Value: model.Numeric{},
	}, got)
}

func TestExtractTapp_CurriedTextMapPreservesTypeArg(t *testing.T) {
	codeGen := &codeGenAst{}

	curried := &daml.Type{
		Sum: &daml.Type_Tapp{
			Tapp: &daml.Type_TApp{
				Lhs: &daml.Type{
					Sum: &daml.Type_Builtin_{
						Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_TEXTMAP},
					},
				},
				Rhs: &daml.Type{
					Sum: &daml.Type_Builtin_{
						Builtin: &daml.Type_Builtin{Builtin: daml.BuiltinType_BOOL},
					},
				},
			},
		},
	}

	got := codeGen.extractType(nil, curried)

	require.Equal(t, model.TextMap{
		Value: model.Bool{},
	}, got)
}
