package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strconv"
	"strings"

	"github.com/aloder/jenjen/run"
	"github.com/dave/jennifer/jen"
)

func makeTestString(code jen.Code) (*string, error) {
	bb := &bytes.Buffer{}
	err := makeTestFile(code).Render(bb)
	if err != nil {
		return nil, err
	}
	str := bb.String()
	return &str, nil
}
func makeTestFile(code jen.Code) *jen.File {
	file := jen.NewFile("main")
	file.Add(jen.Func().Id("main").Params().Block(makeRet(code)...))
	return file
}

func makeRet(code jen.Code) jen.Statement {
	ret := jen.Null()
	ret.Add(
		jen.Id("ret").Op(":=").Add(code),
		jen.Qual("fmt", "Printf").Call(
			jen.Lit("%#v"),
			jen.Id("ret"),
		),
	)
	return *ret
}

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
func main() {
	s := `
	package main
	import (
	"fmt"
	"io/ioutil"
	)
	
	var i = 1
	var b = 2
	type A struct{
		Name string
	}

	func main() {
		fmt.Println("Hello World!")
		ioutil.TempDir("go", "fs")
	}
	`
	file := GenerateFile([]byte(s), "main", true)
	fmt.Printf("%#v\n", file)
	ret, err := run.Exec(fmt.Sprintf("%#v", file))
	if err != nil {
		panic(err)
	}

	fmt.Println(*ret)
}

var paths = map[string]string{}

func makePathMap(imports []*ast.ImportSpec) map[string]string {
	p := make(map[string]string)
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
		}
		p[name] = pathVal
	}
	return p
}

func GenerateFile(s []byte, packName string, main bool) *jen.File {
	file := jen.NewFile(packName)
	astFile := parseFile(s)
	paths = makePathMap(astFile.Imports)
	decls := []string{}
	var codes []jen.Code
	codes = append(codes, jen.Id("ret").Op(":=").Qual(jenImp, "NewFile").Call(jen.Lit(astFile.Name.String())))
	for _, decl := range astFile.Decls {
		code, name := makeJenCode(decl)
		file.Add(code)
		decls = append(decls, name)
		codes = append(codes, jen.Id("ret").Dot("Add").Call(jen.Id(name).Call()))
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

func genTypeSpec(s *ast.TypeSpec) jen.Code {
	ret := jen.Type().Id(s.Name.String())
	if s.Assign.IsValid() {
		ret.Op("=")
	}
	ret.Add(genExpr(s.Type))
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

	ret.Dot("Op").Call(jen.Lit("="))

	ret.Add(genExprs(s.Values))
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
		i, err := strconv.ParseFloat(b.Value, 64)
		if err != nil {
			return nil
		}
		return jen.Dot("Lit").Call(jen.Lit(i))
	case token.IMAG:
		panic("Cannot parse Imaginary Numbers")
	case token.CHAR:
		if len(b.Value) > 0 {
			return jen.Dot("Lit").Call(jen.Lit(b.Value[0]))
		}
	case token.STRING:
		return jen.Dot("Lit").Call(jen.Lit(b.Value[1 : len(b.Value)-1]))
	}
	return nil
}
