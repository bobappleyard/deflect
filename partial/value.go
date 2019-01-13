package partial

import (
	"go/ast"
	"go/token"
	"strconv"
)

type Value interface {
	Expr() ast.Expr
	Matches(w Value) bool
	Known() bool
	KnownRef() bool
	Op(op token.Token, w Value) Value
	Member(name string) Value
	Call(args []Value) []Value
	Update(w Value)
}

type UnknownValue struct {
	expr ast.Expr
}

func (v *UnknownValue) Expr() ast.Expr                   { return v.expr }
func (v *UnknownValue) Matches(w Value) bool             { return !w.Known() }
func (v *UnknownValue) Known() bool                      { return false }
func (v *UnknownValue) KnownRef() bool                   { return false }
func (v *UnknownValue) Op(op token.Token, w Value) Value { return opExpr(op, v, w) }
func (v *UnknownValue) Member(name string) Value         { return selExpr(v, name) }
func (v *UnknownValue) Call(args []Value) []Value        { return []Value{callExpr(v, args)} }
func (v *UnknownValue) Update(w Value)                   {}

type baseValue struct{}

func (baseValue) Matches(Value) bool               { return false }
func (baseValue) Known() bool                      { return true }
func (baseValue) KnownRef() bool                   { return false }
func (baseValue) Op(op token.Token, w Value) Value { panic("invalid operation") }
func (baseValue) Member(name string) Value         { panic("invalid operation") }
func (baseValue) Call(args []Value) []Value        { panic("invalid operation") }
func (baseValue) Update(w Value)                   {}

type IntValue struct {
	baseValue
	value int64
}

func Int(v int64) *IntValue {
	return &IntValue{value: v}
}

func (v *IntValue) Expr() ast.Expr {
	return &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(v.value, 10)}
}

func (v *IntValue) Matches(w Value) bool {
	if w, ok := w.(*IntValue); ok {
		return w.value == v.value
	}
	return false
}

func (v *IntValue) Op(op token.Token, w Value) Value {
	if w, ok := w.(*IntValue); ok {
		vx, wx := v.value, w.value
		switch op {
		case token.ADD:
			return Int(vx + wx)
		case token.MUL:
			return Int(vx * wx)
		case token.QUO:
			return Int(vx / wx)
		case token.REM:
			return Int(vx % wx)
		case token.AND:
			return Int(vx | wx)
		case token.OR:
			return Int(vx | wx)
		case token.XOR:
			return Int(vx ^ wx)
		case token.SHL:
			return Int(vx << uint(wx))
		case token.SHR:
			return Int(vx >> uint(wx))
		case token.AND_NOT:
			return Int(vx &^ wx)

		case token.EQL:
			return Bool(vx == wx)
		case token.NEQ:
		}
	}
	return opExpr(op, v, w)
}

type StringValue struct {
	baseValue
	value string
}

func String(v string) *StringValue {
	return &StringValue{value: v}
}

func (v *StringValue) Matches(w Value) bool {
	if w, ok := w.(*StringValue); ok {
		return w.value == v.value
	}
	return false
}

func (v *StringValue) Expr() ast.Expr {
	return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(v.value)}
}

func (v *StringValue) Op(op token.Token, w Value) Value {
	if w, ok := w.(*StringValue); ok {
		vx, wx := v.value, w.value
		switch op {
		case token.ADD:
			return String(vx + wx)
		}
	}
	return opExpr(op, v, w)
}

type BoolValue struct {
	baseValue
	value bool
}

var (
	True  = &BoolValue{value: true}
	False = &BoolValue{value: false}
)

func Bool(value bool) *BoolValue {
	if value {
		return True
	}
	return False
}

func (v *BoolValue) Expr() ast.Expr {
	return &ast.Ident{Name: strconv.FormatBool(v.value)}
}

func (v *BoolValue) Matches(w Value) bool {
	if w, ok := w.(*BoolValue); ok {
		return w.value == v.value
	}
	return false
}

func (v *BoolValue) Op(op token.Token, w Value) Value {
	if w, ok := w.(*BoolValue); ok {
		vx, wx := v.value, w.value
		switch op {
		case token.LAND:
			return Bool(vx && wx)
		case token.LOR:
			return Bool(vx || wx)
		}
	}
	return opExpr(op, v, w)
}
