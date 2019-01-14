package partial

import (
	"go/ast"
	"go/token"
)

type ExecScope interface {
	Lookup(name string) Value
	Bind(name string, value Value) ExecScope
}

type analyzer struct {
	next, out Point
	labels    map[string]Point
}

func (a *analyzer) analyze(stmt ast.Stmt, cont Point) Point {
	switch stmt := stmt.(type) {
	case *ast.EmptyStmt:
		return cont

	case *ast.ExprStmt:
		return &evalExpr{stmt.X, cont}

	case *ast.AssignStmt:
		return &assign{stmt.Lhs, stmt.Rhs, stmt.Tok == token.DEFINE, cont}

	case *ast.BlockStmt:
		for i := len(stmt.List) - 1; i >= 0; i-- {
			cont = a.analyze(stmt.List[i], cont)
		}
		return cont

	case *ast.ReturnStmt:
		return &returnValues{stmt.Results}

	case *ast.IfStmt:
		consequent := a.analyze(stmt.Body, cont)
		antecedent := cont
		if stmt.Else != nil {
			antecedent = a.analyze(stmt.Else, cont)
		}
		res := &branch{stmt.Cond, consequent, antecedent}
		if stmt.Init != nil {
			return a.analyze(stmt.Init, res)
		}
		return res

	case *ast.ForStmt:
		loop := &branch{stmt.Cond, nil, cont}
		body := a.inLoop(loop, cont).analyze(stmt.Body, a.analyze(stmt.Post, loop))
		loop.consequent = body
		return a.analyze(stmt.Init, loop)

	case *ast.BranchStmt:
		switch stmt.Tok {
		case token.GOTO:
			return a.labels[stmt.Label.Name]
		case token.BREAK:
			return a.out
		case token.CONTINUE:
			return a.next
		}

	}
	return &unsupported{stmt}
}

func (a *analyzer) inLoop(next, out Point) *analyzer {
	return &analyzer{next, out, a.labels}
}
