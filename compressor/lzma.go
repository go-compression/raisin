package compressor

import (
	"fmt"
	"strings"
)

type MarkovChain struct {
	Value byte
	Nodes *[]MarkovChain
	Occurences int
	MoveUp int
}

func buildMarkovChain(value byte) MarkovChain {
	return MarkovChain{Value: value, Nodes: &[]MarkovChain{}, Occurences: 1, MoveUp: 0}
}

func buildMarkovChainMoveUp(moveUp int) MarkovChain {
	return MarkovChain{Nodes: &[]MarkovChain{}, Occurences: 1, MoveUp: moveUp}
}

func LZMACompress(fileContents []byte, _ bool, _ int) ([]byte) {
	var nodes []MarkovChain
	chain := MarkovChain{Nodes: &nodes}

	stack := []MarkovChain{chain}
	for _, fileByte := range fileContents {
		valueUpStack := FindValueUpStack(fileByte, stack)
		if valueUpStack != -1 {
			moveUpIndex := GetIndexOfNodeWithMoveUp(len(stack) - valueUpStack, *stack[len(stack) - 1].Nodes)
			if moveUpIndex == -1 {
				*stack[len(stack) - 1].Nodes = append(*stack[len(stack) - 1].Nodes, buildMarkovChainMoveUp(len(stack) - valueUpStack))
			} else {
				(*stack[len(stack) - 1].Nodes)[moveUpIndex].Occurences++
			}
			stack = stack[:valueUpStack]
		}
		node := stack[len(stack) - 1] // Get latest node off the list
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
	PrintMarkovChain(&chain, 0)
	// fmt.Println(GetBitsFromChain(&chain, fileContents))
	return []byte("Hello!")
}

func GetBitsFromChain(node *MarkovChain, input []byte) ([]int) {
	if len(input) > 0 {
		val := input[0]
		bits := 0
		nodes := *node.Nodes
		if len(nodes) == 1 {
			bits = 0
		} else if len(nodes) == 2 {
			bits = 1
		} else {
			bits = 8
		}
		index := GetIndexOfNodeWithValue(val, *node.Nodes)
		var lookInNode *MarkovChain
		if index == -1 {
			expandChainMoveUps(node, stack)
		} else {
			lookInNode = &nodes[index]
		}
		return append([]int{bits}, GetBitsFromChain(lookInNode, input[1:])...)
	}
	return []int{-1}
}

func expandChainMoveUps(node *MarkovChain, stack *[]MarkovChain) {
	for _, node := range *node.Nodes {
		if node.MoveUp != 0 {
			*node.Nodes = (*stack)[len(*stack) - node.MoveUp:]
		}
	}
}

// FindValueUpStack returns how far up the stack it should go to find the byte, len(stack) for not found
func FindValueUpStack(lookFor byte, stack []MarkovChain) (int) {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].Value == lookFor {
			return i
		}
	}
	return -1
}

func GetIndexOfNodeWithValue(lookFor byte, nodes []MarkovChain) (int) {
	for i, node := range nodes {
		if node.Value == lookFor {
			return i
		}
	}
	return -1
}

func GetIndexOfNodeWithMoveUp(moveUp int, nodes []MarkovChain) (int) {
	for i, node := range nodes {
		if node.MoveUp == moveUp {
			return i
		}
	}
	return -1
}

func PrintMarkovChain(chain *MarkovChain, indentation int) {
	for _, node := range *chain.Nodes { 
		fmt.Print(strings.Repeat("-", indentation), string([]byte{node.Value}), " Move up:", node.MoveUp, " O: ", node.Occurences, "\n")
		
		if node.Nodes != nil {
			PrintMarkovChain(&node, indentation + 1)
		}
	}
}

func LZMADecompress(fileContents []byte, _ bool) ([]byte) {
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
