package compressor

import (
	"fmt"
	"strings"
	// "errors"
	// "bitbucket.org/sheran_gunasekera/leb128"
	// "encoding/binary"
	"sort"
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

func DMCCompress(fileContents []byte) []byte {
	var nodes []MarkovChain
	chain := MarkovChain{Nodes: &nodes}

	stack := []MarkovChain{chain}
	for _, fileByte := range fileContents {
		print(string(fileByte))
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
	// fmt.Println(leb128.EncodeULeb128(uint32(256)))

	SortNodesByOccurences(&chain)
	PrintMarkovChain(&chain, 0)

	bits := GetBitsFromChain(&chain, fileContents, &[]MarkovChain{})
	fmt.Println(bits)
	decoded := GetOutputFromBits(bits, &chain, &[]MarkovChain{})
	fmt.Println("Decoded:", string(decoded))
	fmt.Println("Lossless markov:", string(decoded) == string(fileContents))
	return []byte(strings.Trim(strings.Join(strings.Fields(fmt.Sprint(bits)), ","), "[]"))
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

func DMCDecompress(fileContents []byte) []byte {
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
