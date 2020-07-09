<?php
$name = get_post_meta($post->ID, "certified_user_name", true);
$course = get_post_meta($post->ID, "course_name", true);
?>
<div style="width:800px; height:600px; padding:20px; text-align:center; border: 10px solid #787878">
<div style="width:750px; height:550px; padding:20px; text-align:center; border: 5px solid #787878">
       <span style="font-size:50px; font-weight:bold">Certificate of Completion</span>
       <br><br>
       <span style="font-size:25px"><i>This is to certify that</i></span>
       <br><br>
       <span style="font-size:30px"><b><?php echo $name; ?></b></span><br/><br/>
       <span style="font-size:25px"><i>has completed the course</i></span> <br/><br/>
       <span style="font-size:30px"><?php echo $course; ?></span> <br/><br/>
</div>
</div>

package main

import (
	"container/heap"
	"fmt"
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

func printCodes(tree HuffmanTree, prefix []byte) {
	switch i := tree.(type) {
	case HuffmanLeaf:
		fmt.Printf("%c\t%d\t%s\n", i.value, i.freq, string(prefix))
	case HuffmanNode:
		prefix = append(prefix, '0')
		printCodes(i.left, prefix)
		prefix = prefix[:len(prefix)-1]

		prefix = append(prefix, '1')
		printCodes(i.right, prefix)
		prefix = prefix[:len(prefix)-1]
	}
}

func main() {
	test := "This is a test string which contains many t's"

	symFreqs := make(map[rune]int)

	for _, c := range test {
		symFreqs[c]++
	}

	exampleTree := buildTree(symFreqs)

	fmt.Println("")
	printCodes(exampleTree, []byte{})
}
