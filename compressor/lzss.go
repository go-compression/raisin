package compressor

import (
	"sort"
	"strconv"
	pb "github.com/cheggaaa/pb/v3"
)

const(
    Opening = "<"
    Closing = ">"
    Separator = ","
)

const MinimumSizeOfReference int = -1   // Use -1 to represent dynamic "smart" reference inclusion

func Compress(fileContents []byte, useProgressBar bool, maxSearchBufferLength int) ([]byte) {
	fileContents = EncodeOpeningSymbols(fileContents)

	var searchBuffer []byte
	var output []byte

	pointer := 0
	checkNextByte := false
	checkStartPointer := 0
	checkOffset := 0
	checkBytesToAdd := make([]byte, 0)
	bar := pb.New(len(fileContents))
	if useProgressBar {
		bar.Set(pb.Bytes, true)
		bar.Start()
	}
	for _, fileByte := range fileContents {
		if useProgressBar {
			bar.Increment()
		}
		index, found := 0, false
		if !checkNextByte {
			index, found = FindReverse(searchBuffer, fileByte)
		} else {
			diminishingReturns := 0
			if maxSearchBufferLength > 0 && len(searchBuffer) > maxSearchBufferLength {
				diminishingReturns = len(searchBuffer) - maxSearchBufferLength
			}
			index, found = FindReverseSlice(searchBuffer[diminishingReturns:], append(checkBytesToAdd, fileByte))
		}

		if found && checkNextByte {
			pointer = len(searchBuffer) - index
			checkStartPointer = pointer
			checkOffset++
			checkBytesToAdd = append(checkBytesToAdd, fileByte)
		} else if found && !checkNextByte {
			pointer = len(searchBuffer) - index

			checkStartPointer = pointer
			checkOffset = 1
			checkNextByte = true
			checkBytesToAdd = append(checkBytesToAdd, fileByte)
		} else {
			if checkNextByte {
				shouldAdd := true

				if MinimumSizeOfReference == -1 {
					if len(getEncoding(checkStartPointer, checkOffset)) > len(checkBytesToAdd) {
						shouldAdd = false 
					}
				}

				if len(checkBytesToAdd) > MinimumSizeOfReference && shouldAdd {
					output = append(output, getEncoding(checkStartPointer, checkOffset)...)
				} else {
					output = append(output, checkBytesToAdd...)
				}
				checkStartPointer = 0
				checkOffset = 0
				checkNextByte = false
				searchBuffer = append(searchBuffer, checkBytesToAdd...)
				// SortByteArray(searchBuffer)
				checkBytesToAdd = make([]byte, 0)
			}
			output = append(output, fileByte)
		}

		if !checkNextByte {
			searchBuffer = append(searchBuffer, fileByte)
			// SortByteArray(searchBuffer)
		}
	}
	if checkNextByte {
		shouldAdd := true

		if MinimumSizeOfReference == -1 {
			if len([]byte(Opening + strconv.Itoa(checkStartPointer) + Separator + strconv.Itoa(checkOffset) + Closing)) > len(checkBytesToAdd) {
				shouldAdd = false 
			}
		}

		if len(checkBytesToAdd) > MinimumSizeOfReference && shouldAdd {
			output = append(output, []byte(Opening + strconv.Itoa(checkStartPointer) + Separator + strconv.Itoa(checkOffset) + Closing)...)
		} else {
			output = append(output, checkBytesToAdd...)
		}
	}
	if useProgressBar {
		bar.Finish()
	}
	return output
}

func getEncoding(relativePointer int, relativeOffset int) ([]byte) {
	return []byte(Opening + strconv.Itoa(relativePointer) + Separator + strconv.Itoa(relativeOffset) + Closing)
}

func Decompress(fileContents []byte, useProgressBar bool) ([]byte) {
	searchBuffer := make([]byte, 0)
	output := make([]byte, 0)

	pointer := 0
	pointerBytes := make([]byte, 0)
	offset := 0
	offsetBytes := make([]byte, 0)
	lookingFor := Opening
	for _, fileByte := range fileContents {
		if lookingFor == Opening && string(fileByte) == Opening {
			lookingFor = Separator
		} else if lookingFor == Separator {
			if string(fileByte) == Separator {
				lookingFor = Closing
				pointer, _ = strconv.Atoi(string(pointerBytes))
				pointerBytes = make([]byte, 0)
			} else {
				pointerBytes = append(pointerBytes, fileByte)
			}
		} else if lookingFor == Closing {
			if string(fileByte) == Closing {
				lookingFor = Opening
				offset, _ = strconv.Atoi(string(offsetBytes))
				offsetBytes = make([]byte, 0)

				absolutePointer := len(searchBuffer) - pointer
				slice := searchBuffer[absolutePointer:absolutePointer + offset]
				
				output = append(output, slice...)
				searchBuffer = append(searchBuffer, slice...)
			} else {
				offsetBytes = append(offsetBytes, fileByte)
			}
		} else {
			output = append(output, fileByte)
			searchBuffer = append(searchBuffer, fileByte)
		}
	}
	output = DecodeOpeningSymbols(output)
	return output
}

const EncodedOpening = 0xff 
const EscapeByte = 0x5c

func EncodeOpeningSymbols(bytes []byte) ([]byte) {
	encoded := make([]byte, 0)
	foundEscape := false
	for _, val := range bytes {
		if string(val) == Opening {
			if foundEscape {
				encoded = append(encoded, EscapeByte)
			}
			val = EncodedOpening
		} else if val == EncodedOpening || val == EscapeByte {
			encoded = append(encoded, EscapeByte)
		} else if val == EscapeByte {  
			if foundEscape {
				encoded = append(encoded, EscapeByte)
			}
			foundEscape = true
		}
		encoded = append(encoded, val)
	}
	return encoded
}

func DecodeOpeningSymbols(bytes []byte) ([]byte) {
	decoded := make([]byte, 0)
	foundEscape := false
	for _, val := range bytes {
		if val == EncodedOpening && !foundEscape {
			newVal := []byte(Opening)
			decoded = append(decoded, newVal...)
		} else if val == EscapeByte && !foundEscape {
			foundEscape = true
		} else {
			foundEscape = false
			decoded = append(decoded, val)
		}
	}
	return decoded
}

func FindSequential(slice []byte, val byte) (int, bool) {
	// Find
    for i, item := range slice {
        if item == val {
			return i, true
		}
    }
    return -1, false
}

func FindReverseSlice(input []byte, val []byte) (int, bool) {
	if len(val) == 0 {
		return -1, false
	}

	nextToCheck := len(val) - 1

	// Find
    for i := len(input) - 1; i >= 0; i-- {
		// if MinimumSizeOfReference < 0 && len(val) > 15 && len(strconv.Itoa(len(input) - i)) > len(val) {
		// 	return -1, false
		// }

		item := input[i]
		
		if val[nextToCheck] != item {
			nextToCheck = len(val) - 1
		}

		if val[nextToCheck] == item {
			// Byte checked out
			nextToCheck--
		}

		if nextToCheck < 0 {
			return i, true
		} else if i == 0 {
			// We've reached the end of the original input but not found the value, so not found
			return -1, false
		}
    }
    return -1, false
}


func FindReverse(slice []byte, val byte) (int, bool) {
	// Find
    for i := len(slice) - 1; i >= 0; i-- {
		item := slice[i]
        if item == val {
			return i, true
		}
        i--
    }
    return -1, false
}

func SortByteArray(src []byte) {
	sort.Slice(src, func(i, j int) bool { return src[i] < src[j]})
}