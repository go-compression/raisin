package engine

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
	"github.com/jedib0t/go-pretty/v6/table"
	"sort"
	"os"
	"html/template"
	"time"
	"sync"
	"strconv"
	"runtime/debug"
)

var Engines = [...]string{"all", "suite", "lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib", "arithmetic"}
var Suites = map[string][]string{"all": Engines[2:], "suite": []string{"lzss", "dmc", "huffman", "mcc", "flate", "gzip", "lzw", "zlib", "arithmetic"}}

type CompressedFile struct {
	CompressionEngine                string
	Compressed            []byte
	Decompressed          []byte
	pos                   int
	MaxSearchBufferLength int
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
			r = newReader.(func(io.Reader, lzw.Order, int) (io.ReadCloser))(b, lzw.MSB, 8)
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
	newWriter := Writers[f.CompressionEngine]
	var b bytes.Buffer
	var w io.WriteCloser
	var err error
	switch f.CompressionEngine {
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

	f.Compressed = append(f.Compressed, compressed...)
	return len(compressed), nil
}

func GetCompressedFileFromPath(path string) (CompressedFile, error) {
	var cf CompressedFile
	fileContents, err := ioutil.ReadFile(path)
	cf = CompressedFile{Compressed: fileContents}
	return cf, err
}

func CompressFile(compressionEngine string, fileString string) {
	fileContents, err := ioutil.ReadFile(fileString)
	check(err)
	fmt.Printf("Compressing...\n")

	file := CompressedFile{MaxSearchBufferLength: 4096}
	file.CompressionEngine = compressionEngine
	file.Write(fileContents)

	var compressedFilePath = filepath.Base(fileString) + ".compressed"
	err = ioutil.WriteFile(compressedFilePath, file.Compressed, 0644)

	fmt.Printf("Original bytes: %v\n", len(fileContents))
	fmt.Printf("Compressed bytes: %v\n", len(file.Compressed))
	percentageDiff := float32(len(file.Compressed)) / float32(len(fileContents)) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", percentageDiff)
}

func DecompressFile(compressionEngine string, fileString string) []byte {
	compressedFile, err := GetCompressedFileFromPath(fileString)
	compressedFile.CompressionEngine = compressionEngine
	check(err)
	fmt.Printf("Decompressing...\n")

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
	CompressionEngine string
	TimeTaken string 
	Ratio float32
    ActualEntropy float32
	Entropy  float64
	Lossless bool
	Failed bool
}

func BenchmarkSuite(files []string, algorithms [][]string, generateHtml bool) string {
	var html string
	timeout := 2 * time.Minute

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

		for _, algorithmsInLayer := range algorithms {
			algorithmsString := strings.Join(algorithmsInLayer[:], ",")
			fmt.Println("Benchmarking", algorithmsString)

			resultChannel := make(chan Result, 1)
			resultChans[algorithmsString] = resultChannel

			wg.Add(1)
			go AsyncBenchmarkFile(resultChannel, &wg, algorithmsInLayer, fileString, true)
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
		}

		t.AppendSeparator()
		for _, result := range failedResults {
			t.AppendRow([]interface{}{result.CompressionEngine, result.TimeTaken, "DNF", "DNF", "DNF", result.Lossless})
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

func AsyncBenchmarkFile(resultChannel chan Result, wg *sync.WaitGroup, compressionEngines []string, fileString string, suite bool) {
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
	result.TimeTaken = fmt.Sprintf("%s", duration.Round(10 * time.Microsecond).String())

	fmt.Printf("%s finished benchmarking\n", algorithmsString)

	resultChannel <- result
}

type Settings struct {
	WriteOutFiles bool
	PrintStats bool
	PrintStatus bool
}

func NewSuiteSettings() Settings {
	s := Settings{}
	s.PrintStatus = true
	return s
}

func BenchmarkFile(algorithms []string, fileString string, settings Settings) Result  {
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

	for _, algorithm := range algorithms {
		file := CompressedFile{MaxSearchBufferLength: 4096}
		file.CompressionEngine = algorithm
		file.Write(content)

		if settings.WriteOutFiles {
			var compressedFilePath = filepath.Base(fileString) + ".compressed"
			err = ioutil.WriteFile(compressedFilePath, file.Compressed, 0644)
		}

		content = file.Compressed
	}

	compressed := content

	if settings.PrintStatus {
		fmt.Printf("%s Decompressing...\n", algorithmsString)
	}

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
		
		if settings.WriteOutFiles {
			var decompressedFilePath = filepath.Base(fileString) + ".decompressed"
			err = ioutil.WriteFile(decompressedFilePath, stream, 0644)
			check(err)
		}
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

	timeTaken := fmt.Sprintf("%s", duration.Round(10 * time.Microsecond).String())

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