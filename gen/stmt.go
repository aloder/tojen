package gen

import (
	"go/ast"
	"go/token"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func stmt(s ast.Stmt) jen.Code {
	ret := jen.Id("jen")
	switch t := s.(type) {
	case *ast.BadStmt:
	case *ast.DeclStmt:
		return gDecl(t.Decl.(*ast.GenDecl))
	case *ast.GoStmt:
		return ret.Dot("Go").Call().Add(genExpr(t.Call))
	case *ast.EmptyStmt:
		return ret.Dot("Empty").Call()
	case *ast.LabeledStmt:
		return ret.Add(ident(t.Label)).Dot("Op").Call(jen.Lit(":")).Dot("Line").Call().Dot("Add").Call(stmt(t.Stmt))
	case *ast.ExprStmt:
		return ret.Add(genExpr(t.X))
	case *ast.SendStmt:
		return ret.Add(genExpr(t.Chan)).Dot("Op").Call(jen.Lit("<-")).Add(genExpr(t.Value))
	case *ast.IncDecStmt:
		return ret.Add(genExpr(t.X)).Dot("Op").Call(jen.Lit(t.Tok.String()))
	case *ast.AssignStmt:
		return ret.Add(genExprs(t.Lhs)).Dot("Op").Call(jen.Lit(t.Tok.String())).Add(genExprs(t.Rhs))
	case *ast.ReturnStmt:
		return ret.Dot("Return").Call().Add(genExprs(t.Results))
	case *ast.BranchStmt:
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
	case *ast.BlockStmt:
		return blockStatement(t)
	case *ast.IfStmt:
		return ifStmt(t)
	case *ast.CaseClause:
		if t.List == nil {
			return ret.Dot("Default").Call().Dot("Block").Call(stmts(t.Body)...)
		}
		return ret.Dot("Case").Call(genExprsCode2(t.List)...).Dot("Block").Call(stmts(t.Body)...)
	case *ast.SwitchStmt:
		return switchStmt(t)
	case *ast.TypeSwitchStmt:
		var cond []jen.Code
		if t.Init != nil {
			cond = append(cond, stmt(t.Init))
		}
		if t.Assign != nil {
			cond = append(cond, stmt(t.Assign))
		}
		return ret.Dot("Switch").Call(cond...).Add(blockStatement(t.Body))
	case *ast.CommClause:
		if t.Comm == nil {
			return ret.Dot("Default").Call().Dot("Block").Call(stmts(t.Body)...)
		}
		return ret.Dot("Case").Call(stmt(t.Comm)).Dot("Block").Call(stmts(t.Body)...)
	case *ast.SelectStmt:
		return ret.Dot("Select").Call().Add(blockStatement(t.Body))
	case *ast.ForStmt:
		return forStmt(t)
	case *ast.RangeStmt:
		return rangeStmt(t)
	}
	panic("Not Handled: " + reflect.TypeOf(s).String() + " at " + string(s.Pos()))
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
	).Add(blockStatement(t.Body))
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
	return jen.Id("jen").Dot("Switch").Call(cond...).Add(blockStatement(t.Body))
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
	).Add(blockStatement(t.Body))
}

func rangeStmt(t *ast.RangeStmt) jen.Code {
	return jen.Id("jen").Dot("For").Call(
		jen.Id("jen").Add(
			jen.Dot("List").Call(genExprsCode2([]ast.Expr{t.Key, t.Value})...),
		).Dot("Op").Call(
			jen.Lit(t.Tok.String()),
		).Dot("Range").Call().Add(genExpr(t.X)),
	).Add(blockStatement(t.Body))
}
func blockStatement(s *ast.BlockStmt) jen.Code {
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
		code := jen.Qual(jenImp, "Null").Call()
		if len(p.Names) > 1 {
			var names []jen.Code
			for _, n := range p.Names {
				names = append(names, jen.Qual(jenImp, "Id").Call(jen.Lit(n.String())))
			}
			code.Dot("List").Call(names...)
		} else {
			if len(p.Names) == 1 {
				code.Dot("Id").Call(jen.Lit(p.Names[0].String()))
			}
		}
		code.Add(genExpr(p.Type))
		paramsCode = append(paramsCode, code)
	}
	return paramsCode
}
