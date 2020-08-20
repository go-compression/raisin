# Custom Compression Engine

[![Build Status](https://travis-ci.com/mrfleap/custom-compression.svg?branch=master)](https://travis-ci.com/mrfleap/custom-compression) [![Go Report Card](https://goreportcard.com/badge/github.com/mrfleap/custom-compression)](https://goreportcard.com/report/github.com/mrfleap/custom-compression) [![Coverage Status](https://coveralls.io/repos/github/mrfleap/custom-compression/badge.svg?branch=master)](https://coveralls.io/github/mrfleap/custom-compression?branch=master) [![Documentation](https://godoc.org/github.com/yangwenmai/how-to-add-badge-in-github-readme?status.svg)](http://godoc.org/github.com/mrfleap/custom-compression)

[View benchmarks for latest deployment](https://mrfleap.github.io/custom-compression/)

This project contains the source code for a summer mentorship about learning how to implement and create compression algorithms in Go.

This project contains several common compression algorithms implemented in Go along with bindings for builtin Go compression algorithms.

## Usage from the CLI

To start using this package from the command line, install it with the command

```console
$ go install github.com/mrfleap/custom-compression/
```

Once done, you should be able to start using it

```console
$ echo "Hello world!" > test.txt
$ custom-compression compress test.txt
Compressing...
Original bytes: 13
Compressed bytes: 14
Compression ratio: 107.69%
$ cat test.txt.compressed
�ӷ     �?��KD+
               �
$ rm test.txt
$ custom-compression decompress test.txt.compressed
Decompressing...
$ cat test.txt
Hello world!
```

The possible commands include:

- `compress` - Compress a given file and output the compressed contents to a file with ".compressed" at the end
- `decompress` - Decompress a given file and output the decompressed contents to a file without ".compressed" at the end
- `benchmark` - Benchmark a given file and measure the compression ratio, outputs a .compressed and a .decompressed file

The most important flag is the `-algorithm` flag which allows you to specify which algorithm to use during compression, decompression, or benchmarking. By default for `compress` and `decompress` this is `lzss,arithmetic`. The possible algorithms include:

- lzss
- dmc
- huffman
- mcc
- arithmetic
- flate
- gzip
- lzw
- zlib

Here's an example of usage:

```console
$ custom-compression compress -algorithm=arithmetic test.txt
```

You can also combine algorithms together in "layers", this will essentially compress the file with the first algorithm, then the second, etc. This stacking of algorithms is what powers virtually all modern compression, gzip and zip is powered by the FLATE algorithm which is essentially lempel-ziv (similar to lzss) and huffman coding stacked on toip of each other.

```console
$ custom-compression compress -algorithm=lzss,huffman test.txt
Compressing...
Compression ratio: 307.69%
$ custom-compression decompress -algorithm=lzss,huffman test.txt.compressed
Decompressing...
```

On top of this, you can easily compress or decompress multiple files by chaining them together with commas.

```console
$ custom-compression compress test1.txt,test2.txt
Compressing...
Compression ratio: 68.53%
$ ls
test1.txt  test1.txt.compressed  test2.txt  test2.txt.compressed
```

When using `compress` and `decompress` a few more options become available to make it easy to use from the command line:

- `delete` - Delete original file after compression/decompressed (defaults to true for decompression)
- `out` - File name to be outputted (defaults to original file + .compressed for compression and file - .compressed for decompression, only available with a single file being compressed/decompressed)
- `outext` - File extension to be outputted when compressing multiple files (unavailable with a single file being compressed/decompressed)

Let's take at the usage of `delete`, keep in mind that `delete` is on by default for `decompress`ing.

```console
$ echo "Hello world!" > test.txt
$ custom-compression compress -delete test.txt
Compressing...
Compression ratio: 107.69%
$ ls
test.txt.compressed
$ custom-compression decompress -delete test.txt.compressed
Decompressing...
$ ls
test.txt
$ custom-compression compress -delete=false test.txt
Compressing...
Compression ratio: 107.69%
$ ls
test.txt  test.txt.compressed
```

The `out` command simply lets you change what file is outputted when compressing a single file:

```console
$ echo "Hello world!" > test.txt
$ custom-compression compress -out=compressed.txt test.txt
Compressing...
Compression ratio: 107.69%
$ ls
test.txt  compressed.txt
$ custom-compression decompress -out=decompressed.txt compressed.txt
Decompressing...
$ ls
test.txt  decompressed.txt
```

`outext` is similar to `out` but exists for when we compress/decompress multiple files. If `outext` is provided, it will be used as the **out**put **ext**ension for the files. Note that the default for compression for outext is `.compressed` and for decompression it's an empty string (`outext=`) which tells the program to remove the last extension.

```console
$ ls
test1.txt  test2.txt  test3.txt
$ custom-compression compress -delete -outext=.testing test1.txt,test2.txt,test3.txt
Compressing...
Compression ratio: 107.69%
$ ls
test1.txt.testing  test2.txt.testing  test3.txt.testing
$ custom-compression decompress -outext=.decompressed test1.txt.testing,test2.txt.testing,test3.txt.testing
Decompressing...
$ ls
test1.txt  test2.txt  test3.txt
```

## Building

To build the binary from source, simply `go get` the package:

```console
$ go get -u github.com/mrfleap/custom-compression
```

Install the dependencies:

```console
$ go get
```

And build:

```console
$ go build
```

## Usage as a module

To use this package as a module, simply import the engine package and use the io.Reader and io.Writer interfaces.

```go
import (
	"fmt"
	"github.com/mrfleap/custom-compression/engine"
)

func main() {
	text := []byte("Hello world!")

	file := engine.CompressedFile{}
	file.CompressionEngine = "arithmetic"
	file.Write(text)
	fmt.Println("Compressed:", string(file.Compressed))
}
```

## Documentation

Documentation is available at [godoc](https://godoc.org/github.com/mrfleap/custom-compression), please note that most of the code is currently undocumented as it is still a work in progress.
