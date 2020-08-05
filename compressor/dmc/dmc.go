package dmc

import (
	"fmt"
	"strings"
	// "errors"
	// "bitbucket.org/sheran_gunasekera/leb128"
	"encoding/binary"
	"bytes"
	"sort"
	// "unsafe"
	"runtime"
	"io"
	"io/ioutil"
)

type MarkovChain struct {
	Value      byte
	Nodes      *[]MarkovChain
	Occurences int
	MoveUp     int
}

func buildMarkovChain(value byte) MarkovChain {
	return MarkovChain{Value: value, Nodes: &[]MarkovChain{}, Occurences: 1, MoveUp: 0}
}

func buildMarkovChainMoveUp(moveUp int) MarkovChain {
	return MarkovChain{Nodes: &[]MarkovChain{}, Occurences: 1, MoveUp: moveUp}
}

func Compress(fileContents []byte) []byte {
	var m1, m2 runtime.MemStats
	var nodes []MarkovChain
	runtime.ReadMemStats(&m1)
	chain := MarkovChain{Nodes: &nodes}

	stack := []MarkovChain{chain}
	for _, fileByte := range fileContents {
		// print(string(fileByte))
		valueUpStack := FindValueUpStack(fileByte, stack)
		if valueUpStack != -1 {
			moveUpIndex := GetIndexOfNodeWithMoveUp(len(stack)-valueUpStack, *stack[len(stack)-1].Nodes)
			if moveUpIndex == -1 {
				*stack[len(stack)-1].Nodes = append(*stack[len(stack)-1].Nodes, buildMarkovChainMoveUp(len(stack)-valueUpStack))
			} else {
				(*stack[len(stack)-1].Nodes)[moveUpIndex].Occurences++
			}
			stack = stack[:valueUpStack]
		}
		node := stack[len(stack)-1] // Get latest node off the list
		index := GetIndexOfNodeWithValue(fileByte, *node.Nodes)
		if index == -1 {
			// Not found in node, build new one
			newNode := buildMarkovChain(fileByte)
			*node.Nodes = append(*node.Nodes, newNode)

			stack = append(stack, newNode)
		} else {
			// Found in stack, add to occurences
			(*node.Nodes)[index].Occurences++
			if (*node.Nodes)[index].Occurences > 1 {
				// Clone a new state
				stack = append(stack, (*node.Nodes)[index])
			}
		}
	}
	runtime.ReadMemStats(&m2)
	memUsage(&m1, &m2)

	SortNodesByOccurences(&chain)
	// PrintMarkovChain(&chain, 0)
	// fmt.Println("Total upward travels:", UpwardTravels(&chain))
	// fmt.Println("Compiling into bits")

	bits := GetBitsFromChain(&chain, fileContents, &[]MarkovChain{})
	// fmt.Println("Length of bits:", len(bits))

	encodeBits := new(bytes.Buffer)
	for _, num := range bits {
		err := binary.Write(encodeBits, binary.LittleEndian, int8(num))
		check(err)
	}

	// fmt.Println(bits)
	// decoded := GetOutputFromBits(bits, &chain, &[]MarkovChain{})
	// fmt.Println("Decoded:", string(decoded))
	// fmt.Println("Lossless markov:", string(decoded) == string(fileContents))
	// fmt.Println("Bytes of chain:", unsafe.Sizeof(chain))

	return encodeBits.Bytes()
}

func memUsage(m1, m2 *runtime.MemStats) {
	fmt.Println("Alloc:", m2.Alloc-m1.Alloc,
		"TotalAlloc:", m2.TotalAlloc-m1.TotalAlloc,
		"HeapAlloc:", m2.HeapAlloc-m1.HeapAlloc)
}

func GetBitsFromChain(node *MarkovChain, input []byte, stack *[]MarkovChain) []int {
	if len(input) > 0 {
		newStack := append(*stack, *node)
		val := input[0]
		nodes := *node.Nodes

		var transition int

		index := GetIndexOfNodeWithValue(val, *node.Nodes)
		var lookInNode *MarkovChain
		if index == -1 {
			childNodeMoveUps := make([]int, len(*node.Nodes))
			for i, childNode := range *node.Nodes {
				childNodeMoveUps[i] = childNode.MoveUp
				if childNode.MoveUp > 0 {
					if newStack[len(newStack)-childNode.MoveUp].Value == val {
						lookInNode = &newStack[len(newStack)-childNode.MoveUp]
						newStack = newStack[:len(newStack)-childNode.MoveUp+1]
						transition = i
						if len(*node.Nodes) == 1 {
							transition = -1 // There's only 1 possible transition so we don't need to encode it
						}
						break
					}
				}
			}
			newStack = newStack[:len(newStack)-1]
		} else {
			lookInNode = &nodes[index]
			transition = index
			if len(nodes) == 1 {
				transition = -1 // There's only 1 possible transition so we don't need to encode it
			}
		}

		if transition == -1 {
			bitsFromChain := GetBitsFromChain(lookInNode, input[1:], &newStack)
			if len(bitsFromChain) > 1 && bitsFromChain[len(bitsFromChain) - 2] == -2 {
				bitsFromChain[len(bitsFromChain) - 1]++
			}
			return bitsFromChain
		}
		bitsFromChain := GetBitsFromChain(lookInNode, input[1:], &newStack)
		if len(bitsFromChain) > 1 && bitsFromChain[len(bitsFromChain) - 2] == -2 {
			bitsFromChain[len(bitsFromChain) - 2] = -1
		}
		return append([]int{transition}, bitsFromChain...)
	}
	return []int{-2, 0} // -2 represents end of input traversal, 0 is incremented as the end of input traverses down nodes not represented
}

