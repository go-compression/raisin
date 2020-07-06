package main

import (
	"compressor/cmd"
	"fmt"
)

// https://github.com/spf13/cobra#getting-started

func main() {

	cmd.Execute()
	fmt.Printf("Hello")
}
