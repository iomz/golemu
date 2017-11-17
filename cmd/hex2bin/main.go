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

	bs, err := binutil.ParseHexStringToBinString(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Printf(bs)
}
