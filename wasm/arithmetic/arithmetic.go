package main

import (
	// "bytes"
	"fmt"
	"github.com/go-compression/raisin/compressor/arithmetic_logical"
	"syscall/js"
)

func main() {
	// arithmeticEncode("TEST")
	js.Global().Set("arithmeticEncode", jsonWrapper())
	<-make(chan bool)
}

func arithmeticEncode(input string) (float64, float64) {
	fmt.Println("Compressing", input)

	bot, top := arithmetic_logical.Range([]byte(input))
	return bot, top
}

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		input := args[0].String()
		bot, top := arithmeticEncode(input)
		return []interface{}{bot, top}
	})
	return jsonFunc
}
