package gen

import (
	"go/ast"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func genExprs(s []ast.Expr) jen.Code {
	if len(s) == 0 {
		return jen.Null()
	}
	if len(s) == 1 {
		return genExpr(s[0])
	}
	code := genExprsCode(s)
	return jen.Dot("List").Call(code...)
}

func genExprsCode(s []ast.Expr) []jen.Code {
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
		return ident(t)
	case *ast.Ellipsis:
		return ellipsis(t)
	case *ast.BasicLit:
		return basicLit(t)
	case *ast.FuncLit:
		return funcLit(t)
	case *ast.CompositeLit:
		return compositeLit(t)
	case *ast.ParenExpr:
		return parenExpr(t)
	case *ast.SelectorExpr:
		return selectorExpr(t)
	case *ast.IndexExpr:
		return indexExpr(t)
	case *ast.SliceExpr:
		return sliceExpr(t)
	case *ast.TypeAssertExpr:
		return typeAssertExpr(t)
	case *ast.CallExpr:
		return callExpr(t)
	case *ast.StarExpr:
		return starExpr(t)
	case *ast.UnaryExpr:
		return unaryExpr(t)
	case *ast.BinaryExpr:
		return binaryExpr(t)
	case *ast.KeyValueExpr:
		return keyValueExpr(t)
	case *ast.ArrayType:
		return arrayType(t)
	case *ast.StructType:
		return structType(t)
	case *ast.FuncType:
		return funcType(t)
	case *ast.InterfaceType:
		return interfaceType(t)
	case *ast.MapType:
		return mapType(t)
	case *ast.ChanType:
		return chanType(t)
	}
	panic("Not Handled gen expr: " + reflect.TypeOf(s).String() + " at " + string(s.Pos()))
}
func ellipsis(t *ast.Ellipsis) jen.Code {
	return jen.Dot("Op").Call(jen.Lit("...")).Add(genExpr(t.Elt))
}

func funcLit(t *ast.FuncLit) jen.Code {
	return jen.Dot("Func").Call().Add(funcType(t.Type)).Add(blockStmt(t.Body))
}

func compositeLit(t *ast.CompositeLit) jen.Code {
	return jen.Add(genExpr(t.Type)).Dot("Values").Call(genExprsCode(t.Elts)...)
}

func parenExpr(t *ast.ParenExpr) jen.Code {
	return jen.Dot("Parens").Call(jen.Id("jen").Add(genExpr(t.X)))
}

func indexExpr(t *ast.IndexExpr) jen.Code {
	return jen.Add(genExpr(t.X)).Dot("Index").Call(jen.Id("jen").Add(genExpr(t.Index)))
}
func starExpr(t *ast.StarExpr) jen.Code {
	return jen.Dot("Op").Call(jen.Lit("*")).Add(genExpr(t.X))
}
func unaryExpr(t *ast.UnaryExpr) jen.Code {
	return jen.Dot("Op").Call(jen.Lit(t.Op.String())).Add(genExpr(t.X))
}
func binaryExpr(t *ast.BinaryExpr) jen.Code {
	return jen.Add(genExpr(t.X)).Dot("Op").Call(jen.Lit(t.Op.String())).Add(genExpr(t.Y))
}

func keyValueExpr(t *ast.KeyValueExpr) jen.Code {
	ret := jen.Add(genExpr(t.Key)).Dot("Op").Call(jen.Lit(":")).Add(genExpr(t.Value))
	return ret
}

func mapType(t *ast.MapType) jen.Code {
	ret := jen.Dot("Map").Call(
		jen.Id("jen").Add(genExpr(t.Key)),
	).Add(genExpr(t.Value))
	return ret
}

func selectorExpr(t *ast.SelectorExpr) jen.Code {
	dent, ok := t.X.(*ast.Ident)
	if ok {
		path, ok := paths[dent.String()]
		if ok {
			return jen.Dot("Qual").Call(jen.Lit(path), jen.Lit(t.Sel.String()))
		}
	}
	return jen.Add(genExpr(t.X)).Dot("Dot").Call(jen.Lit(t.Sel.String()))
}

func identsList(s []*ast.Ident) jen.Code {
	if len(s) == 0 {
		return jen.Null()
	}
	if len(s) == 1 {
		return ident(s[0])
	}
	var n []jen.Code
	for _, name := range s {
		n = append(n, jen.Id("jen").Add(ident(name)))
	}
	return jen.Dot("List").Call(jen.List(n...))
}

func ident(s *ast.Ident) jen.Code {
	return jen.Dot("Id").Call(jen.Lit(s.String()))
}

func typeAssertExpr(t *ast.TypeAssertExpr) jen.Code {
	ret2 := jen.Add(genExpr(t.X)).Dot("Assert")
	if t.Type == nil {
		return ret2.Call(jen.Id("jen").Dot("Type").Call())
	}
	return ret2.Call(jen.Id("jen").Add(genExpr(t.Type)))
}

func callExpr(t *ast.CallExpr) jen.Code {
	args := genExprsCode(t.Args)
	if t.Ellipsis.IsValid() {
		args[len(args)-1] = jen.Add(args[len(args)-1]).Dot("Op").Call(jen.Lit("..."))
	}
	return jen.Add(genExpr(t.Fun)).Dot("Call").Call(args...)
}

func sliceExpr(t *ast.SliceExpr) jen.Code {
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
}

func chanType(t *ast.ChanType) jen.Code {
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
