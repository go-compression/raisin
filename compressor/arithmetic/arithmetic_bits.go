package arithmetic

import (
	"strconv"
	"fmt"
	"math"
)

const (
	MAX_CODE = 0xffff
	ONE_FOURTH = 0x4000
	ONE_HALF = 2 * ONE_FOURTH
	THREE_FOURTHS = 3 * ONE_FOURTH
	CODE_VALUE_BITS = 16
)


func decode_bits(keys []byte, freqs map[byte]float64, bits bitString) []byte {
	var output []byte
	var high, low, value uint32
	high = MAX_CODE

	val, err := strconv.ParseInt(string(bits)[:CODE_VALUE_BITS], 2, 16)
	if err != nil { panic(err) }
	value = uint32(val)
	bits = bits[CODE_VALUE_BITS:]

	fmt.Println(strconv.FormatUint(uint64(value), 2), len(strconv.FormatUint(uint64(value), 2)))

	for i := 0; i < 8; i++ {
		difference := high - low + 1
		scaled_value := ((value - low + 1) * 257 - 1) / difference

		char, upper, lower, denom := getChar(keys, freqs, scaled_value)
		output = append(output, char)
		
		high = low + (difference*upper)/denom - 1
		low = low + (difference*lower)/denom
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
			next_bit := bits[0]
			bits = bits[1:]
			var next_bit_val uint32
			if next_bit == '1' {next_bit_val = 1}
			value += next_bit_val
			fmt.Println(strconv.FormatUint(uint64(value), 2), len(strconv.FormatUint(uint64(value), 2)))
		}
	}
	return output
}

func encode_bits(keys []byte, freqs map[byte]float64, input []byte) bitString {
	var encodeByte byte
	var bits string
	var pendingBits int

	var high, low uint32
	high = MAX_CODE
	model := newModel()

	for i := 0; i < len(input); i++ {
		encodeByte = input[i]

		difference := (high - low) + 1
		// lower, upper, denom := getProbability(keys, freqs, encodeByte)
		lower, upper, count := model.getProbability(encodeByte)
		high = low + (difference * upper / count) - 1
		low = low + (difference * lower / count)
		for {
			if high < ONE_HALF {
				// Lower half
				bits += "0" + getBitsPending(pendingBits, "0")
				pendingBits = 0
			} else if low >= ONE_HALF {
				// Upper half
				bits += "1" + getBitsPending(pendingBits, "0")
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

func (model *Model) update(input byte) {
	c := int(input)
	for i := c + 1 ; i < 258 ; i++ {
		model.cumulative_frequencies[i]++;
	}
}

func (model *Model) getProbability(input byte) (uint32, uint32, uint32) {
	c := int(input)
	lower, upper, count := model.cumulative_frequencies[c], model.cumulative_frequencies[c+1], model.cumulative_frequencies[257]
    model.update(input);
    return uint32(lower), uint32(upper), uint32(count)
}

func getProbability(keys []byte, freqs map[byte]float64, input byte) (uint32, uint32, uint32) {
	return uint32(input), uint32(input) + 1, 257

	var byteTop, byteBot float64

	sec := getSection(keys, freqs, input)
	for i := 0; i < sec; i++ {
		byteBot += freqs[keys[i]]
	}
	byteTop = byteBot + freqs[keys[sec]]

	return uint32(math.Round(byteBot*denomFloat)), uint32(math.Round(byteTop*denomFloat)), denom
}

func getCount() uint32 {
	return 257
}

func getChar(keys []byte, freqs map[byte]float64, scaled_value uint32) (byte, uint32, uint32, uint32) {
	return byte(int(scaled_value)), scaled_value + 1, scaled_value, 257
	// var byteTop, byteBot float64

	// for i := 0; i < len(keys); i++ {
	// 	byteTop = byteBot + freqs[keys[i]]

	// 	lower, upper := uint32(math.Round(byteBot*denomFloat)), uint32(math.Round(byteTop*denomFloat))
	// 	if upper > scaled_value && scaled_value >= lower {
	// 		// return keys[i], upper, lower, denom
	// 		return keys[i], upper, uint32(10), denom
	// 	}
	// 	byteBot += freqs[keys[i]]
	// }

	panic("Couldn't find char")
}
