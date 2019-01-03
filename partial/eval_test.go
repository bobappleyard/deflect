package partial

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

type testScope map[string]Value

func (s testScope) LookupValue(name string) Value {
	if v, ok := s[name]; ok {
		return v
	}
	return &Unknown{&ast.Ident{Name: name}}
}

func TestEval(t *testing.T) {
	scope := testScope{
		"a": &Int{1},
	}
	for _, test := range []struct {
		name, in, out string
	}{
		{
			"Ident",
			`x`,
			`x`,
		},
		{
			"Lookup",
			`a`,
			`1`,
		},
		{
			"Binary",
			`x+y`,
			`x+y`,
		},
		{
			"Call",
			`f(x)`,
			`f(x)`,
		},
		{
			"Method",
			`recv.method(x,y)`,
			`recv.method(x,y)`,
		},
		{
			"MethodKnown",
			`recv.method(x,a)`,
			`recv.method(x,1)`,
		},
		{
			"Int",
			`1`,
			`1`,
		},
		{
			"Double",
			`2*2`,
			`4`,
		},
		{
			"AndNot",
			`0xf &^ 3`,
			`12`,
		},
		{
			"AddIdent",
			`2*2+x`,
			`4+x`,
		},
		{
			"AddKnown",
			`2*2+a`,
			`5`,
		},
		{
			"Paren",
			`2*(2+a)`,
			`6`,
		},
		{
			"ParenIdent",
			`2*(2+x)`,
			`2*(2+x)`,
		},
		{
			"ParenKnown",
			`x*(2+a)`,
			`x*3`,
		},
		{
			"StringLit",
			`"hello"`,
			`"hello"`,
		},
		{
			"StringCat",
			`"hello" + " world"`,
			`"hello world"`,
		},
		{
			"Bool",
			`true`,
			`true`,
		},
		{
			"BoolOps",
			`true||false`,
			`true`,
		},
		{
			"BoolOps",
			`true&&false`,
			`false`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			in, _ := parser.ParseExpr(test.in)
			expected, _ := parser.ParseExpr(test.out)
			out := Eval(in, scope).Expr()
			var ob, eb bytes.Buffer
			fset := token.NewFileSet()
			format.Node(&ob, fset, out)
			format.Node(&eb, fset, expected)
			if ob.String() != eb.String() {
				t.Errorf("\nexpected\n\t%s\ngot\n\t%s", eb.String(), ob.String())
			}
		})
	}
}
