package partial

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestAnalysis(t *testing.T) {
	for _, test := range []struct {
		name string
		stmt ast.Stmt
		res  Point
	}{
		{
			"BasicBlock",
			&ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.Ident{Name: "f"},
							Args: []ast.Expr{
								&ast.Ident{Name: "x"},
							},
						},
					},
				},
			},
			&evalExpr{
				&ast.CallExpr{
					Fun: &ast.Ident{Name: "f"},
					Args: []ast.Expr{
						&ast.Ident{Name: "x"},
					},
				},
				&returnValues{nil},
			},
		},
		{
			"IfThen",
			&ast.IfStmt{
				Cond: &ast.Ident{Name: "cond"},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.Ident{Name: "result"},
							},
						},
					},
				},
			},
			&branch{
				condition: &ast.Ident{Name: "cond"},
				consequent: &returnValues{
					results: []ast.Expr{
						&ast.Ident{Name: "result"},
					},
				},
				antecedent: &returnValues{nil},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := new(analyzer).analyze(test.stmt, &returnValues{nil})
			if !reflect.DeepEqual(test.res, p) {
				t.Errorf("expected %#v, got %#v", test.res, p)
			}
		})
	}
}
