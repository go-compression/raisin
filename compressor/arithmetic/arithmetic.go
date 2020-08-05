package arithmetic

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
)

type sortBytes []byte

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
	// symFreqsWhole = map[byte]float64{'3': 0.4, '2': 0.5, '1': 0.05, '0': 0.05}
	keys := make(sortBytes, 0)
	for k, _ := range symFreqsWhole {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	fmt.Printf("-------------\n")
	for i := len(keys) - 1; i >= 0; i-- {
		fmt.Printf("%s - %f\n", string(keys[i]), symFreqsWhole[keys[i]])
	}
	fmt.Printf("-------------\n")
	top, bot := encode(keys, symFreqsWhole, input)
	fmt.Println(bot, "-", top, string(input))
	binaryLocation := getRootBinaryPosition(top, bot)
	fmt.Println(binaryLocation)
	return []byte("compress")
}

func getRootBinaryPosition(targetTop float64, targetBot float64) string {
	return getBinaryPosition(targetTop, targetBot, 1, 0)
}

func getBinaryPosition(targetTop float64, targetBot float64, top float64, bot float64) string {
	if targetTop > top && targetBot <= bot {
		return ""
	}
	diff := top - bot
	targetHalfway := targetTop - ((targetTop - targetBot) / 2)
	halfwayPoint := top - (diff / 2)
	if halfwayPoint < targetHalfway {
		// Target range is above halfway point
		return "1" + getBinaryPosition(targetTop, targetBot, top, halfwayPoint)
	} else {
		// Target range is below halfway point
		return "0" + getBinaryPosition(targetTop, targetBot, halfwayPoint, bot)
	}
}

func encode(keys []byte, freqs map[byte]float64, input []byte) (top float64, bottom float64) {
	if len(input) == 0 {
		return -1, -1
	}

	// Pop first byte off the input
	encodeByte := input[0]
	input = input[1:]

	sec := getSection(keys, freqs, encodeByte)
	nextTop, nextBottom := encode(keys, freqs, input)
	size := nextTop - nextBottom
	for i := 0; i < sec; i++ {
		bottom += freqs[keys[i]]
	}
	// fmt.Println("before", bottom, "-", bottom + freqs[keys[sec]], string(encodeByte))
	if nextBottom != -1 { bottom = bottom + (freqs[keys[sec]] * nextBottom)}
	if nextTop != -1 {
		top = bottom + (freqs[keys[sec]] * size)
	} else {
		top = bottom + freqs[keys[sec]]
	}
	
	return top, bottom
}

func getSection(keys []byte, freqs map[byte]float64, input byte) int {
	for i, key := range keys {
		if key == input {
			return i
		}
	}
	return -1
}

func Decompress(input []byte) []byte {
	return []byte("decompress")
}


func (s sortBytes) Less(i, j int) bool {
    return s[i] < s[j]
}

func (s sortBytes) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

func (s sortBytes) Len() int {
    return len(s)
}

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) io.WriteCloser {
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

func NewReader(r io.Reader) io.Reader {
	z := new(Reader)
	z.r = r
	return z
}

func (r *Reader) Read(content []byte) (n int, err error) {
	if r.decompressed == nil {
		r.compressed, err = ioutil.ReadAll(r.r)
		if err != nil { return 0, err }
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