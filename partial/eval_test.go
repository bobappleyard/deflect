package partial

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

type testScope struct {
	parent *testScope
	values map[string]Value
	syms   int
}

func (s *testScope) Lookup(name string) Value {
	if v, ok := s.values[name]; ok {
		return v
	}
	return s.parent.Lookup(name)
}

func (s *testScope) DefineValue(name string, value Value) {
	if s.values == nil {
		s.values = map[string]Value{}
	}
	s.values[name] = value
}

func defineUnknown(scope *testScope, name string) {
	scope.DefineValue(name, &UnknownValue{&ast.Ident{Name: name}})
}

func nodeString(n ast.Node) string {
	var buf bytes.Buffer
	format.Node(&buf, token.NewFileSet(), n)
	return buf.String()
}

func stringNode(s string) ast.Node {
	n, err := parser.ParseExpr(s)
	if err != nil {
		panic(err)
	}
	return n
}

func TestEval(t *testing.T) {
	scope := &testScope{}
	scope.DefineValue("a", Int(1))
	defineUnknown(scope, "x")
	defineUnknown(scope, "y")
	defineUnknown(scope, "f")
	defineUnknown(scope, "recv")
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
			expected := nodeString(stringNode(test.out))
			out := nodeString(Eval(in, scope)[0].Expr())
			if out != expected {
				t.Errorf("\nexpected\n\t%s\ngot\n\t%s", expected, out)
			}
		})
	}
}
