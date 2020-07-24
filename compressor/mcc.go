package compressor

import (
	"fmt"
	"strings"
	"strconv"
)

type Token int
const (
	Read Token = 0
	Up Token = 1
	Select Token = 2
)

type State struct {
	isRoot bool
	token Token
	isTok bool
	symbol byte
	freq int
	transitions *[]*State
	parent *State
}

func (state *State) displayValue() string {
	if state.isTok {
		if state.token == Up {
			return "up"
		} else {
			return "read"
		}
	} else {
		return string(state.symbol)
	}
}

func (state *State) hasSymbol(symbol byte) bool {
	return !state.isTok && symbol == state.symbol
}

func (state *State) parentWithSymbol(symbol byte) int {
	if !state.isTok && !state.isRoot && symbol == state.symbol {
		return 0
	} else {
		if state.parent == nil {
			return -1
		}
		parentsHasSymbol := state.parent.parentWithSymbol(symbol)
		if parentsHasSymbol == -1 {
			return -1
		}
		return 1 + parentsHasSymbol
	}
}

func (state *State) getParent(up int) *State {
	for i := 0; i < up; i++ {
		state = state.parent
	}
	return state
}

func (state *State) bitRepresentation() int {
	for i, childState := range *state.parent.transitions {
		if childState == state {
			return i
		}
	}
	return -1
}

func (state *State) tokState(tok Token) *State {
	for _, childState := range *state.transitions {
		if childState.isTok && childState.token == tok {
			return childState
		}
	}
	return nil
}

func (state *State) getStateFromRepresentation(index int) *State {
	for i, childState := range *state.transitions {
		if i == index {
			return childState
		}
	}
	return nil
}

func createState(symbol byte, parent *State) *State {
	var states []*State
	state := State{symbol: symbol, parent: parent, freq: 1}
	for _, tok := range []Token{Read, Up, Select} {
		// TODO find a better way to count Token frequencies rather than just 100
		states = append(states, &State{token: tok, isTok: true, parent: &state, freq: 100})
	}
	state.transitions = &states
	transitions := append(*(parent.transitions), & state)
	parent.transitions = &transitions
	return &state
}

func createRoot() *State {
	var states []*State
	state := State{isRoot: true}
	for _, tok := range []Token{Read, Up, Select} {
		states = append(states, &State{token: tok, isTok: true, parent: &state, freq: 100})
	}
	state.transitions = &states
	return &state
}

func encodeBytes(fileContents []byte) ([]int, []byte) {
	bitstream := make([]int, 0)
	literals := make([]byte, 0)

	state := createRoot()
	// root := state
	for _, fileByte := range fileContents {
		containsSymbol := false
		var stateWithSymbol *State
		for _, transitionState := range *state.transitions {
			if transitionState.hasSymbol(fileByte) {
				containsSymbol = true
				stateWithSymbol = transitionState
			}
		}

		if !containsSymbol {
			parentWithSymbol := state.parentWithSymbol(fileByte)
			if parentWithSymbol == -1 {
				// Create new state with symbol
				newState := createState(fileByte, state)
				// Output Read token
				bitstream = append(bitstream, state.tokState(Read).bitRepresentation())
				literals = append(literals, fileByte)
				// Enter new state
				state = newState
			} else {
				// A parent does have the state
				// Move up to the parent with the state
				state = state.getParent(parentWithSymbol)
				// Increase frequency
				state.freq++
				// Output token to represent moving up
				for i := 0; i < parentWithSymbol + 1; i++ {
					// Reason for parentWithSymbol + 1
					// Always output an extra up token telling it to move up even if it's the direct parent
					bitstream = append(bitstream, state.tokState(Up).bitRepresentation())
				}
				// Output select token at the end to tell the decoder to use this state
				bitstream = append(bitstream, state.tokState(Select).bitRepresentation())
				// Enter state
			}
		} else {
			// The state contains the symbol
			// Enter the state with the symbol
			state = stateWithSymbol
			// Update the frequency
			state.freq++
			// Output the corresponding bit representation
			bitstream = append(bitstream, state.bitRepresentation())
		}

	}

	// printTransitions(*root, 0)
	return bitstream, literals
}

func decodeBytes(bitstream []int, literals []byte) []byte {
	state := createRoot()
	// root := state

	output := make([]byte, 0)

	movingUp := false

	for _, bit := range bitstream {
		childState := state.getStateFromRepresentation(bit)
		if childState == nil {
			panic("Couldn't find child state with bit representation")
		}

		if childState.isTok {
			if childState.token == Read {
				// Read token
				// Pop char from beginning of literal stream
				symbol := literals[0]
				literals = literals[1:]
				// Output symbol too outstream
				output = append(output, symbol)
				// Create new state with symbol in parent state
				newState := createState(symbol, childState.parent)
				// Enter new state
				state = newState
			} else if childState.token == Up {
				// Up token
				// If we're not already in the process of moving up
				if !movingUp {
					// Remember we're moving up and don't enter the parent state
					// because we're already in the "first" parent state
					movingUp = true
				} else {
					// Enter the parent state
					state = state.parent
					if state.parent == nil {
						panic("Trying to go up past root node")
					}
				}
			} else if childState.token == Select {
				// Output token to outstream
				output = append(output, state.symbol)
				// Reset moving up status
				movingUp = false
			} else {
				panic("Unknown token passed")
			}
		} else {
			// State represents a literal
			// Enter new state
			state = childState
			// Output the literal
			output = append(output, state.symbol)
		}
	}
	return output
}

const seperator = byte('\\')

func encodeStreamAndLiterals(bitstream []int, literals []byte) []byte {
	bits := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(bitstream)), ","), "[]")
	return append(append([]byte(bits), seperator), literals...)
}

func decodeStreamAndLiterals(bytes []byte) ([]int, []byte) {
	stringInput := string(bytes)
	seperatorIndex := strings.IndexByte(stringInput, seperator)
	bitstrings := strings.Split(stringInput[:seperatorIndex], ",")
	literals := bytes[seperatorIndex + 1:]
	bits := make([]int, len(bitstrings))
	for i, bitstring := range bitstrings {
		num, err := strconv.Atoi(bitstring)
		if err != nil {
			panic(err)
		}
		bits[i] = num
	}
	return bits, literals
}

func MCCCompress(fileContents []byte) []byte {
	bitstream, literals := encodeBytes(fileContents)
	fmt.Println("Rough estimate of bytes:", (len(bitstream)/8) + len(literals))
	return encodeStreamAndLiterals(bitstream, literals)
}

func MCCDecompress(fileContents []byte) []byte {
	bitstream, literals := decodeStreamAndLiterals(fileContents)
	output := decodeBytes(bitstream, literals)
	return output
}

func printTransitions(parent State, indentation int) {
	for _, state := range *parent.transitions {
		fmt.Print(strings.Repeat("-", indentation), state.displayValue(), "\n")

		if state.transitions != nil {
			printTransitions(*state, indentation + 1)
		}
	}
}