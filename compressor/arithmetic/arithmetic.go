package arithmetic

import (
	"fmt"
	"io"
	"io/ioutil"
)

func Compress(input []byte) []byte {
	symFreqs := make(map[byte]int)
	total := len(input)
	for _, c := range input {
		symFreqs[c]++
	}
	symFreqsWhole := make(map[byte]float64, len(symFreqs))
	for c, freq := range symFreqs {
		symFreqsWhole[c] = float64(freq) / float64(total)
	}
	fmt.Println(encode(symFreqsWhole, input))
	return []byte("compress")
}

func encode(freqs map[byte]float64, input []byte) (top float64, bottom float64) {
	if len(input) == 0 {
		return 0, 0
	}

	// Pop first byte off the input
	encodeByte := input[0]
	input = input[1:]

	var pos float64
	var prev float64
	for c, freq := range freqs {
		pos += freq
		if c == encodeByte {
			bottom = prev
			top = pos
			nextTop, nextBottom := encode(freqs, input)
			return top - nextTop/10, bottom + nextBottom/10
		}
		prev = pos
	}
	return 0, 0
}

func Decompress(input []byte) []byte {
	return []byte("decompress")
}


type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	z := new(Writer)
	z.w = w
	return z
}

func (writer *Writer) Write(data []byte) (n int, err error) {
	compressed := Compress(data)
	writer.w.Write(compressed)
	return len(compressed), nil
}

func (writer *Writer) Close() error {
	return nil
}

type Reader struct {
	r            io.Reader
	compressed   []byte
	decompressed []byte
	pos          int
}

func NewReader(r io.Reader) (*Reader, error) {
	z := new(Reader)
	z.r = r
	var err error
	z.compressed, err = ioutil.ReadAll(r)
	return z, err
}

func (r *Reader) Read(content []byte) (n int, err error) {
	if r.decompressed == nil {
		r.decompressed = Decompress(r.compressed)
	}
	bytesToWriteOut := len(r.decompressed[r.pos:])
	if len(content) < bytesToWriteOut {
		bytesToWriteOut = len(content)
	}
	for i := 0; i < bytesToWriteOut; i++ {
		content[i] = r.decompressed[r.pos:][i]
	}
	if len(r.decompressed[r.pos:]) <= len(content) {
		err = io.EOF
	} else {
		r.pos += len(content)
	}
	return bytesToWriteOut, err
}

func (r *Reader) Close() error {
	return nil
}
