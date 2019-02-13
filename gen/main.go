package gen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
)

var jenImp = "github.com/dave/jennifer/jen"

func funcDecl(s *ast.FuncDecl) jen.Code {
	ret := jen.Qual("github.com/dave/jennifer/jen", "Func").Call()
	if s.Recv != nil {
		ret.Dot("Params").Call(fieldList(s.Recv)...)
	}
	ret.Add(ident(s.Name))
	ret.Add(funcType(s.Type))
	ret.Add(blockStmt(s.Body))
	return ret
}

var paths = map[string]string{}
var formating = false

// GenerateFileBytes takes an array of bytes and transforms it into jennifer
// code
func GenerateFileBytes(s []byte, packName string, main bool, formating bool) ([]byte, error) {
	file := GenerateFile(s, packName, main)
	b := &bytes.Buffer{}
	err := file.Render(b)
	if err != nil {
		return s, err
	}
	ret := b.Bytes()
	if formating {
		ret = formatNulls(ret)
		ret = formatStructs(ret)
		ret = formatBlock(ret)
		ret = formatParams(ret)
		ret, err = goFormat(ret)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func imports(imports []*ast.ImportSpec) (map[string]string, []jen.Code) {
	p := make(map[string]string)
	anonImports := []jen.Code{}
	for _, i := range imports {
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
		p[name] = pathVal
	}
	return p, anonImports
}

// GenerateFile Generates a jennifer file given a series of bytes a package name
// and if you want a main function or not
func GenerateFile(s []byte, packName string, main bool) *jen.File {
	file := jen.NewFile(packName)
	astFile := parseFile(s)
	var anonImports []jen.Code
	// paths is a global variable to map the exported object to the import
	paths, anonImports = imports(astFile.Imports)

	// generate the generative code based on the file
	decls := []string{}
	for _, decl := range astFile.Decls {
		code, name := makeJenCode(decl)
		file.Add(code)
		decls = append(decls, name)
	}

	// generate the function that pieces togeather all the code
	var codes []jen.Code
	codes = append(codes, genNewJenFile(astFile.Name.String()))
	// add anon imports i.e. _ for side effects
	if len(anonImports) > 0 {
		codes = append(codes, jen.Id("ret").Dot("Anon").Call(anonImports...))
	}
	// add the generated functions to the created jen file
	for _, name := range decls {
		codes = append(codes, jen.Id("ret").Dot("Add").Call(jen.Id(name).Call()))
	}
	// return the created jen file
	codes = append(codes, jen.Return().Id("ret"))
	// add the patch function to the output file
	file.Add(
		jen.Func().Id("genFile").Params().Op("*").Qual(jenImp, "File").Block(codes...),
	)
	// if main then generate a main function that prints out the output of the
	// patch function
	if main {
		file.Add(genMainFunc())
	}
	return file
}

func genNewJenFile(name string) jen.Code {
	return jen.Id("ret").Op(":=").Qual(jenImp, "NewFile").Call(jen.Lit(name))
}

func genMainFunc() jen.Code {
	return jen.Func().Id("main").Params().Block(
		jen.Id("ret").Op(":=").Id("genFile").Call(),
		jen.Qual("fmt", "Printf").Call(
			jen.Lit("%#v"),
			jen.Id("ret"),
		),
	)
}

func makeJenCode(s ast.Decl) (jen.Code, string) {
	inner := jen.Null()
	name := ""
	switch t := s.(type) {
	case *ast.GenDecl:
		name = "genDeclAt" + strconv.Itoa(int(t.TokPos))
		inner.Add(genDecl(t))
	case *ast.FuncDecl:
		name = "genFunc" + t.Name.String()
		inner.Add(funcDecl(t))
	}
	return makeJenFileFunc(name, inner), name
}
func makeJenFileFunc(name string, block jen.Code) jen.Code {
	return jen.Func().Id(name).Params().Qual(jenImp, "Code").Block(
		jen.Return().Add(block),
	)
}

func parseFile(code []byte) *ast.File {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	return f
}

func genDecl(g *ast.GenDecl) jen.Code {
	ret := jen.Qual(jenImp, "Null").Call()
	for _, spec := range g.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			ret.Add(valueSpec(s))
		case *ast.TypeSpec:
			ret.Add(typeSpec(s))
		}
	}
	return ret
}

func typeSpec(s *ast.TypeSpec) jen.Code {
	return jen.Dot("Type").Call().Add(ident(s.Name)).Add(genExpr(s.Type))
}

func valueSpec(s *ast.ValueSpec) jen.Code {
	ret := jen.Dot("Var").Call()
	ret.Add(identsList(s.Names))
	ret.Add(genExpr(s.Type))
	if len(s.Values) > 0 {
		ret.Dot("Op").Call(jen.Lit("="))
		ret.Add(genExprs(s.Values))
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
