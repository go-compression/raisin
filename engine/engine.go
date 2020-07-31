package algorithm

import (
	"fmt"
	lz "github.com/mrfleap/custom-compression/compressor/lz"
	huffman "github.com/mrfleap/custom-compression/compressor/huffman"
	mcc "github.com/mrfleap/custom-compression/compressor/mcc"
	flate "compress/flate"
	gzip "compress/gzip"
	lzw "compress/lzw"
	zlib "compress/zlib"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	ent "github.com/kzahedi/goent/discrete"
	"math"
	"bytes"
	"github.com/jedib0t/go-pretty/table"
	"sort"
	"os"
	"html/template"
	"time"
	"sync"
)

var Engines = [...]string{"all", "suite", "lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib"}
var Suites = map[string][]string{"all": Engines[:], "suite": []string{"lzss", "huffman", "flate", "gzip", "lzw", "zlib"}}

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
		case "zlib":
			var b bytes.Buffer
			b.Write(f.compressed)
			r, err := zlib.NewReader(&b)
			check(err)
			f.decompressed, err = ioutil.ReadAll(r)
			check(err)
			r.Close()
		case "lzw":
			var b bytes.Buffer
			var err error
			b.Write(f.compressed)
			r := lzw.NewReader(&b, lzw.MSB, 8)
			f.decompressed, err = ioutil.ReadAll(r)
			check(err)
			r.Close()
		case "flate":
			var b bytes.Buffer
			var err error
			b.Write(f.compressed)
			r := flate.NewReader(&b)
			check(err)
			f.decompressed, err = ioutil.ReadAll(r)
			check(err)
			r.Close()
		case "gzip":
			var b bytes.Buffer
			b.Write(f.compressed)
			r, err := gzip.NewReader(&b)
			check(err)
			f.decompressed, err = ioutil.ReadAll(r)
			check(err)
			r.Close()
		case "all":
			panic("Cannot decompress with all formats")
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
	case "zlib":
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(content)
		w.Close()
		compressed = b.Bytes()
	case "lzw":
		var b bytes.Buffer
		w := lzw.NewWriter(&b, lzw.MSB, 8)
		w.Write(content)
		w.Close()
		compressed = b.Bytes()
	case "flate":
		var b bytes.Buffer
		w, err := flate.NewWriter(&b, 9)
		check(err)
		w.Write(content)
		w.Close()
		compressed = b.Bytes()
	case "gzip":
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		w.Write(content)
		w.Close()
		compressed = b.Bytes()
	case "all":
		panic("Cannot compress with all formats")
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

func CompressFile(engine string, fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("Compressing...\n")

	file := CompressedFile{maxSearchBufferLength: 4096}
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

type Result struct {
	engine string
	timeTaken string 
	ratio float32
    bitsPerSymbol float32
	entropy  float64
	lossless bool
	failed bool
}

func BenchmarkSuite(files []string, algorithms []string, generateHtml bool) string {
	var html string
	timeout := 1 * time.Minute

	for i, fileString := range files {
		fmt.Printf("Compressing file %d/%d - %s\n", i + 1, len(files), fileString)
		results := make([]Result, 0)
		failedResults := make([]Result, 0)

		fileContents, err := ioutil.ReadFile(fileString)
		check(err)
		fileSize := int64(len(fileContents))

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleLight)
		t.AppendHeader(table.Row{"engine", "time taken", "compression ratio", "bits per symbol", "entropy", "lossless"})

		resultChans := make(map[string]chan Result)
		var wg sync.WaitGroup

		for _, engineName := range algorithms {
			fmt.Println("Benchmarking", engineName)

			resultChannel := make(chan Result, 1)
			resultChans[engineName] = resultChannel

			wg.Add(1)
			go AsyncBenchmarkFile(resultChannel, &wg, engineName, fileString, true)
		}

		waitTimeout(&wg, timeout)

		for engineName, resultChan := range resultChans {
			select {
			case result := <-resultChan:
				if result.failed {
					failedResults = append(failedResults, result)
				} else {
					results = append(results, result)
				}
			default:
				result := Result{}
				result.engine = engineName
				result.timeTaken = fmt.Sprintf(">%s", timeout)
				result.lossless = false
				result.failed = true
				failedResults = append(failedResults, result)
			}
		}

		sort.Slice(results, func(i, j int) bool {
			return results[j].ratio > results[i].ratio
		})

		for _, result := range results {
			t.AppendRow([]interface{}{result.engine, result.timeTaken, fmt.Sprintf("%.2f%%", result.ratio), fmt.Sprintf("%.2f", result.bitsPerSymbol), fmt.Sprintf("%.2f", result.entropy), result.lossless})
		}

		t.AppendSeparator()
		for _, result := range failedResults {
			t.AppendRow([]interface{}{result.engine, result.timeTaken, "DNF", "DNF", "DNF", result.lossless})
		}
		t.AppendSeparator()
		t.AppendRow(table.Row{"File", fileString, "Size", ByteCountSI(fileSize)})
		
		t.Render()
		if generateHtml {
			html = html + "<br>" + t.RenderHTML()
		}
	}
	if generateHtml {
		tmpl := template.Must(template.ParseFiles("templates/benchmark.html"))
		var b bytes.Buffer
		tmpl.Execute(&b, struct{Tables template.HTML}{Tables: template.HTML(html)})
		return b.String()
	} else {
		return ""
	}
}

