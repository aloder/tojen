package gen

import (
	"go/ast"

	"github.com/dave/jennifer/jen"
)

func funcType(s *ast.FuncType) jen.Code {
	var ret jen.Statement
	params := fieldList(s.Params)
	ret.Dot("Params").Call(params...)
	results := fieldList(s.Results)
	if len(results) > 0 {
		ret.Dot("Params").Call(results...)
	}
	return &ret
}
func arrayType(s *ast.ArrayType) jen.Code {
	return jen.Dot("Index").Call().Add(genExpr(s.Elt))
}
func structType(s *ast.StructType) jen.Code {
	return jen.Dot("Struct").Call(fieldList(s.Fields)...)
}

func interfaceType(s *ast.InterfaceType) jen.Code {
	return jen.Dot("Interface").Call(fieldList(s.Methods)...)
}
