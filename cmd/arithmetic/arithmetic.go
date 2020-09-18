package main

import (
	"bytes"
	"fmt"
	"github.com/go-compression/raisin/compressor/arithmetic"
	"syscall/js"
)

func main() {
	js.Global().Set("arithmeticEncode", jsonWrapper())
	<-make(chan bool)
}

func arithmeticEncode(input string) string {
	fmt.Println("Compressing", input)

	var output bytes.Buffer

	w := arithmetic.NewWriter(&output)

	w.Write([]byte(input))
	w.Close()

	compressed := output.Bytes()
	fmt.Println("Compressed", string(compressed))
	return string(compressed)
}

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		input := args[0].String()
		output := arithmeticEncode(input)
		return output
	})
	return jsonFunc
}
