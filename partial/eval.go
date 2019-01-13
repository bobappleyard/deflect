package partial

import (
	"go/ast"
	"go/token"
	"strconv"
)

type EvalScope interface {
	Lookup(name string) Value
}

func Eval(expr ast.Expr, scope EvalScope) []Value {
	switch expr := expr.(type) {
	case *ast.BasicLit:
		switch expr.Kind {
		case token.STRING:
			v, _ := strconv.Unquote(expr.Value)
			return []Value{String(v)}
		case token.INT:
			v, _ := strconv.ParseInt(expr.Value, 0, 64)
			return []Value{Int(v)}
		}

	case *ast.Ident:
		switch expr.Name {
		case "true":
			return []Value{True}
		case "false":
			return []Value{False}
		}
		return []Value{scope.Lookup(expr.Name)}

	case *ast.BinaryExpr:
		left := Eval(expr.X, scope)
		right := Eval(expr.Y, scope)
		return []Value{left[0].Op(expr.Op, right[0])}

	case *ast.CallExpr:
		f := Eval(expr.Fun, scope)[0]
		return f.Call(evalArgs(expr.Args, scope))

	case *ast.ParenExpr:
		inner := Eval(expr.X, scope)[0]
		if inner.Known() {
			return []Value{inner}
		}
		return []Value{&UnknownValue{&ast.ParenExpr{X: inner.Expr()}}}

	case *ast.SelectorExpr:
		recv := Eval(expr.X, scope)
		return []Value{recv[0].Member(expr.Sel.Name)}
	}
	return []Value{&UnknownValue{expr}}
}

func evalArgs(args []ast.Expr, scope EvalScope) []Value {
	if len(args) == 1 {
		return Eval(args[0], scope)
	}
	res := make([]Value, len(args))
	for i, a := range args {
		res[i] = Eval(a, scope)[0]
	}
	return res
}
