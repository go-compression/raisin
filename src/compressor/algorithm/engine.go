package algorithm

import (
	"fmt"
	"io/ioutil"
)

func compress_file(fileString) {
	dat, err := ioutil.ReadFile("/tmp/dat")
	check(err)
	fmt.Print(string(dat))
	fmt.Println("File contents: %s", file_contents)
}