func AsyncBenchmarkFile(resultChannel chan Result, wg *sync.WaitGroup, engine string, fileString string, suite bool) {
	defer wg.Done()

	errorHandler := func() {
		if recover() != nil {
			fmt.Printf("%s errored during execution, continuing", engine)
			result := Result{}
			result.engine = engine
			result.timeTaken = "failed"
			result.lossless = false
			result.failed = true
			resultChannel <- result
		}
	}

	defer errorHandler()
	start := time.Now()
	result := BenchmarkFile(engine, fileString, suite)
	duration := time.Since(start)
	result.timeTaken = fmt.Sprintf("%s", duration)

	fmt.Printf("%s finished benchmarking", engine)

	resultChannel <- result
}

func BenchmarkFile(engine string, fileString string, suite bool) Result  {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("Compressing...\n")

	symbolFrequencies := make(map[byte]int)
	for _, c := range []byte(fileContents) {
		symbolFrequencies[c]++
	}
	total := len([]byte(fileContents))
	freqs := make([]float64, len(symbolFrequencies))
	i := 0
	for _, freq := range symbolFrequencies {
		freqs[i] = float64(freq) / float64(total)
		i++
	}

	file := CompressedFile{maxSearchBufferLength: 4096}
	file.engine = engine
	file.Write(fileContents)

	if !suite {
		var compressedFilePath = filepath.Base(fileString) + ".compressed"
		err = ioutil.WriteFile(compressedFilePath, file.compressed, 0644)
	}

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
	
	if !suite {
		var decompressedFilePath = filepath.Base(fileString) + ".decompressed"
		err = ioutil.WriteFile(decompressedFilePath, stream, 0644)
		check(err)
	}

	lossless := reflect.DeepEqual(fileContents, file.decompressed)
	percentageDiff := float32(len(file.compressed)) / float32(len(fileContents)) * 100
	entropy := ent.Entropy(freqs, math.Log)
	bps := float32(len(file.compressed) * 8) / float32(len(fileContents))

	if !suite {
		fmt.Printf("Lossless: %t\n", lossless)

		fmt.Printf("Original bytes: %v\n", len(fileContents))
		fmt.Printf("Compressed bytes: %v\n", len(file.compressed))
		if !lossless {
			fmt.Printf("Decompressed bytes: %v\n", len(file.decompressed))
		}
		fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
		fmt.Printf("Shannon entropy: %.2f\n", entropy)
		fmt.Printf("Average bits per symbol: %.2f\n", bps)
	}
	return Result{engine, "", percentageDiff, bps, entropy, lossless, false}
}