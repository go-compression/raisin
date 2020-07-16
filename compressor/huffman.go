package compressor

import (
	"container/heap"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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

func buildTree(symFreqs map[rune]int) HuffmanTree {
	var trees treeHeap
	for c, f := range symFreqs {
		trees = append(trees, HuffmanLeaf{f, c})
	}
	heap.Init(&trees)
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

func encode(tree HuffmanTree, input string) {

	answer := ""
	tempV := make([]rune, 0)
	tempB := make([]string, 0)
	vals, bin := printCodes(tree, []byte{}, tempV, tempB)
	for _, char := range input {
		diff := len(string(bin[indexOf(char, vals)])) % 8
		if len(string(bin[indexOf(char, vals)])) < 8 {
			diff = (8 - len(string(bin[indexOf(char, vals)])))
		}
		answer = answer + (bin[indexOf(char, vals)])
		for i := 0; i < diff; i++ {
			answer = answer + "0"
		}
	}
	bits := bitString(answer)
	final := bits.AsByteSlice()
	fmt.Println(final)
	permissions := 0664
	err := ioutil.WriteFile("file.txt", final, os.FileMode(permissions))
	check(err)
	decode(tree)

}
func decode(tree HuffmanTree) {
	fileContents, err := ioutil.ReadFile("file.txt")
	check(err)
	byteArr := fileContents
	content := make([]string, 0)
	for _, n := range byteArr {
		hold := fmt.Sprintf("%08b", n)
		content = append(content, hold)
	}
	fmt.Println(content)
	tempV := make([]rune, 0)
	tempB := make([]string, 0)
	vals, bin := printCodes(tree, []byte{}, tempV, tempB)
	for i, item := range bin {
		new := item
		diff := len(item) % 8
		if len(item) < 8 {
			diff = (8 - len(item))
		}
		for i := 0; i < diff; i++ {
			new = new + "0"
		}
		bin[i] = new
	}
	for _, item := range content {
		fmt.Print(string(vals[indexOfString(item, bin)]))
	}
}
func main() {
	fileContents, err := ioutil.ReadFile("huffman-input.txt")
	check(err)
	content := string(fileContents)
	//fmt.Println(content)
	symFreqs := make(map[rune]int)

	for _, c := range content {
		symFreqs[c]++
	}

	exampleTree := buildTree(symFreqs)

	// fmt.Println("")
	encode(exampleTree, content)
	decode(exampleTree)
}
