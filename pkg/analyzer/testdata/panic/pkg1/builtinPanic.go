package pkg1

import "fmt"

func builtinPanicTest() {
	panic(nil) // want "using builtin panic function"

	fmt.Println("aaa")
}
