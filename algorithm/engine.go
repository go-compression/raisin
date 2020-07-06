package algorithm

import (
	"fmt"
	"io/ioutil"
)

func CompressFile(fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Println("File contents: %s", string(fileContents))
}
