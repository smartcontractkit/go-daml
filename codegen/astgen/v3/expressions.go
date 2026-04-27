package v3

import (
	"fmt"

	daml "github.com/digital-asset/dazl-client/v8/go/api/com/digitalasset/daml/lf/archive/daml_lf_2"
	"github.com/smartcontractkit/go-daml/codegen/model"
)

func (c *codeGenAst) extractExpression(pkg *daml.Package, expr *daml.Expr) (model.DamlExpression, error) {
	switch v := expr.GetSum().(type) {
	case *daml.Expr_InternedExpr:
		prim := pkg.InternedExprs[v.InternedExpr]
		if prim == nil {
			return nil, fmt.Errorf("unknown InternedExpr: %d", v.InternedExpr)
		}
		return c.extractExpression(pkg, prim)
	case *daml.Expr_BuiltinLit:
		return c.extractExpBuiltinLit(pkg, v.BuiltinLit)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr.GetSum())
	}
}

func (c *codeGenAst) extractExpBuiltinLit(pkg *daml.Package, expr *daml.BuiltinLit) (model.DamlExpression, error) {
	switch v := expr.GetSum().(type) {
	case *daml.BuiltinLit_Int64:
		return model.Int64Literal{
			Value: v.Int64,
		}, nil
	case *daml.BuiltinLit_NumericInternedStr:
		value := pkg.InternedStrings[v.NumericInternedStr]
		return model.NumericLiteral{
			Value: value,
		}, nil
	case *daml.BuiltinLit_TextInternedStr:
		value := pkg.InternedStrings[v.TextInternedStr]
		return model.TextLiteral{
			Value: value,
		}, nil
	default:
		// Can't handle timestamps and dates, as they aren't Go constants
		return nil, fmt.Errorf("unsupported BuiltinLit type: %T", expr.GetSum())
	}
}
