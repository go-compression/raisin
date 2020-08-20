package engine

import (
	"bytes"
	flate "compress/flate"
	gzip "compress/gzip"
	lzw "compress/lzw"
	zlib "compress/zlib"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	ent "github.com/kzahedi/goent/discrete"
	arithmetic "github.com/mrfleap/raisin/compressor/arithmetic"
	dmc "github.com/mrfleap/raisin/compressor/dmc"
	huffman "github.com/mrfleap/raisin/compressor/huffman"
	lz "github.com/mrfleap/raisin/compressor/lz"
	mcc "github.com/mrfleap/raisin/compressor/mcc"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Engines is a slice of strings representing possible algorithms.
var Engines = [...]string{"all", "suite", "lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib", "arithmetic"}

// Suites is a map of strings to strings representing a suite name and the contained algorithms.
var Suites = map[string][]string{"all": Engines[2:], "suite": {"lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib", "arithmetic"}}

// CompressedFile is a struct used to read a compressed file or write to a compressed file.
type CompressedFile struct {
	CompressionEngine     string
	Compressed            []byte
	Decompressed          []byte
	pos                   int
	MaxSearchBufferLength int
}

// Readers represents a map of algorithm names to their NewReader interfaces.
var Readers = map[string]interface{}{
	"lzss":       lz.NewReader,
	"dmc":        dmc.NewReader,
	"mcc":        mcc.NewReader,
	"huffman":    huffman.NewReader,
	"arithmetic": arithmetic.NewReader,
	"zlib":       zlib.NewReader,
	"flate":      flate.NewReader,
	"gzip":       gzip.NewReader,
	"lzw":        lzw.NewReader,
}

func (f *CompressedFile) Read(content []byte) (int, error) {
	if f.Decompressed == nil {
		newReader := Readers[f.CompressionEngine]
		var r io.Reader
		var b io.Reader
		b = bytes.NewReader(f.Compressed)
		var err error
		switch f.CompressionEngine {
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
			r = newReader.(func(io.Reader, lzw.Order, int) io.ReadCloser)(b, lzw.MSB, 8)
		}
		check(err)
		f.Decompressed, err = ioutil.ReadAll(r)
		check(err)
	}
	bytesToWriteOut := len(f.Decompressed[f.pos:])
	if len(content) < bytesToWriteOut {
		bytesToWriteOut = len(content)
	}
	for i := 0; i < bytesToWriteOut; i++ {
		content[i] = f.Decompressed[f.pos:][i]
	}
	var err error
	if len(f.Decompressed[f.pos:]) <= len(content) {
		err = io.EOF
	} else {
		f.pos += len(content)
	}
	return bytesToWriteOut, err
}

// Writers represents a map of algorithm names to their NewWriter interfaces.
var Writers = map[string]interface{}{
	"lzss":       lz.NewWriter,
	"dmc":        dmc.NewWriter,
	"mcc":        mcc.NewWriter,
	"huffman":    huffman.NewWriter,
	"arithmetic": arithmetic.NewWriter,
	"zlib":       zlib.NewWriter,
	"flate":      flate.NewWriter,
	"gzip":       gzip.NewWriter,
	"lzw":        lzw.NewWriter,
}

func (f *CompressedFile) Write(content []byte) (int, error) {
	var compressed []byte
	newWriter := Writers[f.CompressionEngine]
	var b bytes.Buffer
	var w io.WriteCloser
	var err error
	switch f.CompressionEngine {
	default:
		w = newWriter.(func(io.Writer) io.WriteCloser)(&b)
	case "zlib":
		w = newWriter.(func(r io.Writer) *zlib.Writer)(&b)
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

	f.Compressed = append(f.Compressed, compressed...)
	return len(compressed), nil
}

// GetCompressedFileFromPath takes a path variable and returns a CompressedFile object or an error.
func GetCompressedFileFromPath(path string) (CompressedFile, error) {
	var cf CompressedFile
	fileContents, err := ioutil.ReadFile(path)
	cf = CompressedFile{Compressed: fileContents}
	return cf, err
}

// CompressFiles takes a set of compression algorithms as a string and multiple file paths as a slice and writes out the files in the same path with the extension appended.
func CompressFiles(algorithms []string, files []string, extension string) {
	for _, file := range files {
		CompressFile(algorithms, file, file+extension)
	}
}

// CompressFile takes a set of compression algorithms as a string and a path to a file and writes out the file  in the same path with .compressed appended to the end.
func CompressFile(algorithms []string, path string, output string) []byte {
	fileContents, err := ioutil.ReadFile(path)
	check(err)
	fmt.Printf("Compressing...\n")

	compressed := compress(fileContents, algorithms)

	err = ioutil.WriteFile(output, compressed, 0644)

	fmt.Printf("Original bytes: %v\n", len(fileContents))
	fmt.Printf("Compressed bytes: %v\n", len(compressed))
	percentageDiff := float32(len(compressed)) / float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)

	return compressed
}

// DecompressFiles takes a set of compression algorithms as a string and and multiple file paths as a slice and writes out the decompressed files in the same path with .decompressed appended to the end.
func DecompressFiles(algorithms []string, files []string, extension string) {
	for _, file := range files {
		path := file + extension
		if strings.TrimSpace(extension) == "" {
			ext := filepath.Ext(file)
			path = strings.TrimSuffix(file, ext)
		}
		DecompressFile(algorithms, file, path)
	}
}

// DecompressFile takes a set of compression algorithms as a string and a path to a file and writes out the decompressed file in the same path with .decompressed appended to the end.
func DecompressFile(algorithms []string, path string, output string) []byte {
	fileContents, err := ioutil.ReadFile(path)
	check(err)
	fmt.Printf("Decompressing...\n")

	decompressed := decompress(fileContents, algorithms)

	err = ioutil.WriteFile(output, decompressed, 0644)
	check(err)

	return decompressed
}

// Result is an intermediary object used to represent the benchmarked results of a certain file and algorithm.
type Result struct {
	CompressionEngine string
	TimeTaken         string
	Ratio             float32
	ActualEntropy     float32
	Entropy           float64
	Lossless          bool
	Failed            bool
}

// BenchmarkSuite takes a set of files and algorithms and returns the result as an html table if generateHTML is set.
// The result is also outputted to stdout.
func BenchmarkSuite(files []string, algorithms [][]string, generateHTML bool) (string, []Result) {
	var html string
	var allResults []Result
	timeout := 1 * time.Minute

	for i, fileString := range files {
		fmt.Printf("Compressing file %d/%d - %s\n", i+1, len(files), fileString)
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

		for _, algorithmsInLayer := range algorithms {
			algorithmsString := strings.Join(algorithmsInLayer[:], ",")
			fmt.Println("Benchmarking", algorithmsString)

			resultChannel := make(chan Result, 1)
			resultChans[algorithmsString] = resultChannel

			wg.Add(1)
			go AsyncBenchmarkFile(resultChannel, &wg, algorithmsInLayer, fileString)
		}

		waitTimeout(&wg, timeout)

		for compressionEngineName, resultChan := range resultChans {
			select {
			case result := <-resultChan:
				if result.Failed {
					failedResults = append(failedResults, result)
				} else {
					results = append(results, result)
				}
			default:
				result := Result{}
				result.CompressionEngine = compressionEngineName
				result.TimeTaken = fmt.Sprintf(">%s", timeout)
				result.Lossless = false
				result.Failed = true
				failedResults = append(failedResults, result)
			}
		}

		sort.Slice(results, func(i, j int) bool {
			if results[j].Lossless && results[i].Lossless {
				return results[j].Ratio > results[i].Ratio
			} else if results[j].Lossless && !results[i].Lossless {
				return false
			} else if !results[j].Lossless && results[i].Lossless {
				return true
			} else {
				return results[j].Ratio > results[i].Ratio
			}
		})

		for _, result := range results {
			t.AppendRow([]interface{}{result.CompressionEngine, result.TimeTaken, fmt.Sprintf("%.2f%%", result.Ratio), fmt.Sprintf("%.2f", result.ActualEntropy), fmt.Sprintf("%.2f", result.Entropy), result.Lossless})
			allResults = append(allResults, result)
		}

		t.AppendSeparator()
		for _, result := range failedResults {
			t.AppendRow([]interface{}{result.CompressionEngine, result.TimeTaken, "DNF", "DNF", "DNF", result.Lossless})
			allResults = append(allResults, result)
		}
		t.AppendSeparator()
		t.AppendRow(table.Row{"File", fileString, "Size", ByteCountSI(fileSize)})

		t.Render()
		if generateHTML {
			html = html + "<br>" + t.RenderHTML()
		}
	}
	if generateHTML {
		tmpl := template.Must(template.ParseFiles("templates/benchmark.html"))
		var b bytes.Buffer
		tmpl.Execute(&b, struct {
			Tables  template.HTML
			Created string
		}{Tables: template.HTML(html), Created: strconv.FormatInt(time.Now().Unix(), 10)})
		return b.String(), allResults
	}
	return "", allResults
}

// AsyncBenchmarkFile takes a channel to push the result, a waitgroup, engines, a file string and runs the benchmark.
// The function will push the result to the channel or push a failed result if it is able to catch an error during execution.
func AsyncBenchmarkFile(resultChannel chan Result, wg *sync.WaitGroup, compressionEngines []string, fileString string) {
	defer wg.Done()

	algorithmsString := strings.Join(compressionEngines[:], ",")

	errorHandler := func() {
		if r := recover(); r != nil {
			fmt.Printf("%s errored during execution, continuing\n", algorithmsString)
			fmt.Println("Err:", r)
			fmt.Println(string(debug.Stack()))
			fmt.Println("Continuing")
			result := Result{}
			result.CompressionEngine = algorithmsString
			result.TimeTaken = "failed"
			result.Lossless = false
			result.Failed = true
			resultChannel <- result
		}
	}

	defer errorHandler()
	start := time.Now()
	result := BenchmarkFile(compressionEngines, fileString, NewSuiteSettings())
	duration := time.Since(start)
	result.TimeTaken = fmt.Sprintf("%s", duration.Round(10*time.Microsecond).String())

	fmt.Printf("%s finished benchmarking\n", algorithmsString)

	resultChannel <- result
}

// Settings represents an object that can be used to modify the settings when benchmarking files with BenchmarkFile
type Settings struct {
	WriteOutFiles bool
	PrintStats    bool
	PrintStatus   bool
}

// NewSuiteSettings returns common settings for a testing suite as a Settings object
func NewSuiteSettings() Settings {
	s := Settings{}
	s.PrintStatus = true
	return s
}

// BenchmarkFile takes a set of algorithms, a file path, and a settings object.
// It benchmarks the file and returns the result as a Result object.
func BenchmarkFile(algorithms []string, fileString string, settings Settings) Result {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)

	algorithmsString := strings.Join(algorithms[:], ",")

	if settings.PrintStatus {
		fmt.Printf("%s Compressing...\n", algorithmsString)
	}

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

	start := time.Now()

	content := fileContents

	content = compress(content, algorithms)

	if settings.WriteOutFiles {
		var compressedFilePath = filepath.Base(fileString) + ".compressed"
		err = ioutil.WriteFile(compressedFilePath, content, 0644)
	}

	compressed := content

	if settings.PrintStatus {
		fmt.Printf("%s Decompressing...\n", algorithmsString)
	}

	content = decompress(content, algorithms)

	if settings.WriteOutFiles {
		var decompressedFilePath = filepath.Base(fileString) + ".decompressed"
		err = ioutil.WriteFile(decompressedFilePath, content, 0644)
		check(err)
	}

	decompressed := content

	duration := time.Since(start)

	lossless := reflect.DeepEqual(fileContents, decompressed)
	percentageDiff := float32(len(compressed)) / float32(len(fileContents)) * 100
	entropy := ent.Entropy(freqs, math.Log)

	symbolFrequencies = make(map[byte]int)
	for _, c := range content {
		symbolFrequencies[c]++
	}
	total = len(compressed)
	freqs = make([]float64, len(symbolFrequencies))
	i = 0
	for _, freq := range symbolFrequencies {
		freqs[i] = float64(freq) / float64(total)
		i++
	}
	actualEntropy := float32(ent.Entropy(freqs, math.Log))

	timeTaken := fmt.Sprintf("%s", duration.Round(10*time.Microsecond).String())

	if settings.PrintStats {
		fmt.Printf("Lossless: %t\n", lossless)

		fmt.Printf("Original bytes: %v\n", len(fileContents))
		fmt.Printf("Compressed bytes: %v\n", len(compressed))
		if !lossless {
			fmt.Printf("Decompressed bytes: %v\n", len(decompressed))
		}
		fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
		fmt.Printf("Original Shannon entropy: %.2f\n", entropy)
		fmt.Printf("Compressed Shannon entropy: %.2f\n", actualEntropy)
		fmt.Printf("Time taken: %s\n", timeTaken)
	}
	return Result{algorithmsString, timeTaken, percentageDiff, actualEntropy, entropy, lossless, false}
}

func compress(content []byte, algorithms []string) []byte {
	for _, algorithm := range algorithms {
		file := CompressedFile{MaxSearchBufferLength: 4096}
		file.CompressionEngine = algorithm
		file.Write(content)

		content = file.Compressed
	}
	return content
}

func decompress(content []byte, algorithms []string) []byte {
	for i := len(algorithms) - 1; i >= 0; i-- {
		algorithm := algorithms[i]
		file := CompressedFile{}
		file.Compressed = content
		file.CompressionEngine = algorithm

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

		content = file.Decompressed
	}
	return content
}
