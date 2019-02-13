tojen
======

tojen is a code generator that generates
[jennifer](http://www.github.com/dave/jennifer) code from a existing file.

## Why?

Well writing code that generates code is tedious. This tool removes some of the
tedium by setting up a base that can be changed and extended to suit your needs
for code generation.

This was mostly inspired by the functionality of the go [text/template](https://golang.org/pkg/text/template/) system. The advantage is that static code was easy to write, but dynamic typesafe code was a big challenge. Also as the project grew the templates would get harder and harder to read. 

I created this project to further bridge the gap between the advantages of the text/template and keeping all the generation in the go language with jennifer. 

## How?

The command line is all you need.

```
go install github.com/aloder/jenjen
```
In your terminal
```
tojen gen [source file]
```
This just takes the sourcefile and outputs the code in the terminal.

```
tojen gen [source file] [output file]
```
This takes the source file and outputs the code in the specified file

## Example


Say you want to generate a static struct.



```go
package model

type User struct {
  Name     string
  Email    string
  Password string
}
```

Running the command 

`tojen gen [path to user file] [output file]`

Generates this
```go
package main

import jen "github.com/dave/jennifer/jen"

func genDeclAt16() jen.Code {
	return jen.Null().Type().Id("User").Struct(
		jen.Id("Name").Id("string"),
		jen.Id("Email").Id("string"),
		jen.Id("Password").Id("string"))
}
func genFile() *jen.File {
	ret := jen.NewFile("model")
	ret.Add(genDeclAt16())
	return ret
}
```

The Idea of this package is not to generate and forget but rather to establish a
boilerplate that allows you to extend and modify.

If I only wanted the user struct code I would modify it to this

```go
func genUserStruct() jen.Code {
	return jen.Type().Id("User").Struct(
		jen.Id("Name").Id("string"),
		jen.Id("Email").Id("string"),
		jen.Id("Password").Id("string"))
}
```
Now we have usable generation of static code that can be used in a project using jennifer. 

