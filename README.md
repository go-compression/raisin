# Custom Compression Engine

[![Build Status](https://travis-ci.com/mrfleap/custom-compression.svg?branch=master)](https://travis-ci.com/mrfleap/custom-compression)

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
$ custom-compression -algorithm arithmetic compress test.txt
Compressing...
Original bytes: 13
Compressed bytes: 14
Compression ratio: 107.69%
$ cat test.txt.compressed
�ӷ     �?��KD+
               �
$ rm test.txt
$ custom-compression -algorithm arithmetic decompress test.txt.compressed
Decompressing...
$ cat test.txt
Hello world!
```

The possible commands include:

```
compress - Compress a given file and output the compressed contents to a file with ".compressed" at the end
decompress - Decompress a given file and output the decompressed contents to a file without ".compressed" at the end
benchmark - Benchmark a given file and measure the compression ratio, outputs a .compressed and a .decompressed file
```

The most important flag is the `-algorithm` flag which allows you to specify which algorithm to use during compression, decompression, or benchmarking. The possible algorithms include:

- lzss
- dmc
- huffman
- mcc
- arithmetic
- flate
- gzip
- lzw
- zlib
- all
- suite

The last two algorithms, `all` and `suite`, can only be used with the benchmark commands. When used, they will benchmark a set of algorithms and return the results as a table.

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
