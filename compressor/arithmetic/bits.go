package arithmetic

import (
	"errors"
)

type bitSlice []bool

func GetNextBit(bits bitSlice) (uint32, bitSlice) {
	if len(bits) <= 0 {
		return 0, bits
	}
	next_bit := bits[0]
	bits = bits[1:]
	if next_bit {
		return 1, bits
	}
	return 0, bits
}

func PushBits(bits bitSlice, bit bool, pendingBits int) bitSlice {
	bits = append(bits, bit)
	bits = append(bits, PushBitsPending(pendingBits, bit)...)
	return bits
}

func PushBitsPending(pendingBits int, bit bool) bitSlice {
	additionalBits := make(bitSlice, pendingBits)
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

func (bits bitSlice) Pack() bitSlice {
	var padded bitSlice

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

func (bits bitSlice) Unpack() bitSlice {
	for i := 0; i < len(bits); i++ {
		bit := bits[i]
		if bit == true {
			return bits[i+1:]
		}
	}
	panic("Couldn't unpack")
}

const BYTE_SIZE = 8

func (bits bitSlice) AsByteSlice() (error, []byte) {
	if len(bits)%BYTE_SIZE != 0 {
		return errors.New("Bits are not packed, cannot convert to byte slice"), nil
	}

	bytes := make([]byte, len(bits)/BYTE_SIZE)

	for i := 0; i < (len(bits) / BYTE_SIZE); i++ {
		var byteInt int
		bools := bits[(i * BYTE_SIZE):((i + 1) * BYTE_SIZE)]
		for j := 0; j < BYTE_SIZE; j++ {
			byteInt <<= 1
			if bools[j] {
				byteInt += 1
			}
		}
		bytes[i] = byte(byteInt)
	}
	return nil, bytes
}

func FromByteSlice(bytes []byte) bitSlice {
	bits := make(bitSlice, len(bytes)*8)

	for i := 0; i < len(bytes); i++ {
		byteInt := int(bytes[i])
		for j := 0; j < BYTE_SIZE; j++ {
			bits[((i+1)*8)-j-1] = (byteInt & (1 << j)) != 0
		}
	}
	return bits
}
