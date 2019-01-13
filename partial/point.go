package partial

import (
	"go/ast"
	"go/token"
)

type Point interface {
	Successors(scope ExecScope) []State
}

type State struct {
	point Point
	scope ExecScope
}

type unsupported struct {
	stmt ast.Stmt
}

func (p *unsupported) Successors(scope ExecScope) []State {
	return nil
}

type evalExpr struct {
	expr ast.Expr
	cont Point
}

func (p *evalExpr) Successors(scope ExecScope) []State {
	return []State{{p.cont, scope}}
}

type returnValues struct {
	results []ast.Expr
}

func (p *returnValues) Successors(scope ExecScope) []State {
	return nil
}

type branch struct {
	condition              ast.Expr
	consequent, antecedent Point
}

func (p *branch) Successors(scope ExecScope) []State {
	v := Eval(p.condition, scope)[0]
	if v.Matches(True) {
		return []State{{p.consequent, bindKnownValues(p.condition, scope)}}
	}
	if v.Matches(False) {
		return []State{{p.antecedent, scope}}
	}
	return []State{
		{p.consequent, bindKnownValues(p.condition, scope)},
		{p.antecedent, scope},
	}
}

func bindKnownValues(e ast.Expr, scope ExecScope) ExecScope {
	bin, ok := e.(*ast.BinaryExpr)
	if !ok {
		return scope
	}
	switch bin.Op {
	case token.EQL:
		if id, ok := bin.X.(*ast.Ident); ok {
			return scope.Bind(id.Name, Eval(bin.Y, scope)[0])
		}
	case token.AND:
		return bindKnownValues(bin.X, bindKnownValues(bin.Y, scope))
	}
	return scope
}

type assign struct {
	lhs, rhs []ast.Expr
	define   bool
	cont     Point
}

func (p *assign) Successors(scope ExecScope) []State {
	rhs := evalArgs(p.lhs, scope)
	for i, lhs := range p.lhs {
		rhs := rhs[i]
		if ref, ok := lhs.(*ast.Ident); ok {
			scope = scope.Bind(ref.Name, rhs)
		}
	}
	return []State{{p.cont, scope}}
}
