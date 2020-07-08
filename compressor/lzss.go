package compressor

import (
	"strconv"
)

const(
    Opening = "<"
    Closing = ">"
    Separator = ","
)

func Compress(fileContents []byte) ([]byte) {
	searchBuffer := make([]byte, 0)
	output := make([]string, 0)

	pointer := 0
	checkNextByte := false
	checkStartPointer := 0
	// lastPointer := 0
	lastIndex := 0
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
			index, found = FindSequential(searchBuffer[lastIndex + 1:], fileByte)
		}

		if found && checkNextByte && index == 0 {
			pointer = len(searchBuffer) - index
			checkOffset++
			// lastPointer = pointer
			lastIndex++
			checkBytesToAdd = append(checkBytesToAdd, fileByte)
		} else if found && !checkNextByte {
			pointer = len(searchBuffer) - index

			checkStartPointer = pointer
			// lastPointer = pointer
			lastIndex = index
			checkOffset = 1
			checkNextByte = true
			checkBytesToAdd = append(checkBytesToAdd, fileByte)
		} else {
			if checkNextByte {
				output = append(output, Opening + strconv.Itoa(checkStartPointer) + Separator + strconv.Itoa(checkOffset) + Closing)
				checkStartPointer = 0
				checkOffset = 0
				checkNextByte = false
				searchBuffer = append(searchBuffer, checkBytesToAdd...)
				checkBytesToAdd = make([]byte, 0)
			}
			output = append(output, string(fileByte))
		}

		if !checkNextByte {
			searchBuffer = append(searchBuffer, fileByte)
		}
	}
	if checkNextByte {
		output = append(output, Opening + strconv.Itoa(checkStartPointer) + Separator + strconv.Itoa(checkOffset) + Closing)
	}
	return StringsToBytes(output)
}

func Decompress(fileContents []byte) ([]byte) {
	searchBuffer := make([]byte, 0)
	output := make([]string, 0)

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
				
				output = append(output, string(slice))
				searchBuffer = append(searchBuffer, slice...)
			} else {
				offsetBytes = append(offsetBytes, fileByte)
			}
		} else {
			output = append(output, string(fileByte))
			searchBuffer = append(searchBuffer, fileByte)
		}
	}
	return StringsToBytes(output)
}

func StringsToBytes(input []string) ([]byte) {
	output := make([]byte, 0)
	for _, v := range input {
		output = append(output, []byte(v)...)
	}
	return output
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
