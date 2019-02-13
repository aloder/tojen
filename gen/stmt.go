package gen

import (
	"go/ast"
	"go/token"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func stmt(s ast.Stmt) jen.Code {
	switch t := s.(type) {
	case *ast.BadStmt:
	case *ast.DeclStmt:
		return declStmt(t)
	case *ast.GoStmt:
		return goStmt(t)
	case *ast.EmptyStmt:
		return emptyStmt(t)
	case *ast.LabeledStmt:
		return labeledStmt(t)
	case *ast.ExprStmt:
		return exprStmt(t)
	case *ast.SendStmt:
		return sendStmt(t)
	case *ast.IncDecStmt:
		return incDecStmt(t)
	case *ast.AssignStmt:
		return assignStmt(t)
	case *ast.ReturnStmt:
		return returnStmt(t)
	case *ast.BranchStmt:
		return branchStmt(t)
	case *ast.BlockStmt:
		return blockStmt(t)
	case *ast.IfStmt:
		return ifStmt(t)
	case *ast.CaseClause:
		return caseClause(t)
	case *ast.SwitchStmt:
		return switchStmt(t)
	case *ast.TypeSwitchStmt:
		return typeSwitchStmt(t)
	case *ast.CommClause:
		return commClause(t)
	case *ast.SelectStmt:
		return selectStmt(t)
	case *ast.ForStmt:
		return forStmt(t)
	case *ast.RangeStmt:
		return rangeStmt(t)
	}
	panic("Not Handled: " + reflect.TypeOf(s).String() + " at " + string(s.Pos()))
}

func declStmt(t *ast.DeclStmt) jen.Code {
	return genDecl(t.Decl.(*ast.GenDecl))
}

func emptyStmt(t *ast.EmptyStmt) jen.Code {
	return jen.Id("jen").Dot("Empty").Call()
}

func exprStmt(t *ast.ExprStmt) jen.Code {
	return jen.Id("jen").Add(genExpr(t.X))
}

func goStmt(t *ast.GoStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Dot("Go").Call().Add(genExpr(t.Call))
}

func labeledStmt(t *ast.LabeledStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Add(ident(t.Label)).Dot("Op").Call(jen.Lit(":")).Dot("Line").Call().Dot("Add").Call(stmt(t.Stmt))
}

func sendStmt(t *ast.SendStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Add(genExpr(t.Chan)).Dot("Op").Call(jen.Lit("<-")).Add(genExpr(t.Value))
}

func incDecStmt(t *ast.IncDecStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Add(genExpr(t.X)).Dot("Op").Call(jen.Lit(t.Tok.String()))
}

func assignStmt(t *ast.AssignStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Add(genExprs(t.Lhs)).Dot("Op").Call(jen.Lit(t.Tok.String())).Add(genExprs(t.Rhs))
}

func returnStmt(t *ast.ReturnStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Dot("Return").Call().Add(genExprs(t.Results))
}

func caseClause(t *ast.CaseClause) jen.Code {
	ret := jen.Id("jen")
	if t.List == nil {
		return ret.Dot("Default").Call().Dot("Block").Call(stmts(t.Body)...)
	}
	return ret.Dot("Case").Call(genExprsCode(t.List)...).Dot("Block").Call(stmts(t.Body)...)
}

func typeSwitchStmt(t *ast.TypeSwitchStmt) jen.Code {
	ret := jen.Id("jen")
	var cond []jen.Code
	if t.Init != nil {
		cond = append(cond, stmt(t.Init))
	}
	if t.Assign != nil {
		cond = append(cond, stmt(t.Assign))
	}
	return ret.Dot("Switch").Call(cond...).Add(blockStmt(t.Body))
}

func commClause(t *ast.CommClause) jen.Code {
	ret := jen.Id("jen")
	if t.Comm == nil {
		return ret.Dot("Default").Call().Dot("Block").Call(stmts(t.Body)...)
	}
	return ret.Dot("Case").Call(stmt(t.Comm)).Dot("Block").Call(stmts(t.Body)...)
}

func selectStmt(t *ast.SelectStmt) jen.Code {
	ret := jen.Id("jen")
	return ret.Dot("Select").Call().Add(blockStmt(t.Body))
}

func branchStmt(t *ast.BranchStmt) jen.Code {
	ret := jen.Id("jen")
	switch t.Tok {
	case token.BREAK:
		return ret.Dot("Break").Call()
	case token.CONTINUE:
		return ret.Dot("Continue").Call()
	case token.GOTO:
		return ret.Dot("Goto").Call().Add(ident(t.Label))
	case token.FALLTHROUGH:
		return ret.Dot("Fallthrough").Call()
	}
	return nil
}

func ifStmt(t *ast.IfStmt) jen.Code {
	var cond []jen.Code
	if t.Init != nil {
		cond = append(cond, stmt(t.Init))
	}
	if t.Cond != nil {
		cond = append(cond, jen.Id("jen").Add(genExpr(t.Cond)))
	}
	ret := jen.Id("jen").Dot("If").Call(
		cond...,
	).Add(blockStmt(t.Body))
	if t.Else != nil {
		ret.Dot("Else").Call().Add(stmt(t.Else))
	}
	return ret
}

func switchStmt(t *ast.SwitchStmt) jen.Code {
	var cond []jen.Code
	if t.Init != nil {
		cond = append(cond, stmt(t.Init))
	}
	if t.Tag != nil {
		cond = append(cond, jen.Id("jen").Add(genExpr(t.Tag)))
	}
	return jen.Id("jen").Dot("Switch").Call(cond...).Add(blockStmt(t.Body))
}

func forStmt(t *ast.ForStmt) jen.Code {
	ret := jen.Id("jen")
	var code []jen.Code
	if t.Init != nil {
		code = append(code, stmt(t.Init))
	}
	if t.Init == nil && t.Cond != nil && t.Post != nil {
		code = append(code, jen.Id("jen").Dot("Empty").Call())
	}
	if t.Cond != nil {
		code = append(code, jen.Id("jen").Add(genExpr(t.Cond)))
	}
	if t.Post != nil {
		code = append(code, stmt(t.Post))
	}
	return ret.Dot("For").Call(
		code...,
	).Add(blockStmt(t.Body))
}

func rangeStmt(t *ast.RangeStmt) jen.Code {
	return jen.Id("jen").Dot("For").Call(
		jen.Id("jen").Add(
			jen.Dot("List").Call(genExprsCode([]ast.Expr{t.Key, t.Value})...),
		).Dot("Op").Call(
			jen.Lit(t.Tok.String()),
		).Dot("Range").Call().Add(genExpr(t.X)),
	).Add(blockStmt(t.Body))
}

func blockStmt(s *ast.BlockStmt) jen.Code {
	ret := stmts(s.List)
	return jen.Dot("Block").Call(ret...)
}

func stmts(s []ast.Stmt) []jen.Code {
	var ret []jen.Code
	for _, st := range s {
		ret = append(ret, stmt(st))
	}
	return ret
}

func fieldList(fl *ast.FieldList) []jen.Code {
	var paramsCode []jen.Code
	if fl == nil {
		return paramsCode
	}
	for _, p := range fl.List {
		code := jen.Id("jen")
		code.Add(identsList(p.Names))
		code.Add(genExpr(p.Type))
		paramsCode = append(paramsCode, code)
	}
	return paramsCode
}
