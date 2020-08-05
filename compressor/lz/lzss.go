package lz

import (
	"sort"
	"strconv"
	"sync"
	"bytes"
	pb "github.com/cheggaaa/pb/v3"
	"io"
	"fmt"
	"io/ioutil"
)

const(
    Opening = "<"
    Closing = ">"
    Separator = ","
)

const MinimumSizeOfReference int = -1   // Use -1 to represent dynamic "smart" reference inclusion

type Reference struct {
	value []byte
	isReference bool
	negativeOffset int
	size int
}

type Writer struct {
	windowSize int
	useProgressBar bool
	w io.Writer
}

const DefaultWindowSize = 4096

func NewWriter(w io.Writer) *Writer {
	z, _ := NewWriterLevel(w, DefaultWindowSize)
	return z
}

func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	if level < 0 {
		return nil, fmt.Errorf("lzss: invalid compression level: %d", level)
	}
	z := new(Writer)
	z.windowSize = level
	z.useProgressBar = true
	z.w = w
	return z, nil
}

func (writer *Writer) Write(data []byte) (n int, err error) {
	compressed := Compress(data, writer.useProgressBar, writer.windowSize)
	writer.w.Write(compressed)
	return len(compressed), nil
}

func (writer *Writer) Close() error {
	return nil
}

type Reader struct {
	r            io.Reader
	compressed []byte
	decompressed []byte
	pos          int
}

func NewReader(r io.Reader) (*Reader, error) {
	z := new(Reader)
	z.r = r
	var err error
	z.compressed, err = ioutil.ReadAll(r)
	return z, err
}

func (r *Reader) Read(content []byte) (n int, err error) {
	if r.decompressed == nil {
		r.decompressed = Decompress(r.compressed, true)
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

func Compress(fileContents []byte, useProgressBar bool, maxSearchBufferLength int) ([]byte) {
	fileContents = EncodeOpeningSymbols(fileContents)
	var waitgroup sync.WaitGroup

	bar := pb.New(len(fileContents))
	bar.Set(pb.Bytes, true)
	bar.Start()

	output := make([](chan Reference), len(fileContents))

	for i := 0; i < len(fileContents); i++ {
		waitgroup.Add(1)
		output[i] = make(chan Reference, 1)

		startIndex := 0
		searchBuffer := fileContents[:i]
		if maxSearchBufferLength > 0 && len(searchBuffer) > maxSearchBufferLength {
			startIndex = len(searchBuffer) - maxSearchBufferLength
		}

		go CompressorWorkerAsync(&waitgroup, output[i], searchBuffer[startIndex:], []byte{fileContents[i]}, fileContents[i:], bar)
	}

	waitgroup.Wait()

	var finalOutput []byte
	var ignoreNextChars int
	for _, i := range output {
		close(i)
		ref := <-i
		if ignoreNextChars > 0 {
			ignoreNextChars--
		} else if ref.isReference {
			ignoreNextChars = ref.size - 1
			if len(getEncoding(ref.negativeOffset, ref.size)) < ref.size {
				finalOutput = append(finalOutput, getEncoding(ref.negativeOffset, ref.size)...)
			} else {
				finalOutput = append(finalOutput, ref.value...)
			}
		} else {
			finalOutput = append(finalOutput, ref.value...)
		}
	}

	return finalOutput
}

func CompressorWorkerAsync(waitgroup *sync.WaitGroup, output chan<- Reference, searchBuffer []byte, scanBytes []byte, nextBytes []byte, bar *pb.ProgressBar) {
	defer waitgroup.Done()

	out := CompressorWorker(searchBuffer, scanBytes, nextBytes[1:], bar)

	output <- out

	bar.Increment()
}

func CompressorWorker(searchBuffer []byte, scanBytes []byte, nextBytes []byte, bar *pb.ProgressBar) (Reference) {
	index, found := FindReverseSlice(searchBuffer, scanBytes)

	if found {
		negativeOffset := len(searchBuffer) - index
		if len(nextBytes) > 0 {
			checkNextByte := CompressorWorker(searchBuffer, append(scanBytes, nextBytes[0]), nextBytes[1:], bar)
			if checkNextByte.isReference {
				return checkNextByte
			} else {
				return Reference{value: scanBytes, isReference: true, negativeOffset: negativeOffset, size: len(scanBytes)}
			}
		} else {
			return Reference{value: scanBytes, isReference: true, negativeOffset: negativeOffset, size: len(scanBytes)}
		}
	} else {
		return Reference{value: scanBytes}
	}
}

func CompressFileSync2(fileContents []byte, _ bool, maxSearchBufferLength int) ([]byte) {
	fileContents = EncodeOpeningSymbols(fileContents)
	var output []byte

	bar := pb.New(len(fileContents))
	bar.Set(pb.Bytes, true)
	bar.Start()

	var ignoreNextChars int
	for i, fileByte := range fileContents {
		windowStart := 0
		if i > maxSearchBufferLength {
			windowStart = len(fileContents[:i]) - maxSearchBufferLength
		}
		ref := CompressorWorker(fileContents[windowStart:i], []byte{fileByte}, fileContents[i:], bar)

		if ignoreNextChars > 0 {
			ignoreNextChars--
		} else if (ref.isReference) {
			ignoreNextChars = ref.size - 1
			if len(getEncoding(ref.negativeOffset, ref.size)) < ref.size {
				output = append(output, getEncoding(ref.negativeOffset, ref.size)...)
			} else {
				output = append(output, ref.value...)
			}
		} else {
			output = append(output, ref.value...)
		}
		bar.Increment()
	}
	return output
}

func CompressFileSync(fileContents []byte, useProgressBar bool, maxSearchBufferLength int) ([]byte) {
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
	index := bytes.Index(input, val)
	return index, index != -1
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