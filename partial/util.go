package partial

import (
	"go/ast"
	"go/token"
)

type hasExpr interface {
	Expr() ast.Expr
}

func opExpr(op token.Token, x hasExpr, y hasExpr) Value {
	return &UnknownValue{&ast.BinaryExpr{Op: op, X: x.Expr(), Y: y.Expr()}}
}

func callExpr(f hasExpr, args []Value) Value {
	return &UnknownValue{&ast.CallExpr{Fun: f.Expr(), Args: argsExpr(args)}}
}

func selExpr(recv hasExpr, name string) Value {
	return &UnknownValue{&ast.SelectorExpr{X: recv.Expr(), Sel: &ast.Ident{Name: name}}}
}

func argsExpr(args []Value) []ast.Expr {
	res := make([]ast.Expr, len(args))
	for i, a := range args {
		res[i] = a.Expr()
	}
	return res
}
