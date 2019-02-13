package gen

import (
	"go/ast"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func genExprs(s []ast.Expr) jen.Code {
	code := genExprsCode(s)
	if len(code) == 1 {
		return code[0]
	}
	if len(code) == 0 {
		return jen.Null()
	}
	for i := range code {
		code[i] = jen.Id("jen").Add(code[i])
	}
	return jen.Dot("List").Call(code...)
}
func genExprsCode(s []ast.Expr) []jen.Code {
	var code []jen.Code
	for _, expr := range s {
		code = append(code, genExpr(expr))
	}
	return code
}
func genExprsCode2(s []ast.Expr) []jen.Code {
	var code []jen.Code
	for _, expr := range s {
		code = append(code, jen.Id("jen").Add(genExpr(expr)))
	}
	return code
}
func genExpr(s ast.Expr) jen.Code {
	if s == nil {
		return jen.Null()
	}
	switch t := s.(type) {
	case *ast.Ident:
		return jen.Dot("Id").Call(jen.Lit(t.String()))
	case *ast.Ellipsis:
		return jen.Dot("Op").Call(jen.Lit("...")).Add(genExpr(t.Elt))
	case *ast.BasicLit:
		return basicLit(t)
	case *ast.FuncLit:
		return jen.Dot("Func").Call().Add(funcType(t.Type)).Add(blockStatement(t.Body))
	case *ast.CompositeLit:
		return jen.Add(genExpr(t.Type)).Dot("Values").Call(genExprsCode2(t.Elts)...)
	case *ast.ParenExpr:
		return jen.Dot("Parens").Call(jen.Id("jen").Add(genExpr(t.X)))
	case *ast.SelectorExpr:
		dent, ok := t.X.(*ast.Ident)
		if ok {
			path, ok := paths[dent.String()]
			if ok {
				return jen.Dot("Qual").Call(jen.Lit(path), jen.Lit(t.Sel.String()))
			}
		}
		return jen.Add(genExpr(t.X)).Dot("Dot").Call(jen.Lit(t.Sel.String()))
	case *ast.IndexExpr:
		return jen.Add(genExpr(t.X)).Dot("Index").Call(jen.Id("jen").Add(genExpr(t.Index)))
	case *ast.SliceExpr:
		code := []jen.Code{
			jen.Id("jen").Dot("Empty").Call(),
			jen.Id("jen").Dot("Empty").Call(),
		}
		if t.Low != nil {
			code[0] = jen.Id("jen").Add(genExpr(t.Low))
		}
		if t.High != nil {
			code[1] = jen.Id("jen").Add(genExpr(t.High))
		}
		if t.Slice3 {
			code = append(code, jen.Id("jen").Dot("Empty").Call())
			if t.Max != nil {
				code[2] = jen.Id("jen").Add(genExpr(t.Max))
			}
		}
		return jen.Add(genExpr(t.X)).Dot("Index").Call(code...)
	case *ast.TypeAssertExpr:
		ret2 := jen.Add(genExpr(t.X)).Dot("Assert")
		if t.Type == nil {
			return ret2.Call(jen.Id("jen").Dot("Type").Call())
		}
		return ret2.Call(jen.Id("jen").Add(genExpr(t.Type)))
	case *ast.CallExpr:
		args := genExprsCode2(t.Args)
		if t.Ellipsis.IsValid() {
			args[len(args)-1] = jen.Add(args[len(args)-1]).Dot("Op").Call(jen.Lit("..."))
		}
		return jen.Add(genExpr(t.Fun)).Dot("Call").Call(args...)
	case *ast.StarExpr:
		return jen.Dot("Op").Call(jen.Lit("*")).Add(genExpr(t.X))
	case *ast.UnaryExpr:
		return jen.Dot("Op").Call(jen.Lit(t.Op.String())).Add(genExpr(t.X))
	case *ast.BinaryExpr:
		return jen.Add(genExpr(t.X)).Dot("Op").Call(jen.Lit(t.Op.String())).Add(genExpr(t.Y))
	case *ast.KeyValueExpr:
		return jen.Add(genExpr(t.Key)).Dot("Op").Call(jen.Lit(":")).Add(genExpr(t.Value))
	case *ast.ArrayType:
		return arrayType(t)
	case *ast.StructType:
		return structType(t)
	case *ast.FuncType:
		return funcType(t)
	case *ast.InterfaceType:
		return interfaceType(t)
	case *ast.MapType:
		return jen.Dot("Map").Call(jen.Id("jen").Add(genExpr(t.Key))).Add(genExpr(t.Value))
	case *ast.ChanType:
		ret2 := jen.Null()
		if t.Arrow.IsValid() {
			ret2.Dot("Op")
			switch t.Dir {
			case ast.SEND:
				ret2.Call(jen.Lit("->"))
			case ast.RECV:
				ret2.Call(jen.Lit("<-"))
			}
		} else {
			ret2.Dot("Chan").Call()
		}
		return ret2.Add(genExpr(t.Value))
	}
	panic("Not Handled gen expr: " + reflect.TypeOf(s).String() + " at " + string(s.Pos()))
}
