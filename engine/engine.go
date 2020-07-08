package algorithm

import (
	"io/ioutil"
	"path/filepath"
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
}
