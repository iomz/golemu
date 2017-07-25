package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
	//"strconv"
)

func GenerateNLengthBinaryString(n int, max uint) []rune {
	binstr := make([]rune, n)
	sum := uint(0)
	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < n; i++ {
		var b rune
		if rand.Intn(2) == 0 {
			b = '0'
		} else {
			b = '1'
		}
		binstr[i] = b
		if b == '1' {
			sum += uint(math.Pow(float64(2), float64(n - i - 1)))
		}
	}

	fmt.Println("sum: ", sum)

	if max < sum {
		binstr = GenerateNLengthBinaryString(n, max)
	}

	return binstr
}

// Precondition: len(bs) % 8 == 0
func BinaryString2uint8(bs []rune) ([]uint8, error) {
	if len(bs) % 8 != 0 {
		return nil, errors.New("non-8 bit length binary string passed to BinaryString2uint8")
	} else if len(bs) < 8 {
		return nil, errors.New("binary string length less than 8 given to BinaryString2uint8")
	}

	bsSize := len(bs) / 8
	uints := make([]uint8, bsSize)

	for j := 0; j < bsSize; j++ {
		uintRep := uint8(0)
		for i := 0; i < 8; i++ {
			if bs[j*8 - i + 7] == '1' {
				uintRep += uint8(math.Pow(float64(2), float64(i)))
			}
		}
		uints[j] = uintRep
	}

	return uints, nil
}

func main() {
	bs1 := GenerateNLengthBinaryString(5, 25)
	bs2 := GenerateNLengthBinaryString(3, 6)
	bs3 := GenerateNLengthBinaryString(16, 50000)
	bs := append(bs1, bs2...)
	bs = append(bs, bs3...)
	fmt.Println(string(bs))
	uints, err := BinaryString2uint8(bs)
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(uints)
}
