package main

import (
	"go/ast"
	"testing"

	"github.com/aloder/jenjen/run"
	"github.com/stretchr/testify/assert"
)

var header = "package main"

func TestGenDeclLit(t *testing.T) {
	t.Parallel()
	tests := []string{
		"var one = 1",
		"var two = \"two\"",
		`var float = 1.3`,
	}
	for _, tes := range tests {
		test := tes
		t.Run("something", func(t *testing.T) {
			t.Parallel()
			str := header + "\n" + test
			f := parseFile([]byte(str))
			expr := f.Decls[0].(*ast.GenDecl)
			done := gDecl(expr)
			testStr, err := makeTestString(done)
			if err != nil {
				panic(err)
			}
			str2, err := run.Exec(*testStr)
			assert.Nil(t, err)
			if err == nil {
				assert.Equal(t, test, *str2)
			}
		})
	}
}
func TestFuncDecl(t *testing.T) {
	t.Parallel()
	type tc struct {
		name string
		test string
	}
	tests := []tc{
		tc{"main", `func main() {}`},
		tc{"main with one param", `func main(a int) {}`},
		tc{"two", `func two(a, b int) {}`},
		tc{"three", `func three(a, b, c int) {}`},
		tc{"four with two params", `func four(a int, b string) {}`},
		tc{"four with return int", `func four() int {}`},
		tc{"five with return params", `func five() (*string, *error) {}`},
	}
	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			str := header + "\n" + tc.test
			f := parseFile([]byte(str))
			expr := f.Decls[0].(*ast.FuncDecl)
			done := funcDecl(expr)
			testStr, err := makeTestString(done)
			if err != nil {
				panic(err)
			}
			assert.Nil(t, err)
			str2, err := run.Exec(*testStr)
			assert.Equal(t, tc.test, *str2)
		})

	}
}
func TestFuncBlock(t *testing.T) {
	t.Parallel()
	tests := []string{
		`func main() {
	one := 1
}`,
		`func main() {
	for i := 1; i < 10; i++ {
		println(i)
	}
}`,
		`func main() {
	lst := []string{"one", "two", "three"}
	for i, s := range lst {
		println(s)
		println(i)
	}
}`,
		`func main() {
	lst := []string{"one", "two", "three"}
	println(lst[0])
}`,
		`func main() {
	lst := []string{"one", "two", "three"}
	println(lst[0])
	go main()
}`,
	}
	for _, tes := range tests {
		test := tes
		t.Run("something", func(t *testing.T) {
			t.Parallel()
			str := header + "\n" + test
			f := parseFile([]byte(str))
			expr := f.Decls[0].(*ast.FuncDecl)
			done := funcDecl(expr)
			testStr, err := makeTestString(done)
			if err != nil {
				panic(err)
			}
			assert.Nil(t, err)
			str2, err := run.Exec(*testStr)

			assert.Equal(t, test, *str2)
		})
	}
}
func TestStructInterface(t *testing.T) {
	t.Parallel()
	tests := []string{
		`type A struct {
	Name string
}`,
		`type A interface {
	Name() string
}`,
	}
	for _, tes := range tests {
		test := tes
		t.Run("struct/Interface", func(t *testing.T) {
			t.Parallel()
			str := header + "\n" + test
			f := parseFile([]byte(str))
			expr := f.Decls[0].(*ast.GenDecl)
			done := gDecl(expr)
			testStr, err := makeTestString(done)
			if err != nil {
				panic(err)
			}
			assert.Nil(t, err)
			str2, err := run.Exec(*testStr)

			assert.Equal(t, test, *str2)
		})

	}
}
