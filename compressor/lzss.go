package compressor

import (
	"fmt"
	"strconv"
)

const(
    Opening = "<"
    Closing = ">"
    Separator = ","
)

const MinimumSizeOfReference int = -1   // Use -1 to represent dynamic "smart" reference inclusion

func Compress(fileContents []byte) ([]byte) {
	fileContents = EncodeOpeningSymbols(fileContents)

	searchBuffer := make([]byte, 0)
	output := make([]byte, 0)

	pointer := 0
	checkNextByte := false
	checkStartPointer := 0
	checkOffset := 0
	checkBytesToAdd := make([]byte, 0)
	for _, fileByte := range fileContents {
		index, found := 0, false
		if !checkNextByte {
			copySearchBuffer := make([]byte, len(searchBuffer))
			copy(copySearchBuffer, searchBuffer)
			index, found = FindReverse(copySearchBuffer, fileByte)
		} else {
			// We don't need to copy the searchBuffer because FindSequential does not modify it
			index, found = FindReverseSlice(searchBuffer, append(checkBytesToAdd, fileByte))
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
					if len([]byte(Opening + strconv.Itoa(checkStartPointer) + Separator + strconv.Itoa(checkOffset) + Closing)) > len(checkBytesToAdd) {
						shouldAdd = false 
					}
				}

				if len(checkBytesToAdd) > MinimumSizeOfReference && shouldAdd {
					output = append(output, []byte(Opening + strconv.Itoa(checkStartPointer) + Separator + strconv.Itoa(checkOffset) + Closing)...)
				} else {
					output = append(output, checkBytesToAdd...)
				}
				checkStartPointer = 0
				checkOffset = 0
				checkNextByte = false
				searchBuffer = append(searchBuffer, checkBytesToAdd...)
				checkBytesToAdd = make([]byte, 0)
			}
			output = append(output, fileByte)
		}

		if !checkNextByte {
			searchBuffer = append(searchBuffer, fileByte)
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
	return output
}

func Decompress(fileContents []byte) ([]byte) {
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

func EncodeOpeningSymbols(bytes []byte) ([]byte) {
	encoded := make([]byte, len(bytes))
	for i, val := range bytes {
		if string(val) == Opening {
			val = EncodedOpening
		}
		encoded[i] = val
	}
	return encoded
}

func DecodeOpeningSymbols(bytes []byte) ([]byte) {
	decoded := make([]byte, 0)
	for _, val := range bytes {
		if val == EncodedOpening {
			newVal := []byte(Opening)
			decoded = append(decoded, newVal...)
		} else {
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
	foundByte := false
	scanIndex := -1

	// Find
    for i := len(input) - 1; i >= 0; i-- {
		item := input[i]
		scanIndex = i
		for valIndex := len(val) - 1; valIndex >= 0; valIndex-- {
			valItem := val[valIndex]
			if valItem == item {
				if !foundByte {
					// lastIndex = i - valIndex
					foundByte = true
				}
				if valIndex == 0 || scanIndex == 0 {
					if valIndex != 0 {
						foundByte = false // We haven't found all of the values we're scanning for
					}
					break
				} else {
					scanIndex--
					if scanIndex > len(input) || scanIndex < 0 {
						fmt.Printf("Trying to grab %v from %v\n", scanIndex, string(input))
					}
					item = input[scanIndex]
				}
			} else {
				foundByte = false
				break
			}
		}
		if foundByte {
			return scanIndex, foundByte
		}
        
    }
    return scanIndex, foundByte
}

func FindReverse(slice []byte, val byte) (int, bool) {
	// Reverse
	for i := len(slice)/2-1; i >= 0; i-- {
		opp := len(slice)-1-i
		slice[i], slice[opp] = slice[opp], slice[i]
	}
	// Find
	i := len(slice) - 1
    for _, item := range slice {
        if item == val {
			return i, true
		}
        i--
    }
    return -1, false
}