func GetOutputFromBits(bits []int, node *MarkovChain, previousStack *[]MarkovChain) ([]byte) {
	stack := append(*previousStack, *node)
	if len(*node.Nodes) == 1 && bits[0] >= 0 {
		node = &(*node.Nodes)[0]
		if node.MoveUp != 0 {
			moveUp := node.MoveUp
			node = &stack[len(stack) - moveUp]
			stack = stack[:len(stack) - moveUp]
		}
		nodeVal := node.Value
		// fmt.Println(string(nodeVal))
		return append([]byte{nodeVal}, GetOutputFromBits(bits, node, &stack)...)
	} else {
		path := bits[0]
		if path == -1 || path == -2 { // End of stream integers
			// TODO chop off ending integer if bits[1] == 0 during encoding (and handle it here during decoding)
			transitions := bits[1] // Number of transitions to take last is represented as an int at the end
			endingBytes := make([]byte, transitions)
			for i := 0; i < transitions; i++ {
				node = &(*node.Nodes)[0]
				if node.MoveUp != 0 {
					moveUp := node.MoveUp
					node = &stack[len(stack) - moveUp]
					stack = stack[:len(stack) - moveUp + 1]
				} else {
					stack = append(stack, *node)
				}
				endingBytes[i] = node.Value
			}
			// fmt.Println(string(endingBytes))
			return endingBytes
		} else {
			node = &(*node.Nodes)[path]
			if node.MoveUp != 0 {
				moveUp := node.MoveUp
				node = &stack[len(stack) - moveUp]
				stack = stack[:len(stack) - moveUp]
			}
			nodeVal := node.Value
			// fmt.Println(string(nodeVal))
			return append([]byte{nodeVal}, GetOutputFromBits(bits[1:], node, &stack)...)
		}
	}
}

func SortNodesByOccurences(chain *MarkovChain) {
	sort.Slice(*chain.Nodes, func(i, j int) bool {
		return (*chain.Nodes)[i].Occurences > (*chain.Nodes)[j].Occurences
	})
	for _, node := range *chain.Nodes {
		if len(*node.Nodes) > 0 {
			SortNodesByOccurences(&node)
		}
	}
}

func expandChainMoveUps(originNode *MarkovChain, stack *[]MarkovChain) {
	for i, node := range *originNode.Nodes {
		if node.MoveUp != 0 {
			// *(*originNode.Nodes)[i].Nodes = (*stack)[len(*stack) - node.MoveUp:]
			(*originNode.Nodes)[i] = (*stack)[len(*stack)-node.MoveUp]
			// *originNode.Nodes = (*stack)[len(*stack) - node.MoveUp:]
		}
	}
}

// FindValueUpStack returns how far up the stack it should go to find the byte, len(stack) for not found
func FindValueUpStack(lookFor byte, stack []MarkovChain) int {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].Value == lookFor {
			return i
		}
	}
	return -1
}

func GetIndexOfNodeWithValue(lookFor byte, nodes []MarkovChain) int {
	for i, node := range nodes {
		if node.Value == lookFor {
			return i
		}
	}
	return -1
}

func GetIndexOfNodeWithMoveUp(moveUp int, nodes []MarkovChain) int {
	for i, node := range nodes {
		if node.MoveUp == moveUp {
			return i
		}
	}
	return -1
}

func UpwardTravels(chain *MarkovChain) int {
	var total int
	for _, node := range *chain.Nodes {
		if node.Nodes != nil {
			total += UpwardTravels(&node)
		}
		total += node.MoveUp
	}
	return total
}

func PrintMarkovChain(chain *MarkovChain, indentation int) {
	for _, node := range *chain.Nodes {
		char := string([]byte{node.Value})
		if char == " " {
			char = "spc"
		}
		if node.MoveUp != 0 {
			fmt.Print(strings.Repeat("--", indentation), "Up ", node.MoveUp, " O: ", node.Occurences, "\n")
		} else {
			fmt.Print(strings.Repeat("--", indentation), char, " O: ", node.Occurences, "\n")
		}

		if node.Nodes != nil {
			PrintMarkovChain(&node, indentation+1)
		}
	}
}

func Decompress(fileContents []byte) []byte {
	return []byte("Hello!")
}

func bitsFromBytes(bs []byte) []int {
	r := make([]int, len(bs)*8)
	for i, b := range bs {
		for j := 0; j < 8; j++ {
			r[i*8+j] = int(b >> uint(7-j) & 0x01)
		}
	}
	return r
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}


type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
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
	compressed []byte
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