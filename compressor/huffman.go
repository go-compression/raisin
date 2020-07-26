package main

import (
	"container/heap"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

type HuffmanTree interface {
	Freq() int
}

type HuffmanLeaf struct {
	freq  int
	value rune
}

type HuffmanNode struct {
	freq        int
	left, right HuffmanTree
}

func (self HuffmanLeaf) Freq() int {
	return self.freq
}

func (self HuffmanNode) Freq() int {
	return self.freq
}

type treeHeap []HuffmanTree

func (th treeHeap) Len() int { return len(th) }
func (th treeHeap) Less(i, j int) bool {
	return th[i].Freq() < th[j].Freq()
}
func (th *treeHeap) Push(ele interface{}) {
	*th = append(*th, ele.(HuffmanTree))
}
func (th *treeHeap) Pop() (popped interface{}) {
	popped = (*th)[len(*th)-1]
	*th = (*th)[:len(*th)-1]
	return
}
func (th treeHeap) Swap(i, j int) { th[i], th[j] = th[j], th[i] }

var estring string

func buildTree(symFreqs map[rune]int) HuffmanTree {
	// var trees treeHeap
	// for c, f := range symFreqs {
	// 	trees = append(trees, HuffmanLeaf{f, c})
	// }
	// heap.Init(&trees)

	// sort.Sort(sort.Reverse(test))

	var trees treeHeap
	for c, f := range symFreqs {
		trees = append(trees, HuffmanLeaf{f, c})
	}
	heap.Init(&trees)
	sort.Sort(trees)
	new := trees
	sort.Sort(new)
	count := len(symFreqs) * 2
	prev := 0
	var test treeHeap
	for j := 0; j < len(new); j++ {
		temp := new[j]

		switch huff := temp.(type) {
		case HuffmanLeaf:
			if j == 0 {
				prev = huff.freq
				test = append(test, HuffmanLeaf{count, huff.value})
			} else {
				if huff.freq < prev {
					count += 2
					prev = huff.freq
					test = append(test, HuffmanLeaf{count, huff.value})
				} else {
					count += 1
					prev = huff.freq
					test = append(test, HuffmanLeaf{count, huff.value})
				}

			}

		}
	}
	prev = 0
	sort.Sort(trees)
	for i := 0; i < len(trees); i++ {
		temp := new[i]

		switch huff := temp.(type) {
		case HuffmanLeaf:

			if i == 0 {
				prev = huff.freq
				if huff.value == 10 {
					estring = estring + "\\n"
				} else {
					estring = estring + string(huff.value)
				}
			} else {
				if huff.freq == prev+1 {
					if huff.value == 10 {
						estring = estring + "0" + "\\n"
					} else {
						estring = estring + "0" + string(huff.value)
					}
				} else {
					if huff.value == 10 {
						estring = estring + "1" + "\\n"
					} else {
						estring = estring + "1" + string(huff.value)
					}
				}
				prev = huff.freq
			}

		}
	}
	for test.Len() > 1 {
		a := heap.Pop(&test).(HuffmanTree)
		b := heap.Pop(&test).(HuffmanTree)

		heap.Push(&test, HuffmanNode{a.Freq() + b.Freq(), a, b})
	}
	return heap.Pop(&test).(HuffmanTree)
}
func rebuildTree(symFreqs map[rune]int) HuffmanTree {
	var trees treeHeap
	for c, f := range symFreqs {
		trees = append(trees, HuffmanLeaf{f, c})
	}
	heap.Init(&trees)
	sort.Sort(sort.Reverse(trees))
	for trees.Len() > 1 {
		a := heap.Pop(&trees).(HuffmanTree)
		b := heap.Pop(&trees).(HuffmanTree)

		heap.Push(&trees, HuffmanNode{a.Freq() + b.Freq(), a, b})
	}

	return heap.Pop(&trees).(HuffmanTree)
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
func printCodes(tree HuffmanTree, prefix []byte, vals []rune, bin []string) ([]rune, []string) {
	switch i := tree.(type) {
	case HuffmanLeaf:
		vals = append(vals, rune(i.value))
		bin = append(bin, string(prefix))
		fmt.Printf("%c\t%d\t%s\n", i.value, i.freq, string(prefix))
		return vals, bin
	case HuffmanNode:
		prefix = append(prefix, '0')
		vals, bin = printCodes(i.left, prefix, vals, bin)
		prefix = prefix[:len(prefix)-1]

		prefix = append(prefix, '1')
		vals, bin = printCodes(i.right, prefix, vals, bin)
		prefix = prefix[:len(prefix)-1]
	}
	return vals, bin
}
func findCodes(tree HuffmanTree, og HuffmanTree, data string, answer string, i int, max int) string {
	if i <= max {
		switch huff := tree.(type) {
		case HuffmanLeaf:
			answer = answer + string(huff.value)
			if i < max-1 {
				findCodes(og, og, data, answer, i, max)
			} else {
				fmt.Println(answer)
				file, err := os.Create("decompressed.txt")
				check(err)
				_, err = io.WriteString(file, answer)
				check(err)

				return answer
			}
		case HuffmanNode:
			if string(data[i]) == "0" {
				answer = findCodes(huff.left, og, string(data), answer, i+1, max)
			} else if string(data[i]) == "1" {
				answer = findCodes(huff.right, og, string(data), answer, i+1, max)
			}
		}
	} else {
		fmt.Println(answer)
		file, err := os.Create("decompressed.txt")
		check(err)
		_, err = io.WriteString(file, answer)
		check(err)

		return answer
	}
	return answer
}
func indexOf(word rune, data []rune) int {
	for k, v := range data {
		if word == v {
			return k
		}
	}
	return -1
}
func indexOfString(word string, data []string) int {
	for k, v := range data {
		if word == v {
			return k
		}
	}
	return -1
}

type bitString string

func (b bitString) AsByteSlice() []byte {
	var out []byte
	var str string

	for i := len(b); i > 0; i -= 8 {
		if i-8 < 0 {
			str = string(b[0:i])
		} else {
			str = string(b[i-8 : i])
		}
		v, err := strconv.ParseUint(str, 2, 8)
		if err != nil {
			panic(err)
		}
		out = append([]byte{byte(v)}, out...)
	}
	return out
}

var ostring string

func encodeTree(tree HuffmanTree) {
	switch huff := tree.(type) {
	case HuffmanLeaf:
		ostring = estring + "1"
		if huff.value != 10 {
			ostring = ostring + string(huff.value)
		} else {
			ostring = ostring + "\\n"
		}
	case HuffmanNode:
		ostring = ostring + "0"
		encodeTree(huff.right)
		encodeTree(huff.left)
	}
}

var decodedTree HuffmanTree
var treeH treeHeap

func decodeTree(tree string) HuffmanTree {
	count := 1
	symFreqs := make(map[rune]int)
	j := 0
	if rune(tree[j]) == 92 && rune(tree[j+1]) == 110 {
		symFreqs[10] = count
		j += 2
	} else {
		symFreqs[rune(tree[j])] = count
		j = 1
	}
	for i := j; i < len(tree)-1; i++ {
		if string(tree[i]) == "0" {
			if rune(tree[i+1]) == 92 && rune(tree[i+2]) == 110 {
				count++
				symFreqs[10] = count
				i++
			} else {
				count++
				symFreqs[rune(tree[i+1])] = count
			}
		} else {
			count += 2
			if rune(tree[i+1]) == 92 && rune(tree[i+2]) == 110 {
				symFreqs[10] = count
				i++
			} else {
				symFreqs[rune(tree[i+1])] = count
			}
		}
		i++

	}
	return rebuildTree(symFreqs)
}
func encode(tree HuffmanTree, input string) {

	answer := ""
	tempV := make([]rune, 0)
	tempB := make([]string, 0)
	vals, bin := printCodes(tree, []byte{}, tempV, tempB)
	for i := 0; i < len(input); i++ {
		answer = answer + (bin[indexOf(rune(input[i]), vals)])
	}

	encodeTree(tree)
	fmt.Println(len(answer))
	diff := bitString(string(strconv.FormatInt(int64(8-len(answer)%8), 2)))
	first := diff.AsByteSlice()
	bits := bitString(answer)
	final := bits.AsByteSlice()
	fmt.Println(diff)
	fmt.Println(bits)
	test := append(first, final...)

	file, err := os.Create("huffman-compressed.bin")
	check(err)
	file.WriteString(estring + "\n")
	file.Write(test)
	decode()

}
func decode() {
	fileContents, err := ioutil.ReadFile("huffman-compressed.bin")
	check(err)
	file_content := string(fileContents)
	lines := strings.Split(file_content, "\n")
	tree := decodeTree(lines[0])
	tempV := make([]rune, 0)
	tempB := make([]string, 0)
	_, bin := printCodes(tree, []byte{}, tempV, tempB)
	fmt.Println(bin)
	var real string
	for i := 1; i < len(lines); i++ {
		real = real + lines[i]
	}

	byteArr := []byte(real)
	content := make([]string, 0)
	contentString := ""
	var diff int64
	for i, n := range byteArr {
		if i != 0 {
			hold := fmt.Sprintf("%08b", n)
			content = append(content, hold)
			contentString = contentString + hold
		} else {
			hold := fmt.Sprintf("%08b", n)
			diff, err = strconv.ParseInt(hold, 2, 64)
			check(err)
		}
	}
	contentString = contentString[int(diff):]
	findCodes(tree, tree, contentString, "", 0, len(contentString))

}
func main() {
	fileContents, err := ioutil.ReadFile("huffman-input.txt")
	check(err)
	content := string(fileContents)
	symFreqs := make(map[rune]int)

	for _, c := range content {
		symFreqs[c]++
	}
	exampleTree := buildTree(symFreqs)

	encode(exampleTree, content)
}
