package arithmetic

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
		if i == bitsToAdd - 1 {
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
			return bits[i + 1:]
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
	// freqs := map[byte]float64{'H': 0.2, 'e': 0.2, 'l': 0.4, 'o': 0.2} // testing
	// freqs := map[byte]float64{'H': 0.5, 'I': 0.5} // testing
	// keys := buildKeys(freqs)
	// printFreqs(freqs, keys)
	
	// bits := encode(keys, freqs, input)
	// fmt.Println(bits, len(bits))
	// bits = bits.pack()
	bits := encode(input)
	// fmt.Println(bits) 

	return bits.pack().AsByteSlice()
}

func Decompress(input []byte) []byte {
	inputString := fmt.Sprintf("%08b", input)
	bits := bitString(strings.Replace(inputString[1:len(inputString)-1], " ", "", -1))
	bits = bits.unpack()
	// fmt.Println(string(bits), len(bits))

	// freqs := map[byte]float64{'H': 0.2, 'e': 0.2, 'l': 0.4, 'o': 0.2} // testing
	// freqs := map[byte]float64{'H': 0.5, 'I': 0.5} // testing
	// keys := buildKeys(freqs)
	// printFreqs(freqs, keys)

	output := decode(bits)
	// fmt.Println(string(output))
	
	return output
}

const (
	MAX_CODE = 0xffff
	ONE_FOURTH = 0x4000
	ONE_HALF = 2 * ONE_FOURTH
	THREE_FOURTHS = 3 * ONE_FOURTH
	CODE_VALUE_BITS = 16
)


func decode(bits bitString) []byte {
	var output []byte
	var high, low, value uint32
	high = MAX_CODE
	bits = bits + "10"
	fmt.Println(string(bits)[:CODE_VALUE_BITS])

	val, err := strconv.ParseInt(string(bits)[:CODE_VALUE_BITS], 2, 32)
	if err != nil { panic(err) }
	value = uint32(val)
	bits = bits[CODE_VALUE_BITS:]

	model := newModel()

	for {
		difference := high - low + 1
		scaled_value := ((value - low + 1) * model.getCount() - 1) / difference

		char, lower, upper, count := model.getChar(scaled_value)

		if ( char == 256 ) {
			// EOF char 
			break;
		}

		output = append(output, byte(char))

		if count == 0 {
			fmt.Printf("Uh oh")
			char, lower, upper, count = model.getChar(scaled_value)
		}
		
		high = low + (difference*upper)/count - 1
		low = low + (difference*lower)/count
		for {
			if high < ONE_HALF {
				//do nothing, bit is a zero
			} else if low >= ONE_HALF {
				value -= ONE_HALF;  //subtract one half from all three code values
				low -= ONE_HALF;
				high -= ONE_HALF;
			} else if low >= ONE_FOURTH && high < THREE_FOURTHS {
				value -= ONE_FOURTH;
				low -= ONE_FOURTH;
				high -= ONE_FOURTH;
			} else {
				break
			}
			low <<= 1;
			high <<= 1;
			high++;
			value <<= 1;
			
			var nextBit uint32
			nextBit, bits = getNextBit(bits)
			value += nextBit
			if value > 65536 {
				fmt.Println("uh oh")
			}
			// fmt.Println(strconv.FormatUint(uint64(value), 2), len(strconv.FormatUint(uint64(value), 2)))
		}
	}
	return output
}

func getNextBit(bits bitString) (uint32, bitString) {
	if len(bits) <= 0 { return 0, bits }
	next_bit := bits[0]
	bits = bits[1:]
	if next_bit == '1' { return 1, bits }
	return 0, bits
}

func encode(input []byte) bitString {
	var toEncode int
	var bits string
	var pendingBits int

	var high, low uint32
	high = MAX_CODE
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
			if high < ONE_HALF {
				// Lower half
				bits += "0" + getBitsPending(pendingBits, "0")
				pendingBits = 0
			} else if low >= ONE_HALF {
				// Upper half
				bits += "1" + getBitsPending(pendingBits, "1")
				pendingBits = 0
				// fmt.Println(strconv.FormatUint(uint64(high), 2))
			} else if (low >= ONE_FOURTH && high < THREE_FOURTHS) {
				pendingBits++
				low -= ONE_FOURTH; 
          		high -= ONE_FOURTH 
			} else {
				break
			}
			high <<= 1
			high++ 
			low <<= 1
			high &= MAX_CODE
			low &= MAX_CODE
		}
	}
	return bitString(bits)
}

const denom = uint32(100)
const denomFloat = float64(denom)

type Model struct {
	cumulative_frequencies []int
}

func newModel() *Model {
	model := Model{make([]int, 258)}
	for i := 0; i < 258; i++ {
		model.cumulative_frequencies[i] = i;
	}
	return &model
}

func (model *Model) update(input int) {
	for i := input + 1 ; i < 258 ; i++ {
		model.cumulative_frequencies[i]++;
	}
}

func (model *Model) getProbability(input int) (uint32, uint32, uint32) {
	lower, upper, count := model.cumulative_frequencies[input], model.cumulative_frequencies[input+1], model.cumulative_frequencies[257]
    model.update(input);
    return uint32(lower), uint32(upper), uint32(count)
}

func (model *Model) getCount() uint32 {
	return uint32(model.cumulative_frequencies[257])
}

func (model *Model) getChar(scaled_value uint32) (int, uint32, uint32, uint32) {
	var char int
	for i := 0; i < 257; i++ {
      if scaled_value < uint32(model.cumulative_frequencies[i+1]) {
        char = i
        lower, upper, count := model.cumulative_frequencies[i], model.cumulative_frequencies[i+1], model.cumulative_frequencies[257]
        model.update(char)
        return char, uint32(lower), uint32(upper), uint32(count)
	  }
	}
	return ' ', 0, 0, 0
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
	for k, _ := range freqs {
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