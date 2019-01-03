package partial

import (
	"go/ast"
	"go/token"
	"strconv"
)

type Scope interface {
	LookupValue(name string) Value
}

func Eval(expr ast.Expr, scope Scope) Value {
	switch expr := expr.(type) {
	case *ast.BasicLit:
		switch expr.Kind {
		case token.STRING:
			v, _ := strconv.Unquote(expr.Value)
			return &String{v}
		case token.INT:
			v, _ := strconv.ParseInt(expr.Value, 0, 64)
			return &Int{v}
		}

	case *ast.Ident:
		switch expr.Name {
		case "true":
			return &Bool{true}
		case "false":
			return &Bool{false}
		}
		return scope.LookupValue(expr.Name)

	case *ast.BinaryExpr:
		left := Eval(expr.X, scope)
		right := Eval(expr.Y, scope)
		return left.Op(expr.Op, right)

	case *ast.CallExpr:
		args := evalArgs(expr.Args, scope)
		if f, ok := expr.Fun.(*ast.SelectorExpr); ok {
			recv := Eval(f.X, scope)
			return recv.Method(f.Sel.Name, args)
		}
		f := Eval(expr.Fun, scope)
		return f.Call(args)

	case *ast.ParenExpr:
		inner := Eval(expr.X, scope)
		if _, ok := inner.(*Unknown); ok {
			return &Unknown{&ast.ParenExpr{X: inner.Expr()}}
		}
		return inner
	}
	return &Unknown{expr}
}

func evalArgs(args []ast.Expr, scope Scope) []Value {
	res := make([]Value, len(args))
	for i, a := range args {
		res[i] = Eval(a, scope)
	}
	return res
}
