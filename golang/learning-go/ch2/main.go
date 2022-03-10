package main

import (
	"fmt"
	"reflect"
)

func main() {
	var x float64 = 0
	var y float64 = 0
	fmt.Println(reflect.TypeOf(x / y).String())
	const c int = 1
	fmt.Println(c)

	const t float32 = 2.01e38
	fmt.Println(t)
}
