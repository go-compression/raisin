package main

import (
	"io/ioutil"
)

// Algorithim Deatils
/*
	If the symbol has not occured in context, return escape symbol
		Use the next smaller context

	Each time a a symbol is encountered, the count corresponding to that symbol is updated in each table.


*/
func generate_prediction_tables() {

}
func main() {
	//defer profile.Start().Stop()
	fileContents, err := ioutil.ReadFile("huffman-input.txt")

}
