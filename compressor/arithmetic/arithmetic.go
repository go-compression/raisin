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
	// symFreqsWhole = map[byte]float64{'0': 0.05, '1': 0.05, '2': 0.5, '3': 0.4}
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
	bits, top, bot := encodeLoop(true, keys, symFreqsWhole, input)
	fmt.Println(bot, "-", top, string(input))
	binaryLocation := bits + getRootBinaryPosition(top, bot)
	fmt.Println(binaryLocation) 

	bot, top = bitsToRange(binaryLocation)
	fmt.Println("Result", bot, "-", top, string(binaryLocation))

	_, shouldBeTop, shouldBeBot := encodeLoop(false, keys, symFreqsWhole, input)
	fmt.Println("Should be", shouldBeBot, "-", shouldBeTop, string(input))
	fmt.Println("Within range:", (shouldBeBot <= bot && shouldBeTop > top))
	// binaryLocation := getRootBinaryPosition(top, bot)
	// fmt.Println(binaryLocation)

	// output := string(bitsToBytes(binaryLocation, keys, symFreqsWhole))
	// fmt.Println(output)

	return []byte("compress")
}

func encodeLoop(finite bool, keys []byte, freqs map[byte]float64, input []byte) (string, float64, float64) {
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
		if finite { fmt.Println(byteBot, "-", byteTop, string(encodeByte)) }
		
		size := freqsPassed * (byteTop - byteBot)
		
		bottom = bottom + (freqsPassed * byteBot)
		top = bottom + size

		freqsPassed *= freqs[keys[sec]]

		if finite {
			fmt.Println(bottom, "-", top)
			if bottom >= 0.5 {
				bits += "1" + pendingBits(pending, "1")
				fmt.Println("Diff", top - bottom)
				top = (top - 0.5) * 2
				bottom = (bottom - 0.5) * 2
				freqsPassed *= 2
				fmt.Println("1" + pendingBits(pending, "1"), "Scaled to", bottom, top)
				fmt.Println("Diff", top - bottom)
				pending = 0
			} else if top < 0.5 {
				bits += "0" + pendingBits(pending, "0")
				top *= 2
				bottom *= 2
				freqsPassed *= 2
				fmt.Println("1" + pendingBits(pending, "0"), "Scaled to", bottom, top)
				pending = 0
			} else if (bottom >= 0.25 && bottom < 0.5) && (top <= 0.75 && top > 0.5) {
				top = (top - 0.25) * 2 
				bottom = (bottom - 0.25) * 2
				freqsPassed *= 2
				fmt.Println("-", "Scaled to", bottom, top)
				pending++
			}
			fmt.Println()
		}
	}

	return bits, top, bottom
}

func pendingBits(pendingBits int, bit string) (additionalBits string) {
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

func bitsToBytes(bits string, keys []byte, freqs map[byte]float64) []byte {
	var output []byte
	top, bot := float64(1), float64(0)

	for i := 0; i < len(bits); i++ {
		bit := bits[i]

		midpoint := bot + ((top - bot) / 2 )

		if bit == '1' {
			// Top half
			bot = midpoint
		} else {
			// Bottom half
			top = midpoint
		}
		
		var newBits []byte
		newBits, top, bot = getBitsFromRange(top, bot, keys, freqs)
		output = append(output, newBits...)
	}
	return output
}

func getBitsFromRange(top float64, bot float64, keys []byte, freqs map[byte]float64) ([]byte, float64, float64) {
	var bits []byte

	freqBot := float64(0)
	for _, c := range keys {
		freq := freqs[c]
		freqTop := freqBot + freq
		if freqBot <= bot && (freqTop + freqBot) > top {
			bits = append(bits, c) 
			top /= freq
			bot /= freq
			var newBits []byte
			newBits, top, bot = getBitsFromRange(top, bot, keys, freqs)
			bits = append(bits, newBits...)
		}
		freqBot += freq
	}

	return bits, top, bot
}

func bitsToRange(bits string) (float64, float64) {
	bot, top := float64(0), float64(1)

	for i := 0; i < len(bits); i++ {
		bit := bits[i]

		midpoint := bot + ((top - bot) / 2 )

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


func encode(keys []byte, freqs map[byte]float64, input []byte) (top float64, bottom float64) {
	if len(input) == 0 {
		return -1, -1
	}

	// Pop first byte off the input
	encodeByte := input[0]
	input = input[1:]

	sec := getSection(keys, freqs, encodeByte)
	for i := 0; i < sec; i++ {
		bottom += freqs[keys[i]]
	}
	top = bottom + freqs[keys[sec]]
	fmt.Println("before", bottom, "-", top, string(encodeByte))
	
	// fmt.Println(getRootBinaryPosition(top, bottom))

	half := ((top - bottom) / 2)
	middle := bottom + half

	if middle > 0.5 {
		fmt.Println("1")
		bottom = bottom - (top - bottom)
	} else {
		fmt.Println("0")
		top = top + (top - bottom)
	}

	nextTop, nextBottom := encode(keys, freqs, input)
	size := nextTop - nextBottom
	if nextBottom != -1 { bottom = bottom + (freqs[keys[sec]] * nextBottom)}
	if nextTop != -1 {
		top = bottom + (freqs[keys[sec]] * size)
	}
	
	return top, bottom
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