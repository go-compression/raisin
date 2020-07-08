package algorithm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	lzss "github.com/mrfleap/custom-compression/compressor"
)

func CompressFile(fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	var compressedContents = lzss.Compress(fileContents)
	var compressedFilePath = filepath.Base(fileString) + ".compressed"
	err = ioutil.WriteFile(compressedFilePath, compressedContents, 0644)

	var decompressedContents = lzss.Decompress(compressedContents)
	var decompressedFilePath = filepath.Base(fileString) + ".decompressed"
	err = ioutil.WriteFile(decompressedFilePath, decompressedContents, 0644)
	check(err)
	
	lossless := reflect.DeepEqual(fileContents, decompressedContents)
	fmt.Printf("Lossless: %t\n", lossless)

	fmt.Printf("Original bytes: %v\nCompressed bytes: %v\n", len(fileContents), len(compressedContents))
	percentageDiff := float32(len(compressedContents))/float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
}
