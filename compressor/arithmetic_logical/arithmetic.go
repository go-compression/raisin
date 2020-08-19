package arithmetic_logical

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
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

func (b bitString) AsByteSlice() []byte {
	var out []byte
	var str string

	for i := len(b); i > 0; i -= 8 {
		if i-8 < 0 {
			str = string(b[0:i])
		} else {
			str = string(b[i-8 : i])
		}
		v, err := strconv.ParseUint(str, 2, 8)
		if err != nil {
			panic(err)
		}
		out = append([]byte{byte(v)}, out...)
	}
	return out
}

func Compress(input []byte) []byte {
	// freqs := buildFreqTable(input)
	freqs := map[byte]float64{'H': 0.2, 'e': 0.2, 'l': 0.4, 'o': 0.2} // testing
	// freqs := map[byte]float64{'H': 0.5, 'I': 0.5} // testing
	keys := buildKeys(freqs)
	printFreqs(freqs, keys)

	// bits := encode(keys, freqs, input)
	// fmt.Println(bits, len(bits))
	// bits = bits.pack()
	bits, top, bot := encode(true, keys, freqs, input)
	fmt.Println(bot, "-", top, string(input))
	binaryLocation := bits + getRootBinaryPosition(top, bot) + "1"
	fmt.Println(binaryLocation)

	bot, top = bitsToRange(binaryLocation)
	fmt.Println("Result", bot, "-", top, string(binaryLocation))

	_, shouldBeTop, shouldBeBot := encode(false, keys, freqs, input)
	fmt.Println("Should be", shouldBeBot, "-", shouldBeTop, string(input))
	fmt.Println("Within range:", (shouldBeBot <= bot && shouldBeTop > top))

	return bitString(binaryLocation).pack().AsByteSlice()
}

func Decompress(input []byte) []byte {
	inputString := fmt.Sprintf("%08b", input)
	bits := bitString(strings.Replace(inputString[1:len(inputString)-1], " ", "", -1))
	bits = bits.unpack()
	fmt.Println(string(bits), len(bits))

	freqs := map[byte]float64{'H': 0.2, 'e': 0.2, 'l': 0.4, 'o': 0.2} // testing
	// freqs := map[byte]float64{'H': 0.5, 'I': 0.5} // testing
	keys := buildKeys(freqs)
	printFreqs(freqs, keys)

	output := decode(keys, freqs, bits)
	fmt.Println(string(output))

	return output
}

func decode(keys []byte, freqs map[byte]float64, bits bitString) []byte {
	bot, top := float64(0), float64(1)
	var output []byte

	for i := 0; i < len(bits); i++ {
		bit := bits[i]

		// Enter half based on bit
		midpoint := bot + ((top - bot) / 2)
		if bit == '1' {
			// Top half
			bot = midpoint
		} else {
			// Bottom half
			top = midpoint
		}

		var char byte
		found := true
		for found {
			// var charTop, charBot float64
			char, found, _, _ = getCharFromRange(top, bot, keys, freqs)
			if found {
				output = append(output, char)
				fmt.Println(string(output))
				top = 1
				bot = 0
			}
		}
	}
	return output
}

func getCharFromRange(top float64, bot float64, keys []byte, freqs map[byte]float64) (byte, bool, float64, float64) {
	var byteTop, byteBot float64
	for i := 0; i < len(keys); i++ {
		char := keys[i]
		freq := freqs[char]
		byteTop = byteBot + freq
		if byteBot <= bot && byteTop > top {
			// Found char, scale up
			return char, true, byteTop, byteBot
		}
		byteBot += freq
	}
	return 0, false, 0, 0
}

func findChar(top float64, bot float64, keys []byte, freqs map[byte]float64) (float64, float64, []byte) {
	var byteTop, byteBot float64
	var output []byte
	for i := 0; i < len(keys); i++ {
		char := keys[i]
		freq := freqs[char]
		byteTop = byteBot + freq
		if byteBot <= bot && byteTop > top {
			// Found char, scale up
			fmt.Println(char)
			output = append(output, char)
			top *= 2
			bot *= 2
			var chars []byte
			top, bot, chars = findChar(top, bot, keys, freqs)
			output = append(output, chars...)
		}
		byteBot += freq
	}
	return top, bot, output
}

func encode(finite bool, keys []byte, freqs map[byte]float64, input []byte) (string, float64, float64) {
	var encodeByte byte
	var bits string
	var pending int
	top, bottom := float64(1), float64(0)

	freqsPassed := float64(1)

	for i := 0; i < len(input); i++ {
		encodeByte = input[i]

		var byteTop, byteBot float64

		sec := getSection(keys, freqs, encodeByte)
		for i := 0; i < sec; i++ {
			byteBot += freqs[keys[i]]
		}
		byteTop = byteBot + freqs[keys[sec]]

		size := freqsPassed * (byteTop - byteBot)

		bottom = bottom + (freqsPassed * byteBot)
		top = bottom + size

		freqsPassed *= freqs[keys[sec]]

		for finite {
			if bottom >= 0.5 {
				bits += "1" + getBitsPending(pending, "1")
				top = (top - 0.5) * 2
				bottom = (bottom - 0.5) * 2
				freqsPassed *= 2
				// fmt.Println("1" + getBitsPending(pending, "1"), "Scaled to", bottom, top)
				pending = 0
			} else if top < 0.5 {
				bits += "0" + getBitsPending(pending, "0")
				top *= 2
				bottom *= 2
				freqsPassed *= 2
				// fmt.Println("1" + getBitsPending(pending, "0"), "Scaled to", bottom, top)
				pending = 0
			} else if (bottom >= 0.25) && (top < 0.75) {
				top = (top - 0.25) * 2
				bottom = (bottom - 0.25) * 2
				freqsPassed *= 2
				// fmt.Println("-", "Scaled to", bottom, top)
				pending++
			} else {
				break
			}
			// fmt.Println()
		}
	}

	return bits, top, bottom
}

func getBitsPending(pendingBits int, bit string) (additionalBits string) {
	switch bit {
	case "0":
		for i := 0; i < pendingBits; i++ {
			additionalBits += "1"
		}
	case "1":
		for i := 0; i < pendingBits; i++ {
			additionalBits += "0"
		}
	default:
		panic("Invalid bit")
	}
	return
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
	} else {
		// Target range is below halfway point
		return "0" + getBinaryPosition(targetTop, targetBot, halfwayPoint, bot)
	}
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

func (r *Reader) Close() error {
	return nil
}
