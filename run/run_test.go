package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunExec(t *testing.T) {
	one := `
	package main

	func main() {
		println("Hello World!")
	}
	`
	out, err := Exec(one)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World!\n", *out)
}
