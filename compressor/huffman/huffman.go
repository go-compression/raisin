package huffman

import (
	"container/heap"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/profile"
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
func remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
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
	fmt.Println("building tree")
	type sorter struct {
		Key   rune
		Value int
	}
	var keys []int
	var values []int
	for i, j := range symFreqs {
		keys = append(keys, int(i))
		values = append(values, j)
	}
	sort.Ints(keys)
	sort.Ints(values)
	var temp1 []rune
	var temp2 []int
	//symFreqs2 := make(map[rune]int)
	for _, value := range values {
		for i, key := range keys {
			if symFreqs[rune(key)] == value {
				temp1 = append(temp1, rune(key))
				temp2 = append(temp2, value)
				keys = remove(keys, i)
				break
			}
		}
	}
	//build tree
	var trees treeHeap
	for i := 0; i < len(symFreqs); i++ {
		trees = append(trees, HuffmanLeaf{temp2[i], temp1[i]})
	}
	heap.Init(&trees)
	//	estring = strconv.Itoa(len(symFreqs))
	//sort.Sort(trees)
	for trees.Len() > 1 {
		a := heap.Pop(&trees).(HuffmanTree)
		b := heap.Pop(&trees).(HuffmanTree)

		heap.Push(&trees, HuffmanNode{a.Freq() + b.Freq(), a, b})
	}
	return heap.Pop(&trees).(HuffmanTree)
}

func rebuildTree(symFreqs map[rune]int) HuffmanTree {
	type sorter struct {
		Key   rune
		Value int
	}
	var toSort []sorter
	for i, j := range symFreqs {
		toSort = append(toSort, sorter{i, j})
	}
	sort.Slice(toSort, func(i, j int) bool {
		return toSort[i].Value < toSort[j].Value
	})
	symFreqs2 := make(map[rune]int)
	for _, item := range toSort {
		symFreqs2[item.Key] = item.Value
	}

	var trees treeHeap
	for c, f := range symFreqs2 {
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

var answer string

func findCodes(tree HuffmanTree, og HuffmanTree, data string, i int, max int) string {
	if i <= max {
		switch huff := tree.(type) {
		case HuffmanLeaf:
			answer = answer + string(huff.value)
			if i < max {
				findCodes(og, og, data, i, max)
			} else {
				//fmt.Println(answer)
				file, err := os.Create("decompressed2.txt")
				check(err)
				_, err = io.WriteString(file, answer)
				check(err)

			}
		case HuffmanNode:
			if string(data[i]) == "0" {
				findCodes(huff.left, og, string(data), i+1, max)
			} else if string(data[i]) == "1" {
				findCodes(huff.right, og, string(data), i+1, max)
			}
		}
	} else {
		//fmt.Println(answer)
		file, err := os.Create("decompressed2.txt")
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
		ostring = ostring + "1"
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
	symFreqs := make(map[rune]int)
	temp := ""
	var freq int
	for i := 0; i < len(tree); i++ {
		if string(tree[i]) != "|" {
			temp = temp + string(tree[i])
		} else {
			freq, _ = strconv.Atoi(temp)
			temp = ""
			if string(tree[i+1]) == "\\" && string(tree[i+2]) == "n" {
				symFreqs[10] = freq
				i++
			} else {
				symFreqs[rune(tree[i+1])] = freq
			}
			i++
		}
	}
	//fmt.Print(symFreqs)
	return buildTree(symFreqs)
}
func encode(tree HuffmanTree, input string) []byte {
	fmt.Println("encoding")
	answer := ""
	tempV := make([]rune, 0)
	tempB := make([]string, 0)
	vals, bin := printCodes(tree, []byte{}, tempV, tempB)
	for i := 0; i < len(input); i++ {
		fmt.Println(string(rune(input[i])))
		if indexOf(rune(input[i]), vals) != -1 {
			answer = answer + (bin[indexOf(rune(input[i]), vals)])
		} else {
			answer = answer + (bin[len(bin)-1])
		}
	}

	for i := 0; i < len(vals); i++ {
		if vals[i] != 10 {
			ostring = ostring + string(bin[i]) + "3" + string(vals[i])
		} else {
			ostring = ostring + string(bin[i]) + "3\\n"
		}
	}
	//Println(len(answer))
	diff := bitString(string(strconv.FormatInt(int64(8-len(answer)%8), 2)))
	first := diff.AsByteSlice()
	bits := bitString(answer)
	final := bits.AsByteSlice()
	//	fmt.Println(diff)
	fmt.Println(bits)
	test := append(first, final...)

	return append([]byte(estring), append([]byte("\\\n"), test...)...)
}
func decode(fileContents []byte) []byte {
	fmt.Println("decoding")
	file_content := string(fileContents)
	lines := strings.Split(file_content, "\\\n")
	tree := decodeTree(lines[0])

	byteArr := []byte(strings.Join(strings.Split(string(fileContents), "\\\n")[1:], ""))
	content := make([]string, 0)
	contentString := ""
	var diff int64
	var err error
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
	fmt.Println(int(diff))
	contentString = contentString[int(diff):]
	fmt.Println(contentString)
	answer := findCodes(tree, tree, contentString, 0, len(contentString))
	return []byte(answer)
}

func Compress(fileContents []byte) []byte {
	content := string(fileContents)
	symFreqs := make(map[rune]int)

	for _, c := range content {
		symFreqs[c]++
	}
	for key, val := range symFreqs {
		if key != 10 {
			estring = estring + strconv.Itoa(val) + "|" + string(key)
		} else {
			estring = estring + strconv.Itoa(val) + "|\\n"
		}
	}
	exampleTree := buildTree(symFreqs)

	out := encode(exampleTree, content)

	return out
}

func Decompress(fileContents []byte) []byte {
	decoded := decode(fileContents)
	return decoded
}

func main() {
	defer profile.Start().Stop()
	fileContents, err := ioutil.ReadFile("huffman-input.txt")
	check(err)
	content := string(fileContents)
	symFreqs := make(map[rune]int)

	for _, c := range content {
		symFreqs[c]++
	}
	for key, val := range symFreqs {
		if key != 10 {
			estring = estring + strconv.Itoa(val) + "|" + string(key)
		} else {
			estring = estring + strconv.Itoa(val) + "|\\n"
		}
	}
	exampleTree := buildTree(symFreqs)

	out := encode(exampleTree, content)
	file, err := os.Create("huffman-compressed.bin")
	check(err)
	file.Write(out)

	fileContents, err2 := ioutil.ReadFile("huffman-compressed.bin")
	check(err2)
	decoded := decode(fileContents)

	file, err = os.Create("decompressed2.txt")
	check(err)
	_, err = io.WriteString(file, string(decoded))
	check(err)

}
