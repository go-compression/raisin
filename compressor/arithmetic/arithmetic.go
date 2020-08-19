package arithmetic

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	// "strings"
	// "reflect"
)

type sortBytes []byte
type bitString string

func (bits bitString) pack() bitString {
	var padded bitString

	bitsToAdd := 8 - (len(bits) % 8)
	for i := 0; i < bitsToAdd; i++ {
		if i == bitsToAdd-1 {
			padded += "1"
		} else {
			padded += "0"
		}
	}
	return padded + bits
}

func (bits bitString) unpack() bitString {
	for i := 0; i < len(bits); i++ {
		bit := bits[i]
		if bit == '1' {
			return bits[i+1:]
		}
	}
	panic("Couldn't unpack")
}

func (bits bitString) AsByteSlice() []byte {
	var out []byte
	var str string

	for i := len(bits); i > 0; i -= 8 {
		if i-8 < 0 {
			str = string(bits[0:i])
		} else {
			str = string(bits[i-8 : i])
		}
		v, err := strconv.ParseUint(str, 2, 8)
		if err != nil {
			panic(err)
		}
		out = append([]byte{byte(v)}, out...)
	}
	return out
}

// Compress takes a slice of bytes and returns a slice of bytes representing the compressed stream
func Compress(input []byte) []byte {

	bits := encode(input)

	err, bytes := bits.Pack().AsByteSlice()
	if err != nil {
		panic(err)
	}
	return bytes
}

// Decompress takes a slice of bytes and returns a slice of bytes representing the decompressed stream
func Decompress(input []byte) []byte {
	bits := FromByteSlice(input).Unpack()

	output := decode(bits)

	return output
}

const (
	maxCode       = 0xffff
	oneFourth     = 0x4000
	oneHalf       = 2 * oneFourth
	threeFourths  = 3 * oneFourth
	codeValueBits = 16
	maxFreq       = 16383
)

func decode(bits BitSlice) []byte {
	var output []byte
	var high, low, value uint32
	high = maxCode
	bits = append(bits, []bool{true, false}...)

	for i := 0; i < codeValueBits; i++ {
		value <<= 1
		if bits[i] {
			value++
		}
	}

	// val, err := strconv.ParseInt(bits[:codeValueBits], 2, 32)
	// if err != nil { panic(err) }
	// value = uint32(val)
	bits = bits[codeValueBits:]

	model := newModel()

	for {
		difference := high - low + 1
		scaledValue := ((value-low+1)*model.getCount() - 1) / difference

		char, lower, upper, count := model.getChar(scaledValue)

		if char == 256 {
			// EOF char
			break
		}

		output = append(output, byte(char))

		// if count == 0 {
		// 	fmt.Printf("Uh oh")
		// 	char, lower, upper, count = model.getChar(scaled_value)
		// }

		high = low + (difference*upper)/count - 1
		low = low + (difference*lower)/count
		for {
			if high < oneHalf {
				//do nothing, bit is a zero
			} else if low >= oneHalf {
				value -= oneHalf //subtract one half from all three code values
				low -= oneHalf
				high -= oneHalf
			} else if low >= oneFourth && high < threeFourths {
				value -= oneFourth
				low -= oneFourth
				high -= oneFourth
			} else {
				break
			}
			low <<= 1
			high <<= 1
			high++
			value <<= 1

			var nextBit uint32
			nextBit, bits = GetNextBit(bits)
			value += nextBit
			// if value > 65536 {
			// 	fmt.Println("uh oh")
			// }
			// fmt.Println(strconv.FormatUint(uint64(value), 2), len(strconv.FormatUint(uint64(value), 2)))
		}
	}
	return output
}

func getNextBit(bits bitString) (uint32, bitString) {
	if len(bits) <= 0 {
		return 0, bits
	}
	nextBit := bits[0]
	bits = bits[1:]
	if nextBit == '1' {
		return 1, bits
	}
	return 0, bits
}

func encode(input []byte) BitSlice {
	var toEncode int
	var bits BitSlice
	var pendingBits int

	var high, low uint32
	high = maxCode
	model := newModel()

	inputChars := make([]int, len(input))

	for i := 0; i < len(input); i++ {
		inputChars[i] = int(input[i])
	}

	// Add EOF character to end of input
	inputChars = append(inputChars, 256)

	for i := 0; i < len(inputChars); i++ {
		toEncode = inputChars[i]

		difference := (high - low) + 1
		lower, upper, count := model.getProbability(toEncode)
		high = low + (difference * upper / count) - 1
		low = low + (difference * lower / count)
		for {
			if high < oneHalf {
				// Lower half
				bits = PushBits(bits, false, pendingBits)
				pendingBits = 0
			} else if low >= oneHalf {
				// Upper half
				bits = PushBits(bits, true, pendingBits)
				pendingBits = 0
				// fmt.Println(strconv.FormatUint(uint64(high), 2))
			} else if low >= oneFourth && high < threeFourths {
				pendingBits++
				low -= oneFourth
				high -= oneFourth
			} else {
				break
			}
			high <<= 1
			high++
			low <<= 1
			high &= maxCode
			low &= maxCode
		}
	}
	return bits
}

