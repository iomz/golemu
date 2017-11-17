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

	for i := 0; i < len(bs); i++ {
		r := binutil.ParseRuneTo6BinRuneSlice(bs[i])
		fmt.Printf(string(r))
	}
}
