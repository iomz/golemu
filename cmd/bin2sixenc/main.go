package main

import (
	"fmt"
	"os"

	"github.com/iomz/go-llrp/binutil"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Println("Must pass a string input")
	}

	bs := []rune(os.Args[1])

	var i int
	for i = 0; i+6 < len(bs); i += 6 {
		r, err := binutil.Parse6BinRuneSliceToRune([]rune(bs[i : i+6]))
		if err != nil {
			panic(err)
		}
		fmt.Printf(string(r))
		//_ = r
		//fmt.Println(i)
	}
	fmt.Printf("\nRemains: ")
	for ; i<len(bs); i++ {
		fmt.Printf(string(i))
	}
}
