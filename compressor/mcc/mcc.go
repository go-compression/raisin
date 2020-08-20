package mcc

import (
	"fmt"
	huff "github.com/icza/huffman"
	// huffman "github.com/go-compression/raisin/compressor/huffman"
	"io"
	"io/ioutil"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Token int

const (
	Read Token = 0
	// Up token has been replaced by dynamic order-of-two-based up tokens
	// Up Token = 1 // TODO Research linear distances between markov chains
)

type State struct {
	isRoot      bool
	token       Token
	isTok       bool
	symbol      byte
	freq        int
	transitions *[]*State // TODO Try to remove references here as interfaces should modify object with reference to it
	parent      *State
}

func (state *State) displayValue() string {
	if state.isTok {
		if state.token == Token(1) {
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

func (state *State) tokState(tok Token) *State {
	for _, childState := range *state.transitions {
		if childState.isTok && childState.token == tok {
			return childState
		}
	}
	return nil
}

func (state *State) sortByFrequency() {
	sort.Slice((*state.transitions), func(i, j int) bool {
		return (*state.transitions)[i].freq > (*state.transitions)[j].freq
	})
}

func (state *State) combinedValue() int {
	if state.isTok {
		return int(state.token)
	} else {
		return int(state.symbol) + int(math.Pow(2, float64(highest_order_for_up)))
	}
}

func (state *State) bitRepresentation() (int, string) {
	// symFreqs := make(map[rune]int)
	// for _, c := range *state.parent.transitions {
	// 	symFreqs[rune(c.combinedValue())] = c.freq + 1000
	// }
	// testTree := huffman.BuildTree(symFreqs)
	// fmt.Println(testTree)
	// fmt.Println(string(huffman.Encode(testTree, string(rune(state.combinedValue())))))
	leaves := make([]*huff.Node, len(*state.parent.transitions))
	for i, childState := range *state.parent.transitions {
		leaves[i] = &huff.Node{Value: huff.ValueType(childState.combinedValue()), Count: childState.freq + 1000}
	}
	root := huff.Build(leaves)
	// huffman.Print(root)
	// fmt.Println(root.Right.Right.Code())
	bits := getValueOf(root, huff.ValueType(state.combinedValue()))
	// fmt.Println(state.combinedValue(), "-", bits)
	for i, childState := range *state.parent.transitions {
		if childState == state {
			return i, bits
		}
	}
	return -1, bits
}

func getValueOf(root *huff.Node, value huff.ValueType) string {
	var traverse func(n *huff.Node, code uint64, bits byte, lookFor huff.ValueType) (bool, string)

	traverse = func(n *huff.Node, code uint64, bits byte, lookFor huff.ValueType) (bool, string) {
		if n.Left == nil {
			// Leaf
			if n.Value == lookFor {
				// fmt.Printf("'%c': %0"+strconv.Itoa(int(bits))+"b\n", n.Value, code)
				return true, fmt.Sprintf("%0"+strconv.Itoa(int(bits))+"b", code)
			}
			return false, ""
		}
		bits++
		var found bool
		var result string
		found, result = traverse(n.Left, code<<1, bits, lookFor)
		if found {
			return found, result
		}
		found, result = traverse(n.Right, code<<1+1, bits, lookFor)
		if found {
			return found, result
		}
		return false, ""
	}

	found, result := traverse(root, 0, 0, value)
	if !found {
		panic("Not found result")
	}
	return result
}

func (state *State) getStateFromRepresentation(index int) *State {
	for i, childState := range *state.transitions {
		if i == index {
			return childState
		}
	}
	return nil
}

const highest_order_for_up = 8 // 2^8 = 256

func getTokens() []Token {
	tokens := make([]Token, 0)
	for i := 0; i <= 8; i++ {
		magnitude := math.Pow(2, float64(i))
		tokens = append(tokens, Token(magnitude))
	}
	return tokens
}

func generateStateTokens(state *State) []*State {
	var states []*State
	tokens := append([]Token{Read}, getTokens()...)
	freq := 1000
	for i, tok := range tokens {
		// TODO find a better way to count Token frequencies rather than just literals
		if i == 2 {
			freq = 0
		}
		freq -= 100
		states = append(states, &State{token: tok, isTok: true, parent: state, freq: freq})
	}
	return states
}

func createState(symbol byte, parent *State) *State {
	// Create new state
	state := State{symbol: symbol, parent: parent, freq: 1}

	// Create tokens
	states := generateStateTokens(&state)
	state.transitions = &states

	// Add state to parent
	transitions := append(*(parent.transitions), &state)
	parent.transitions = &transitions

	return &state
}

func createRoot() *State {
	// Generate root
	state := State{isRoot: true}

	// Create tokens
	states := generateStateTokens(&state)
	state.transitions = &states
	return &state
}

func encodeBytes(fileContents []byte) ([]int, []byte, int) {
	bitsize := 0
	allbits := make([]string, 0)
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
				tokenState, bits := state.tokState(Read).bitRepresentation()
				allbits = append(allbits, bits)
				bitsize += len(bits)
				bitstream = append(bitstream, tokenState)
				literals = append(literals, fileByte)
				// Enter new state
				state = newState
			} else {
				// A parent does have the state
				// Save original state for moving up
				origState := state
				// Move up to the parent with the state
				state = state.getParent(parentWithSymbol)
				// Increase frequency
				state.freq++
				// Re-sort transitions by frequency
				state.parent.sortByFrequency()

				parentWithSymbol++

				// We use this to track whether we've encoded a move up yet
				encoded := false

				// Try to encode tokens with larger magnitude up tokens
				for i := highest_order_for_up; i >= 0; i-- {
					// Calculate magnitude
					magnitude := int(math.Pow(2, float64(i)))
					// If the amount of times we're moving up is divisible by the magnitude
					if parentWithSymbol-magnitude >= 0 {
						// Calculate how many times it is divisible by
						divisibleTimes := parentWithSymbol / magnitude
						// For each time it is divisible by the magnitude
						for j := 0; j < divisibleTimes; j++ {
							// Output magnitude token
							tokenState, bits := origState.tokState(Token(magnitude)).bitRepresentation()
							allbits = append(allbits, bits)
							bitsize += len(bits)
							bitstream = append(bitstream, tokenState)
							// Remove magnitude from the amount of times to go up
							parentWithSymbol -= magnitude
							// Move up the amount of times represented by the magnitude
							if encoded {
								// We've already encoded a move up so just encode the magnitude
								origState = origState.getParent(magnitude)
							} else {
								// We use -1 because the first "up" tells it to go into the current state, so ignore it
								origState = origState.getParent(magnitude - 1)
								// Remember we've already encoded a magnitude - 1 so don't re-encode it
								encoded = true
							}

						}
					}
				}

				// Output read token at the end to tell the decoder to use this state's symbol
				tokenState, bits := state.tokState(Read).bitRepresentation()
				allbits = append(allbits, bits)
				bitsize += len(bits)
				bitstream = append(bitstream, tokenState)
			}
		} else {
			// The state contains the symbol
			// Enter the state with the symbol
			state = stateWithSymbol
			// Output the corresponding bit representation
			tokenState, bits := state.bitRepresentation()
			allbits = append(allbits, bits)
			bitsize += len(bits)
			bitstream = append(bitstream, tokenState)
			// Update the frequency
			state.freq++
			// Re-sort transitions by frequency
			state.parent.sortByFrequency()
		}

	}

	fmt.Println("True encoded bitlength:", bitsize, "bytes:", bitsize/8)
	// fmt.Println("Bits: " + strings.Join(allbits, ""))

	// printTransitions(*root, 0)
	return bitstream, literals, bitsize
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
				if movingUp {
					// Output token to outstream
					output = append(output, state.symbol)
					// Reset moving up status
					movingUp = false
					// Increase frequency of state
					state.freq++
					// Re-sort transitions by frequency
					state.parent.sortByFrequency()
				} else {
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
				}
			} else {
				moveUpTimes := int(childState.token)
				if !movingUp {
					movingUp = true
					moveUpTimes--
				}
				for i := 0; i < moveUpTimes; i++ {
					if state.parent == nil {
						panic("Trying to go up past root node")
					}
					// Enter the parent state
					state = state.parent
				}
			}
		} else {
			// State represents a literal
			// Enter new state
			state = childState
			// Output the literal
			output = append(output, state.symbol)
			// Update the frequency
			state.freq++
			// Re-sort transitions by frequency
			state.parent.sortByFrequency()
		}
	}
	return output
}

