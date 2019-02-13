package main

import (
	"fmt"
	jen "github.com/dave/jennifer/jen"
)

func genDeclAt15() jen.Code {
	return jen.Null()
}
func genFuncsum() jen.Code {
	return jen.Null().Func().Id("sum").Params(jen.Null().Id("s").Index().Id("int"), jen.Null().Id("c").Chan().Id("int")).Block(jen.Id("sum").Op(":=").Lit(0), jen.For(jen.List(jen.Id("_"), jen.Id("v")).Op(":=").Range().Id("s")).Block(jen.Id("sum").Op("+=").Id("v")), jen.Id("sum").Id("c"))
}
func genFuncmain() jen.Code {
	return jen.Null().Func().Id("main").Params().Block(jen.Id("s").Op(":=").Index().Id("int").Values(jen.Lit(7), jen.Lit(2), jen.Lit(8), jen.Op("-").Lit(9), jen.Lit(4), jen.Lit(0)), jen.Id("c").Op(":=").Id("make").Call(jen.Chan().Id("int")), jen.Go().Id("sum").Call(jen.Id("s").Index(jen.Empty(), jen.Id("len").Call(jen.Id("s")).Op("/").Lit(2)), jen.Id("c")), jen.Go().Id("sum").Call(jen.Id("s").Index(jen.Id("len").Call(jen.Id("s")).Op("/").Lit(2), jen.Empty()), jen.Id("c")), jen.List(jen.Id("x"), jen.Id("y")).Op(":=").List(jen.Op("<-").Id("c"), jen.Op("<-").Id("c")), jen.Qual("fmt", "Println").Call(jen.Id("x"), jen.Id("y"), jen.Id("x").Op("+").Id("y")))
}
func genFile() *jen.File {
	ret := jen.NewFile("main")
	ret.Add(genDeclAt15())
	ret.Add(genFuncsum())
	ret.Add(genFuncmain())
	return ret
}
func main() {
	ret := genFile()
	fmt.Printf("%#v", ret)
}
