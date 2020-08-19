package arithmetic

import (
	"errors"
)

// BitSlice represents a slice of bits (represented as bools)
type BitSlice []bool

// GetNextBit returns the next bit and a new bit slice with the first bit popped off.
// If there are no more bits, then the function returns 0.
func GetNextBit(bits BitSlice) (uint32, BitSlice) {
	if len(bits) <= 0 {
		return 0, bits
	}
	nextBit := bits[0]
	bits = bits[1:]
	if nextBit {
		return 1, bits
	}
	return 0, bits
}

// PushBits takes a BitSlice, a bit, and the number of pendingBits.
// PushBits returns the BitSlice with the bit and pendingBits pushed.
func PushBits(bits BitSlice, bit bool, pendingBits int) BitSlice {
	bits = append(bits, bit)
	bits = append(bits, PushBitsPending(pendingBits, bit)...)
	return bits
}

func PushBitsPending(pendingBits int, bit bool) BitSlice {
	additionalBits := make(BitSlice, pendingBits)
	switch bit {
	case false: // 0
		for i := 0; i < pendingBits; i++ {
			additionalBits[i] = true // 1
		}
	case true: // 1
		for i := 0; i < pendingBits; i++ {
			additionalBits[i] = false // 0
		}
	}
	return additionalBits
}

// Pack packs a BitSlice to fit within 8 bit bytes
func (bits BitSlice) Pack() BitSlice {
	var padded BitSlice

	bitsToAdd := 8 - (len(bits) % 8)
	for i := 0; i < bitsToAdd; i++ {
		if i == bitsToAdd-1 {
			padded = append(padded, true)
		} else {
			padded = append(padded, false)
		}
	}
	return append(padded, bits...)
}

// Unpack unpacks a BitSlice from it's byte representation into the original BitSlice
func (bits BitSlice) Unpack() BitSlice {
	for i := 0; i < len(bits); i++ {
		bit := bits[i]
		if bit == true {
			return bits[i+1:]
		}
	}
	panic("Couldn't unpack")
}

// ByteSize is the size of a byte in bits (typically 8)
const ByteSize = 8

func (bits BitSlice) AsByteSlice() (error, []byte) {
	if len(bits)%ByteSize != 0 {
		return errors.New("Bits are not packed, cannot convert to byte slice"), nil
	}

	bytes := make([]byte, len(bits)/ByteSize)

	for i := 0; i < (len(bits) / ByteSize); i++ {
		var byteInt int
		bools := bits[(i * ByteSize):((i + 1) * ByteSize)]
		for j := 0; j < ByteSize; j++ {
			byteInt <<= 1
			if bools[j] {
				byteInt += 1
			}
		}
		bytes[i] = byte(byteInt)
	}
	return nil, bytes
}

func FromByteSlice(bytes []byte) BitSlice {
	bits := make(BitSlice, len(bytes)*8)

	for i := 0; i < len(bytes); i++ {
		byteInt := int(bytes[i])
		for j := 0; j < ByteSize; j++ {
			bits[((i+1)*8)-j-1] = (byteInt & (1 << j)) != 0
		}
	}
	return bits
}