const separator = byte('\\')

func encodeStreamAndLiterals(bitstream []int, literals []byte) []byte {
	bits := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(bitstream)), ","), "[]")
	return append(append([]byte(bits), separator), literals...)
}

func decodeStreamAndLiterals(bytes []byte) ([]int, []byte) {
	stringInput := string(bytes)
	separatorIndex := strings.IndexByte(stringInput, separator)
	bitstrings := strings.Split(stringInput[:separatorIndex], ",")
	literals := bytes[separatorIndex+1:]
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

func Compress(fileContents []byte) []byte {
	bitstream, literals, bitsize := encodeBytes(fileContents)
	fmt.Println("Character bytes:", len(literals))
	fmt.Println("State bits:", bitsize, "bytes:", bitsize/8)
	fmt.Println("True estimate of bytes:", (bitsize/8)+len(literals))
	fmt.Println("-----------------------")
	fmt.Println("Compression Ratio:", float32((bitsize/8)+len(literals))/float32(len(fileContents))*100)
	fmt.Println("-----------------------")

	result := 0
	for _, v := range bitstream {
		result += v
	}

	fmt.Println("Sum:", result)
	return encodeStreamAndLiterals(bitstream, literals)
}

func Decompress(fileContents []byte) []byte {
	bitstream, literals := decodeStreamAndLiterals(fileContents)
	output := decodeBytes(bitstream, literals)
	return output
}

func printTransitions(parent State, indentation int) {
	for _, state := range *parent.transitions {
		fmt.Print(strings.Repeat("-", indentation), state.displayValue(), "-", state.freq, "\n")

		if state.transitions != nil {
			printTransitions(*state, indentation+1)
		}
	}
}

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) io.WriteCloser {
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
	compressed   []byte
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
		if err != nil {
			return 0, err
		}
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
