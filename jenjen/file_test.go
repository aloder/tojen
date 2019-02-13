package jenjen

import (
	"bytes"
	"go/format"
	"testing"

	"github.com/aloder/jenjen/run"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type tcg struct {
	Name string
	Code string
}

var tests = []tcg{
	tcg{
		"first",
		`package main

	func main() {}
	`,
	},
	tcg{
		"empty statement",
		`package main

	func main() {
		;
	}
	`,
	},

	tcg{
		"var declaration",
		`package main
		func main() {
		var no = false
	}
	`,
	},
	tcg{
		"multi var declaration",
		`package main
		func main() {
		var no, yes = false, true
	}
	`,
	},
	tcg{
		"Multiple VarDecl",
		`package main

	func main() {
		i, x := 1, 2
	}
	`,
	},
	tcg{
		"struct",
		`package main
	type A struct {
		Name string
	}
	func main() {
		v := A{"new"}
		println(v.Name)
	}
	`,
	},
	tcg{
		"interface",
		`package main
	
	type I interface {
	Name() String
	}
	type A struct {
	name string
	}
	func (a *A) Name()string{
		return a.name
	}
	func try(i I)string{
		return i.Name()
	}
	func main() {
		v := A{"New"}
		try(v)
	}
	`,
	},
	tcg{
		"Map",
		`package main

	func main() {
		m := map[string]int{"one": 1}
		println(m["one"])
	}
	`,
	},
	tcg{
		"Type Assert",
		`package main

	func main() {
		var i interface{} = "hello"
		s := i.(string)
		println(s)
	}
	`,
	},
	tcg{
		"If statement",
		`package main
	func f() int {
		return 1
	}
	func main() {
		i := 1
		if i < 10 {
			println(i)
		}
	}
	func ifElse() {
		if i < 20 {
			println(i)
		} else {
			println(20)
		}
	}
	func ifInit() {
		if x := f(); x <10{
			println(x)
		}
	}
	`,
	},
	tcg{
		"switch statement Init",
		`package main

	func main() {
		switch x := "hello"; x {
		case "hi":
			println("hi")
		case "hello":
			println("hello")
		}
	}`,
	},
	tcg{
		"switch statement Type",
		`package main

func do(i interface{}) {
	switch v := i.(type) {
	case int:
		println(v*2)
	case string:
		println(len(v))
	default:
		println("Dont know type")
	}
}
func main() {
	do(21)
	do("hello")
	do(true)
}
`,
	},
	tcg{
		"switch statement Type With Init and fallthrough",
		`package main

func do(i interface{}) {
	switch x:=i; v := x.(type) {
	case int:
		println(v*2)
	case string:
		println(len(v))
		fallthrough
	case bool:
		println(v)	
	default:
		println("Dont know type")
	}
}
func main() {
	do(21)
	do("hello")
	do(true)
}
`,
	},
	tcg{
		"For Branch Statments",
		`package main

func do(i interface{}) {
	for x := 0; x < 20; x++ {
		if x ==10 {
			break
		}
		if x == 1 {
			continue
		}
	}
}
func main() {
    fmt.Println(1)
    goto End
    fmt.Println(2)
End:
    fmt.Println(3)
}
`,
	},
	tcg{
		"Channel Tour Go Example",
		`package main

import "fmt"

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum
}
func main() {
	s := []int{7, 2, 8, -9, 4, 0}
	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c
	fmt.Println(x, y, x+y)
}
`,
	},
	tcg{
		"Select",
		`package main

func main() {
	var c, c1, c2, c3, c4 chan int
	var i1, i2 int
	select {
		case i1 = <- c1:
			print("received", i1, " from c1")
			print("more ")
		case c2 <- i2:
			print("again")
		case  i3, ok := (<-c3):
		default:
			print("No communication")
	}
}
`,
	},
	tcg{
		"imports + annon",
		`package main
	import (
	"fmt"
	"io/ioutil"
	_ "lib/math"
	)
		
	func main() {
		fmt.Println("Hello World!")
		ioutil.TempDir("go", "fs")
	}
`,
	},
	tcg{
		"function literal",
		`package main
	import (
	"fmt"
	"sort"
	)
		
	func main() {
		people := []string{"Alice", "Bob", "Dave"}
		sort.Slice(people, func(i, j int) bool {
			return len(people[i]) < len(people[j])
		})
		fmt.Println(people)
	}
`,
	},
	tcg{
		"Escaped Characters",
		`package main

func main() {
	print("\n")
}`,
	},
	tcg{
		"slices of slices",
		`package main

import (
	"fmt"
	"strings"
)

func main() {
	board := [][]string{[]string{"_", "_", "_"},[]string{"_", "_", "_"},[]string{"_", "_", "_"}}
	board[0][0] = "X"
	board[2][2] = "O"
	board[1][2] = "X"
	board[1][0] = "O"
	board[0][2] = "X"
	for i := 0; i < len(board); i++ {
		fmt.Printf("%s\n", strings.Join(board[i], " "))
	}
}`,
	},
	tcg{
		"Ellipsis",
		`package main

import "fmt"

func Sum(nums ...int) int {
	res := 0
	for _, n := range nums {
		res += n
	}
	return res
}
func main() {
	s := []int{1, 2, 3}
	fmt.Println(Sum(s...))
}`,
	},
	tcg{
		"Triple Slice",
		`package main

func main() {
	source := []string{"Apple", "Orange", "Plum", "Banana", "Grape"}
	takeOneCapOne := source[2:3:3]
	println(takeOneCapOne)
}`,
	},
	tcg{
		"Fibonacci channels",
		`package main

import (
	"fmt"
	"time"
)

func main() {
	fibonacci := func() chan uint64 {
		c := make(chan uint64)
		go func() {
			var x, y uint64 = 0, 1
			for ; y < (1 << 63); c <- y {
				x, y = y, x+y
			}
			close(c)
		}()
		return c
	}
	c := fibonacci()
	for x, ok := <-c; ok; x, ok = <-c {
		time.Sleep(time.Second)
		fmt.Println(x)
	}
}`,
	},
	tcg{
		"Literals",
		`package main

func main() {
	flo := 1.2
	flo2 := -1.2
	b := true
	x := false
	i := 1
	str := "hello World"
	ch := 'a'
}`,
	},
}

func TestFile(t *testing.T) {
	for i, tc := range tests {
		test := tc
		t.Run(tc.Name, func(t *testing.T) {
			fmtBytes, err := format.Source([]byte(test.Code))
			if err != nil {
				assert.Nil(t, errors.Wrap(err, "Formating error on number: "+string(i)+" name: "+test.Name))
				return
			}
			goFormatTest := string(fmtBytes)
			file := GenerateFile([]byte(test.Code), "main", true)
			resultB := &bytes.Buffer{}
			err = file.Render(resultB)
			if err != nil {
				assert.Nil(t, err, "Could not render test file: \n"+goFormatTest)
				return
			}
			ret, err := run.Exec(resultB.String())
			if err != nil {
				assert.Nil(t, err, "Could not execute rendered test file: \n"+resultB.String())
				return
			}
			fmtBytes, err = format.Source([]byte(*ret))
			if err != nil {
				assert.Nil(t, err, "Could not format file: \n"+*ret+"\n\n"+resultB.String())
				return
			}
			assert.Equal(t, goFormatTest, string(fmtBytes), "Gen Code: \n"+resultB.String())
		})
	}
}
