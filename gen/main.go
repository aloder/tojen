package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
)

func funcDecl(s *ast.FuncDecl) jen.Code {
	ret := jen.Qual("github.com/dave/jennifer/jen", "Null").Call().Dot("Func").Call()
	if s.Recv != nil {
		ret.Dot("Params").Call(fieldList(s.Recv)...)
	}
	ret.Add(ident(s.Name))
	ret.Add(funcType(s.Type))
	ret.Add(blockStatement(s.Body))
	return ret
}

var paths = map[string]string{}

// GenerateFile Generates a jennifer file given a series of bytes a package name
// and if you want a main function or not
func GenerateFile(s []byte, packName string, main bool) *jen.File {
	file := jen.NewFile(packName)
	astFile := parseFile(s)
	paths = make(map[string]string)
	anonImports := []jen.Code{}
	for _, i := range astFile.Imports {
		pathVal := i.Path.Value[1 : len(i.Path.Value)-1]
		name := pathVal

		if i.Name == nil {
			idx := strings.Index(pathVal, "/")
			if idx != -1 {
				name = pathVal[idx+1:]
			}
		} else {
			name = i.Name.String()
			if name == "." {
				panic(". imports not supported")
			}
			if name == "_" {
				anonImports = append(anonImports, jen.Lit(pathVal))
				continue
			}
		}
		paths[name] = pathVal
	}
	decls := []string{}
	var codes []jen.Code
	codes = append(codes, jen.Id("ret").Op(":=").Qual(jenImp, "NewFile").Call(jen.Lit(astFile.Name.String())))
	for _, decl := range astFile.Decls {
		code, name := makeJenCode(decl)
		file.Add(code)
		decls = append(decls, name)
		codes = append(codes, jen.Id("ret").Dot("Add").Call(jen.Id(name).Call()))
	}
	if len(anonImports) > 0 {
		codes = append(codes, jen.Id("ret").Dot("Anon").Call(anonImports...))
	}
	codes = append(codes, jen.Return().Id("ret"))
	file.Add(
		jen.Func().Id("genFile").Params().Op("*").Qual(jenImp, "File").Block(codes...),
	)

	if main {
		file.Add(jen.Func().Id("main").Params().Block(
			jen.Id("ret").Op(":=").Id("genFile").Call(),
			jen.Qual("fmt", "Printf").Call(
				jen.Lit("%#v"),
				jen.Id("ret"),
			),
		))
	}
	return file
}
func makeJenCode(s ast.Decl) (jen.Code, string) {
	inner := jen.Null()
	name := ""
	switch t := s.(type) {
	case *ast.GenDecl:
		name = "genDeclAt" + strconv.Itoa(int(t.TokPos))
		inner.Add(gDecl(t))
	case *ast.FuncDecl:
		name = "genFunc" + t.Name.String()
		inner.Add(funcDecl(t))
	}
	return jen.Func().Id(name).Params().Qual(jenImp, "Code").Block(
		jen.Return().Add(inner),
	), name
}

var jenImp = "github.com/dave/jennifer/jen"

func parseFile(code []byte) *ast.File {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func gDecl(g *ast.GenDecl) jen.Code {
	ret := jen.Qual(jenImp, "Null").Call()
	for _, spec := range g.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			ret.Add(genValueSpec(s))
		case *ast.TypeSpec:
			ret.Dot("Type").Call().Add(ident(s.Name)).Add(genExpr(s.Type))

		}
	}
	return ret
}

func ident(s *ast.Ident) jen.Code {
	return jen.Dot("Id").Call(jen.Lit(s.String()))
}
func genValueSpec(s *ast.ValueSpec) jen.Code {
	ret := jen.Dot("Var").Call()
	if len(s.Names) == 1 {
		ret.Add(ident(s.Names[0]))
	} else if len(s.Names) > 1 {
		var n []jen.Code
		for _, name := range s.Names {
			n = append(n, jen.Id("jen").Add(ident(name)))
		}
		ret.Dot("List").Call(jen.List(n...))
	}
	ret.Add(genExpr(s.Type))

	if len(s.Values) > 0 {
		exprs := genExprs(s.Values)
		ret.Dot("Op").Call(jen.Lit("="))
		ret.Add(exprs)
	}
	return ret
}

func basicLit(b *ast.BasicLit) jen.Code {
	switch b.Kind {
	case token.INT:
		i, err := strconv.ParseInt(b.Value, 10, 32)
		if err != nil {
			return nil
		}
		return jen.Dot("Lit").Call(jen.Lit(int(i)))
	case token.FLOAT:
		return jen.Dot("Lit").Call(jen.Id(b.Value))
	case token.IMAG:
		panic("Cannot parse Imaginary Numbers")
	case token.CHAR:
		return jen.Dot("Id").Call(jen.Id("\"" + b.Value + "\""))
	case token.STRING:
		return jen.Dot("Lit").Call(jen.Id(b.Value))
	}
	return nil
}
