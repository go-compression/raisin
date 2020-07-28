package algorithm

import (
	"fmt"
	lz "github.com/mrfleap/custom-compression/compressor/lz"
	huffman "github.com/mrfleap/custom-compression/compressor/huffman"
	mcc "github.com/mrfleap/custom-compression/compressor/mcc"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
)

var Engines = [...]string{"lzss", "dmc", "huffman", "mcc"}

type CompressedFile struct {
	engine                string
	compressed            []byte
	decompressed          []byte
	pos                   int
	maxSearchBufferLength int
}

func (f *CompressedFile) Read(content []byte) (int, error) {
	if f.decompressed == nil {
		switch f.engine {
		case "lzss":
			f.decompressed = lz.Decompress(f.compressed, true)
		case "dmc":
			f.decompressed = mcc.DMCDecompress(f.compressed)
		case "mcc":
			f.decompressed = mcc.Decompress(f.compressed)
		case "huffman":
			f.decompressed = huffman.Decompress(f.compressed)
		default:
			f.decompressed = lz.Decompress(f.compressed, true)
		}
		
	}
	bytesToWriteOut := len(f.decompressed[f.pos:])
	if len(content) < bytesToWriteOut {
		bytesToWriteOut = len(content)
	}
	for i := 0; i < bytesToWriteOut; i++ {
		content[i] = f.decompressed[f.pos:][i]
	}
	var err error
	if len(f.decompressed[f.pos:]) <= len(content) {
		err = io.EOF
	} else {
		f.pos += len(content)
	}
	return bytesToWriteOut, err
}

func (f *CompressedFile) Write(content []byte) (int, error) {
	var compressed []byte
	switch f.engine {
	case "lzss":
		compressed = lz.Compress(content, true, f.maxSearchBufferLength)
	case "dmc":
		compressed = mcc.DMCCompress(content)
	case "mcc":
		compressed = mcc.Compress(content)
	case "huffman":
		compressed = huffman.Compress(content)
	default:
		compressed = lz.Compress(content, true, f.maxSearchBufferLength)
	}
	

	f.compressed = append(f.compressed, compressed...)
	return len(compressed), nil
}

func GetCompressedFileFromPath(path string) (CompressedFile, error) {
	var cf CompressedFile
	fileContents, err := ioutil.ReadFile(path)
	cf = CompressedFile{compressed: fileContents}
	return cf, err
}

func CompressFile(engine string, fileString string, maxSearchBufferLength int) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("Compressing...\n")

	file := CompressedFile{maxSearchBufferLength: maxSearchBufferLength}
	file.engine = engine
	file.Write(fileContents)

	var compressedFilePath = filepath.Base(fileString) + ".compressed"
	err = ioutil.WriteFile(compressedFilePath, file.compressed, 0644)

	fmt.Printf("Original bytes: %v\n", len(fileContents))
	fmt.Printf("Compressed bytes: %v\n", len(file.compressed))
	percentageDiff := float32(len(file.compressed)) / float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
}

func DecompressFile(engine string, fileString string) []byte {
	compressedFile, err := GetCompressedFileFromPath(fileString)
	compressedFile.engine = engine
	check(err)
	fmt.Printf("LZSS Decompressing...\n")

	stream := make([]byte, 0)
	out := make([]byte, 512)
	for {
		n, err := compressedFile.Read(out)
		if err != nil && err != io.EOF {
			panic(err)
		} else {
			stream = append(stream, out[0:n]...)
		}

		if err == io.EOF {
			break
		}
	}

	var decompressedFilePath = filepath.Base(strings.Replace(fileString, ".compressed", "", -1))
	err = ioutil.WriteFile(decompressedFilePath, stream, 0644)
	check(err)

	return stream
}

func BenchmarkFile(engine string, fileString string, maxSearchBufferLength int) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("Compressing...\n")

	file := CompressedFile{maxSearchBufferLength: maxSearchBufferLength}
	file.engine = engine
	file.Write(fileContents)

	var compressedFilePath = filepath.Base(fileString) + ".compressed"
	err = ioutil.WriteFile(compressedFilePath, file.compressed, 0644)

	fmt.Printf("Decompressing...\n")
	stream := make([]byte, 0)
	out := make([]byte, 512)
	for {
		n, err := file.Read(out)
		if err != nil && err != io.EOF {
			panic(err)
		} else {
			stream = append(stream, out[0:n]...)
		}

		if err == io.EOF {
			break
		}
	}
	var decompressedFilePath = filepath.Base(fileString) + ".decompressed"
	err = ioutil.WriteFile(decompressedFilePath, stream, 0644)
	check(err)

	lossless := reflect.DeepEqual(fileContents, file.decompressed)
	fmt.Printf("Lossless: %t\n", lossless)

	fmt.Printf("Original bytes: %v\n", len(fileContents))
	fmt.Printf("Compressed bytes: %v\n", len(file.compressed))
	if !lossless {
		fmt.Printf("Decompressed bytes: %v\n", len(file.decompressed))
	}
	percentageDiff := float32(len(file.compressed)) / float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
}
