package algorithm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	lzss "github.com/mrfleap/custom-compression/compressor"
)

func CompressFile(fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("LZSS Compressing...\n")
	var compressedContents = lzss.Compress(fileContents, true)
	var compressedFilePath = filepath.Base(fileString) + ".compressed"
	err = ioutil.WriteFile(compressedFilePath, compressedContents, 0644)

	fmt.Printf("Original bytes: %v\n", len(fileContents))
	fmt.Printf("Compressed bytes: %v\n", len(compressedContents))
	percentageDiff := float32(len(compressedContents))/float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
}

func DecompressFile(fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("LZSS Decompressing...\n")
	var decompressedContents = lzss.Decompress(fileContents, true)
	var decompressedFilePath = filepath.Base(strings.Replace(fileString, ".compressed", "", -1))
	err = ioutil.WriteFile(decompressedFilePath, decompressedContents, 0644)
	check(err)
}

func BenchmarkFile(fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("LZSS Compressing...\n")
	var compressedContents = lzss.Compress(fileContents, true)
	var compressedFilePath = filepath.Base(fileString) + ".compressed"
	err = ioutil.WriteFile(compressedFilePath, compressedContents, 0644)

	fmt.Printf("LZSS Decompressing...\n")
	var decompressedContents = lzss.Decompress(compressedContents, true)
	var decompressedFilePath = filepath.Base(fileString) + ".decompressed"
	err = ioutil.WriteFile(decompressedFilePath, decompressedContents, 0644)
	check(err)
	
	lossless := reflect.DeepEqual(fileContents, decompressedContents)
	fmt.Printf("Lossless: %t\n", lossless)

	fmt.Printf("Original bytes: %v\n", len(fileContents))
	fmt.Printf("Compressed bytes: %v\n", len(compressedContents))
	if !lossless {
		fmt.Printf("Decompressed bytes: %v\n", len(decompressedContents))
	}
	percentageDiff := float32(len(compressedContents))/float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
}