const denom = uint32(100)
const denomFloat = float64(denom)

// Model represents frequency tables for characters in a byte
type Model struct {
	cumulativeFrequencies []int
	frozen                bool
}

func newModel() *Model {
	model := Model{make([]int, 258), false}
	for i := 0; i < 258; i++ {
		model.cumulativeFrequencies[i] = i
	}
	return &model
}

func (model *Model) update(input int) {
	for i := input + 1; i < 258; i++ {
		model.cumulativeFrequencies[i]++
	}
	if model.cumulativeFrequencies[257] >= maxFreq {
		model.frozen = true
		fmt.Println("FROZEN")
	}
}

func (model *Model) getProbability(input int) (uint32, uint32, uint32) {
	lower, upper, count := model.cumulativeFrequencies[input], model.cumulativeFrequencies[input+1], model.cumulativeFrequencies[257]
	if !model.frozen {
		model.update(input)
	}
	return uint32(lower), uint32(upper), uint32(count)
}

func (model *Model) getCount() uint32 {
	return uint32(model.cumulativeFrequencies[257])
}

func (model *Model) getChar(scaledValue uint32) (int, uint32, uint32, uint32) {
	var char int
	for i := 0; i < 257; i++ {
		if scaledValue < uint32(model.cumulativeFrequencies[i+1]) {
			char = i
			lower, upper, count := model.cumulativeFrequencies[i], model.cumulativeFrequencies[i+1], model.cumulativeFrequencies[257]
			if !model.frozen {
				model.update(char)
			}
			return char, uint32(lower), uint32(upper), uint32(count)
		}
	}
	return ' ', 0, 0, 0
}

func bitsToRange(bits string) (float64, float64) {
	bot, top := float64(0), float64(1)

	for i := 0; i < len(bits); i++ {
		bit := bits[i]

		midpoint := bot + ((top - bot) / 2)

		if bit == '1' {
			// Top half
			bot = midpoint
		} else {
			// Bottom half
			top = midpoint
		}
	}

	return bot, top
}

func buildFreqTable(input []byte) map[byte]float64 {
	symFreqs := make(map[byte]int)
	total := len(input)
	for _, c := range input {
		symFreqs[c]++
	}
	symFreqsWhole := make(map[byte]float64, len(symFreqs))
	for c, freq := range symFreqs {
		symFreqsWhole[c] = float64(freq) / float64(total)
	}
	return symFreqsWhole
}

func buildKeys(freqs map[byte]float64) sortBytes {
	keys := make(sortBytes, 0)
	for k := range freqs {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	return keys
}

func printFreqs(freqs map[byte]float64, keys sortBytes) {
	fmt.Printf("-------------\n")
	for i := len(keys) - 1; i >= 0; i-- {
		fmt.Printf("%s - %f\n", string(keys[i]), freqs[keys[i]])
	}
	fmt.Printf("-------------\n")
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
	}
	// Target range is below halfway point
	return "0" + getBinaryPosition(targetTop, targetBot, halfwayPoint, bot)
}

func getSection(keys []byte, freqs map[byte]float64, input byte) int {
	for i, key := range keys {
		if key == input {
			return i
		}
	}
	return -1
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

// Writer takes an io.Writer to write to when compressing
type Writer struct {
	w io.Writer
}

// NewWriter creates an io.WriteCloser object with an io.Writer
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

// Close only exists to satisfy the io.WriteCloser interface
func (writer *Writer) Close() error {
	return nil
}

// Reader takes an io.Reader to read from when decompressing
type Reader struct {
	r            io.Reader
	compressed   []byte
	decompressed []byte
	pos          int
}

// NewReader creates an io.Reader object with an io.Reader
func NewReader(r io.Reader) io.Reader {
	z := new(Reader)
	z.r = r
	return z
}

func (r *Reader) Read(content []byte) (n int, err error) {
	if r.decompressed == nil {
		r.compressed, err = ioutil.ReadAll(r.r)
		if err != nil {
			return 0, err
		}
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

// Close only exists to satisfy the io.WriteCloser interface
func (r *Reader) Close() error {
	return nil
}
