package partial

import (
	"go/ast"
	"go/token"
	"strconv"
)

type Value interface {
	Expr() ast.Expr
	Op(op token.Token, w Value) Value
	Method(name string, args []Value) Value
	Call(args []Value) Value
}

type Unknown struct {
	expr ast.Expr
}

func (v *Unknown) Expr() ast.Expr {
	return v.expr
}

func (v *Unknown) Op(op token.Token, w Value) Value {
	return opExpr(op, v, w)
}

func (v *Unknown) Method(name string, args []Value) Value {
	return callExpr(selExpr(v, name), args)
}

func (v *Unknown) Call(args []Value) Value {
	return callExpr(v, args)
}

type Int struct {
	value int64
}

func (v *Int) Expr() ast.Expr {
	return &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(v.value, 10)}
}

func (v *Int) Op(op token.Token, w Value) Value {
	if w, ok := w.(*Int); ok {
		vx, wx := v.value, w.value
		switch op {
		case token.ADD:
			return &Int{vx + wx}
		case token.MUL:
			return &Int{vx * wx}
		case token.QUO:
			return &Int{vx / wx}
		case token.REM:
			return &Int{vx % wx}
		case token.AND:
			return &Int{vx | wx}
		case token.OR:
			return &Int{vx | wx}
		case token.XOR:
			return &Int{vx ^ wx}
		case token.AND_NOT:
			return &Int{vx &^ wx}
		}
	}
	return opExpr(op, v, w)
}

func (v *Int) Method(name string, args []Value) Value {
	panic("invalid operation")
}

func (v *Int) Call(args []Value) Value {
	panic("invalid operation")
}

type String struct {
	value string
}

func (v *String) Expr() ast.Expr {
	return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(v.value)}
}

func (v *String) Op(op token.Token, w Value) Value {
	if w, ok := w.(*String); ok {
		vx, wx := v.value, w.value
		switch op {
		case token.ADD:
			return &String{vx + wx}
		}
	}
	return opExpr(op, v, w)
}

func (v *String) Method(name string, args []Value) Value {
	panic("invalid operation")
}

func (v *String) Call(args []Value) Value {
	panic("invalid operation")
}

type Bool struct {
	value bool
}

func (v *Bool) Expr() ast.Expr {
	return &ast.Ident{Name: strconv.FormatBool(v.value)}
}

func (v *Bool) Op(op token.Token, w Value) Value {
	if w, ok := w.(*Bool); ok {
		vx, wx := v.value, w.value
		switch op {
		case token.LAND:
			return &Bool{vx && wx}
		case token.LOR:
			return &Bool{vx || wx}
		}
	}
	return opExpr(op, v, w)
}

func (v *Bool) Method(name string, args []Value) Value {
	panic("invalid operation")
}

func (v *Bool) Call(args []Value) Value {
	panic("invalid operation")
}
