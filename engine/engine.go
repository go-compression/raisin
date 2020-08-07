package algorithm

import (
	"fmt"
	lz "github.com/mrfleap/custom-compression/compressor/lz"
	arithmetic "github.com/mrfleap/custom-compression/compressor/arithmetic"
	huffman "github.com/mrfleap/custom-compression/compressor/huffman"
	mcc "github.com/mrfleap/custom-compression/compressor/mcc"
	dmc "github.com/mrfleap/custom-compression/compressor/dmc"
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
	"strconv"
	"runtime/debug"
)

var Engines = [...]string{"all", "suite", "lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib", "arithmetic"}
var Suites = map[string][]string{"all": Engines[2:], "suite": []string{"lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib"}}

type CompressedFile struct {
	engine                string
	compressed            []byte
	decompressed          []byte
	pos                   int
	maxSearchBufferLength int
}

var Readers = map[string]interface{}{
 	"lzss": lz.NewReader,
	"dmc": dmc.NewReader,
	"mcc": mcc.NewReader,
	"huffman": huffman.NewReader,
	"arithmetic": arithmetic.NewReader,
	"zlib": zlib.NewReader,
	"flate": flate.NewReader,
	"gzip": gzip.NewReader,
	"lzw": lzw.NewReader,
}

func (f *CompressedFile) Read(content []byte) (int, error) {
	if f.decompressed == nil {
		newReader := Readers[f.engine]
		var r io.Reader
		var b io.Reader
		b = bytes.NewReader(f.compressed)
		var err error
		switch f.engine {
		default:
			r = newReader.(func(io.Reader) io.Reader)(b)
		case "zlib":
			r, err = newReader.(func(r io.Reader) (io.ReadCloser, error))(b)
		case "flate":
			r = newReader.(func(r io.Reader) io.ReadCloser)(b)
		case "gzip":
			r, err = newReader.(func(r io.Reader) (*gzip.Reader, error))(b)
		case "lzw":
			// LZW requires special parameters for lzw
			r = newReader.(func(io.Reader, lzw.Order, int) (io.ReadCloser))(b, lzw.MSB, 8)
		}
		check(err)
		f.decompressed, err = ioutil.ReadAll(r)
		check(err)		
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

var Writers = map[string]interface{}{
 	"lzss": lz.NewWriter,
	"dmc": dmc.NewWriter,
	"mcc": mcc.NewWriter,
	"huffman": huffman.NewWriter,
	"arithmetic": arithmetic.NewWriter,
	"zlib": zlib.NewWriter,
	"flate": flate.NewWriter,
	"gzip": gzip.NewWriter,
	"lzw": lzw.NewWriter,
}

func (f *CompressedFile) Write(content []byte) (int, error) {
	var compressed []byte
	newWriter := Writers[f.engine]
	var b bytes.Buffer
	var w io.WriteCloser
	var err error
	switch f.engine {
	default:
		w = newWriter.(func(io.Writer) io.WriteCloser)(&b)
	case "zlib":
		w = newWriter.(func(r io.Writer) (*zlib.Writer))(&b)
	case "flate":
		w, err = newWriter.(func(w io.Writer, level int) (*flate.Writer, error))(&b, 9)
	case "gzip":
		w = newWriter.(func(w io.Writer) *gzip.Writer)(&b)
	case "lzw":
		// LZW requires special parameters for lzw
		w = newWriter.(func(w io.Writer, order lzw.Order, litWidth int) io.WriteCloser)(&b, lzw.MSB, 8)
	}
	check(err)
	w.Write(content)
	w.Close()
	compressed = b.Bytes()

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
    actualEntropy float32
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
		t.AppendHeader(table.Row{"engine", "time taken", "compression ratio", "actual entropy", "theoretical entropy", "lossless"})

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
			if results[j].lossless && results[i].lossless {
				return results[j].ratio > results[i].ratio
			} else if results[j].lossless && !results[i].lossless {
				return false
			} else if !results[j].lossless && results[i].lossless {
				return true
			} else {
				return results[j].ratio > results[i].ratio
			}
		})

		for _, result := range results {
			t.AppendRow([]interface{}{result.engine, result.timeTaken, fmt.Sprintf("%.2f%%", result.ratio), fmt.Sprintf("%.2f", result.actualEntropy), fmt.Sprintf("%.2f", result.entropy), result.lossless})
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
		tmpl.Execute(&b, struct{
			Tables template.HTML
			Created string
		}{Tables: template.HTML(html), Created: strconv.FormatInt(time.Now().Unix(), 10)})
		return b.String()
	} else {
		return ""
	}
}

func AsyncBenchmarkFile(resultChannel chan Result, wg *sync.WaitGroup, engine string, fileString string, suite bool) {
	defer wg.Done()

	errorHandler := func() {
		if r := recover(); r != nil {
			fmt.Printf("%s errored during execution, continuing\n", engine)
			fmt.Println("Err:", r)
			fmt.Println(string(debug.Stack()))
			fmt.Println("Continuing")
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
	result.timeTaken = fmt.Sprintf("%s", duration.Round(10 * time.Microsecond).String())

	fmt.Printf("%s finished benchmarking\n", engine)

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

	// if engine == "huffman" { fmt.Println(file.compressed) }

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

	symbolFrequencies = make(map[byte]int)
	for _, c := range []byte(file.compressed) {
		symbolFrequencies[c]++
	}
	total = len([]byte(file.compressed))
	freqs = make([]float64, len(symbolFrequencies))
	i = 0
	for _, freq := range symbolFrequencies {
		freqs[i] = float64(freq) / float64(total)
		i++
	}
	actualEntropy := float32(ent.Entropy(freqs, math.Log))

	if !suite {
		fmt.Printf("Lossless: %t\n", lossless)

		fmt.Printf("Original bytes: %v\n", len(fileContents))
		fmt.Printf("Compressed bytes: %v\n", len(file.compressed))
		if !lossless {
			fmt.Printf("Decompressed bytes: %v\n", len(file.decompressed))
		}
		fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
		fmt.Printf("Original Shannon entropy: %.2f\n", entropy)
		fmt.Printf("Compressed Shannon entropy: %.2f\n", actualEntropy)
	}
	return Result{engine, "", percentageDiff, actualEntropy, entropy, lossless, false}
}