# Raisin

[![Build Status](https://travis-ci.com/go-compression/raisin.svg?branch=master)](https://travis-ci.com/go-compression/raisin) [![Go Report Card](https://goreportcard.com/badge/github.com/go-compression/raisin)](https://goreportcard.com/report/github.com/go-compression/raisin) [![Coverage Status](https://coveralls.io/repos/github/go-compression/raisin/badge.svg?branch=master)](https://coveralls.io/github/go-compression/raisin?branch=master) [![Documentation](https://godoc.org/github.com/yangwenmai/how-to-add-badge-in-github-readme?status.svg)](http://godoc.org/github.com/go-compression/raisin)

[View benchmarks for latest deployment](https://go-compression.github.io/raisin/)

A simple lightweight set of implementations and bindings for compression algorithms written in Go.

This project contains the source code for a summer mentorship about learning how to implement and create different compression algorithms. This includes common algorithms such as huffman, lempel-ziv, and arithmetic along with bindings for builtin Go compression algorithms.

## Usage from the CLI

To start using this package from the command line, install it with `go install`

```console
$ go install github.com/go-compression/raisin/
```

Once done, you should be able to start using it

```console
$ echo "Hello world!" > test.txt
$ raisin test.txt
Compressing...
Original bytes: 13
Compressed bytes: 14
Compression ratio: 107.69%
$ cat test.txt.compressed
�ӷ     �?��KD+
               �
$ rm test.txt
$ raisin -decompress test.txt.compressed
Decompressing...
$ cat test.txt
Hello world!
```

The possible commands include:

- `-compress` - Compress a given file and output the compressed contents to a file with ".compressed" at the end
- `-decompress` - Decompress a given file and output the decompressed contents to a file without ".compressed" at the end
- `-benchmark` - Benchmark a given file and measure the compression ratio, outputs a .compressed and a .decompressed file

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
$ raisin -algorithm=arithmetic test.txt
```

You can also combine algorithms together in "layers", this will essentially compress the file with the first algorithm, then the second, etc. This stacking of algorithms is what powers virtually all modern compression, gzip and zip is powered by the FLATE algorithm which is essentially lempel-ziv (similar to lzss) and huffman coding stacked on toip of each other.

```console
$ raisin -algorithm=lzss,huffman test.txt
Compressing...
Compression ratio: 307.69%
$ raisin -decompress -algorithm=lzss,huffman test.txt.compressed
Decompressing...
```

On top of this, you can easily compress or decompress multiple files by chaining them together with commas.

```console
$ raisin test1.txt,test2.txt
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
$ raisin -delete test.txt
Compressing...
Compression ratio: 107.69%
$ ls
test.txt.compressed
$ raisin -decompress -delete test.txt.compressed
Decompressing...
$ ls
test.txt
$ raisin -delete=false test.txt
Compressing...
Compression ratio: 107.69%
$ ls
test.txt  test.txt.compressed
```

The `out` command simply lets you change what file is outputted when compressing a single file:

```console
$ echo "Hello world!" > test.txt
$ raisin -out=compressed.txt test.txt
Compressing...
Compression ratio: 107.69%
$ ls
test.txt  compressed.txt
$ raisin -decompress -out=decompressed.txt compressed.txt
Decompressing...
$ ls
test.txt  decompressed.txt
```

`outext` is similar to `out` but exists for when we compress/decompress multiple files. If `outext` is provided, it will be used as the **out**put **ext**ension for the files. Note that the default for compression for outext is `.compressed` and for decompression it's an empty string (`outext=`) which tells the program to remove the last extension.

```console
$ ls
test1.txt  test2.txt  test3.txt
$ raisin -delete -outext=.testing test1.txt,test2.txt,test3.txt
Compressing...
Compression ratio: 107.69%
$ ls
test1.txt.testing  test2.txt.testing  test3.txt.testing
$ raisin -decompress -outext=.decompressed test1.txt.testing,test2.txt.testing,test3.txt.testing
Decompressing...
$ ls
test1.txt  test2.txt  test3.txt
```

## Benchmarking

You can use the `benchmark` command to generate benchmarked results for a set of algorithms, layers, and files. This is helpful for generating results in a table, [website](https://go-compression.github.io/raisin/), or in bindings for other languages such as python (see the `ai` folder).

Usage is relatively similar to the `compress` and `decompress` commands.

```console
$ echo "Hello world!" > test.txt
$ echo "abcabcabcabcabcabcabcabc" > test2.txt
$ raisin -benchmark -algorithm=lzss,huffman,arithmetic,gzip,[lzss,arithmetic] test.txt,test2.txt
┌─────────────────┬────────────┬───────────────────┬────────────────┬─────────────────────┬──────────┐
│ ENGINE          │ TIME TAKEN │ COMPRESSION RATIO │ ACTUAL ENTROPY │ THEORETICAL ENTROPY │ LOSSLESS │
├─────────────────┼────────────┼───────────────────┼────────────────┼─────────────────────┼──────────┤
│ lzss            │ 350µs      │ 100.00%           │ 2.20           │ 2.20                │ true     │
│ arithmetic      │ 50µs       │ 107.69%           │ 2.12           │ 2.20                │ true     │
│ lzss,arithmetic │ 210µs      │ 107.69%           │ 2.12           │ 2.20                │ true     │
│ gzip            │ 280µs      │ 284.62%           │ 1.14           │ 2.20                │ true     │
│ huffman         │ 190µs      │ 307.69%           │ 1.08           │ 2.20                │ true     │
├─────────────────┼────────────┼───────────────────┼────────────────┼─────────────────────┼──────────┤
│ File            │ test.txt   │ Size              │ 13 B           │                     │          │
└─────────────────┴────────────┴───────────────────┴────────────────┴─────────────────────┴──────────┘
┌─────────────────┬────────────┬───────────────────┬────────────────┬─────────────────────┬──────────┐
│ ENGINE          │ TIME TAKEN │ COMPRESSION RATIO │ ACTUAL ENTROPY │ THEORETICAL ENTROPY │ LOSSLESS │
├─────────────────┼────────────┼───────────────────┼────────────────┼─────────────────────┼──────────┤
│ lzss,arithmetic │ 160µs      │ 84.00%            │ 1.25           │ 1.22                │ true     │
│ lzss            │ 310µs      │ 84.00%            │ 1.25           │ 1.22                │ true     │
│ arithmetic      │ 160µs      │ 84.00%            │ 1.25           │ 1.22                │ true     │
│ huffman         │ 170µs      │ 92.00%            │ 1.24           │ 1.22                │ true     │
│ gzip            │ 430µs      │ 120.00%           │ 1.17           │ 1.22                │ true     │
├─────────────────┼────────────┼───────────────────┼────────────────┼─────────────────────┼──────────┤
│ File            │ test2.txt  │ Size              │ 25 B           │                     │          │
└─────────────────┴────────────┴───────────────────┴────────────────┴─────────────────────┴──────────┘
```

A larger example, taken from the `.travis.yml` file to generate the [benchmark page](https://go-compression.github.io/raisin/). Notice the `-generate` flag, this tells it to generate an html file and output it as `index.html`, which is then used and uploaded to the [GitHub Pages branch](https://github.com/go-compression/raisin/tree/gh-pages). Keep in mind the program expects a template file to be at `templates/benchmark.html` relative to your working directory. The command is as follows:

```console
$ raisin -benchmark -generate -algorithm=lzss,dmc,huffman,flate,gzip,lzw,zlib,arithmetic,[lzss,huffman],[lzss,arithmetic],[arithmetic,huffman] alice29.txt,asyoulik.txt,cp.html,fields.c,grammar.lsp,kennedy.xls,lcet10.txt,plrabn12.txt,ptt5,sum,xargs.1
```

Shout-out to [jedib0t](https://github.com/jedib0t) for his wonderful [go-pretty module](https://github.com/jedib0t/go-pretty) for generating these tables and the HTML tables used in the GitHub Pages site.

## Building

To build the binary from source, simply `go get` the package:

```console
$ go get -u github.com/go-compression/raisin
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
	"github.com/go-compression/raisin/engine"
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

Documentation is available at [godoc](https://godoc.org/github.com/go-compression/raisin), please note that most of the code is currently undocumented as it is still a work in progress.
