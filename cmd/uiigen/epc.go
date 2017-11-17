// This file contains EPC binary encoding schemes
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/iomz/go-llrp/binutil"
)

// GetFilterValue returns filter value as rune slice
func GetFilterValue(fv string) (filter []rune) {
	if fv != "" {
		n, _ := strconv.ParseInt(fv, 10, 32)
		filter = []rune(fmt.Sprintf("%.3b", n))
	} else {
		filter, _ = binutil.GenerateNLengthRandomBinRuneSlice(3, 7)
	}
	return
}

// GetItemReference converts ItemReference value to rune slice
func GetItemReference(ir string, cpSizes []int) (itemReference []rune) {
	if ir != "" {
		itemReference = binutil.ParseDecimalStringToBinRuneSlice(ir)
		// If the itemReference is short, pad zeroes to the left
		if 44-cpSizes[0] > len(itemReference) {
			leftPadding := binutil.GenerateNLengthZeroPaddingRuneSlice(44 - cpSizes[0] - len(itemReference))
			itemReference = append(leftPadding, itemReference...)
		} else if 44-cpSizes[0] > len(itemReference) {
			// if the resulting itemReference is bigger than it is supposed to be...
			panic("invalid itemreference!")
		}
	} else {
		itemReference, _ = binutil.GenerateNLengthRandomBinRuneSlice(44-cpSizes[0], uint(math.Pow(float64(10), float64(13-cpSizes[1]))))
	}
	return
}

// GetPartitionAndCompanyPrefix returns partition value and companyPrefix and each size
// takes companyprefix raw string
func GetPartitionAndCompanyPrefix(cp string) (partition []rune, companyPrefix []rune, cpSizes []int) {
	var pValue uint
	// If company prefix is already supplied
	if cp != "" {
		companyPrefix = binutil.ParseDecimalStringToBinRuneSlice(cp)
		switch len(cp) {
		case 12:
			pValue = 0
			cpSizes = []int{40, 12}
		case 11:
			pValue = 1
			cpSizes = []int{37, 11}
		case 10:
			pValue = 2
			cpSizes = []int{34, 10}
		case 9:
			pValue = 3
			cpSizes = []int{30, 9}
		case 8:
			pValue = 4
			cpSizes = []int{27, 8}
		case 7:
			pValue = 5
			cpSizes = []int{24, 7}
		case 6:
			pValue = 6
			cpSizes = []int{20, 6}
		}
		partition = []rune(fmt.Sprintf("%.3b", pValue))
		// If the companyPrefix is short, pad zeroes to the left
		if len(companyPrefix) != cpSizes[0] {
			leftPadding := binutil.GenerateNLengthZeroPaddingRuneSlice(cpSizes[0] - len(companyPrefix))
			companyPrefix = append(leftPadding, companyPrefix...)
		}
	} else {
		// SGTIN Partition Table
		partition, pValue = binutil.GenerateNLengthRandomBinRuneSlice(3, 6)
		switch pValue {
		case 0:
			cpSizes = []int{40, 12}
		case 1:
			cpSizes = []int{37, 11}
		case 2:
			cpSizes = []int{34, 10}
		case 3:
			cpSizes = []int{30, 9}
		case 4:
			cpSizes = []int{27, 8}
		case 5:
			cpSizes = []int{24, 7}
		case 6:
			cpSizes = []int{20, 6}
		}
		companyPrefix, _ = binutil.GenerateNLengthRandomBinRuneSlice(cpSizes[0], uint(math.Pow(float64(10), float64(cpSizes[1]))))
	}
	return
}

// GetSerial converts serial to rune slice
func GetSerial(s string, serialLength int) (serial []rune) {
	if s != "" {
		serial = binutil.ParseDecimalStringToBinRuneSlice(s)
		if serialLength > len(serial) {
			leftPadding := binutil.GenerateNLengthZeroPaddingRuneSlice(serialLength - len(serial))
			serial = append(leftPadding, serial...)
		}
	} else {
		serial, _ := binutil.GenerateNLengthRandomBinRuneSlice(serialLength, uint(math.Pow(float64(2), float64(serialLength))))
		_ = serial
	}
	return serial
}

// MakeRuneSliceOfGIAI96 generates GIAI-96
func MakeRuneSliceOfGIAI96(cp string, fv string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	// TODO: cap overflow
	indivisualAssetReference, _ := binutil.GenerateNLengthRandomBinRuneSlice(82-cpSizes[0], uint(math.Pow(float64(10), float64(25-cpSizes[1]))))

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, indivisualAssetReference...)

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := binutil.ParseBinRuneSliceToUint8Slice(bs)
	if err != nil {
		fmt.Println("Something went wrong!")
		fmt.Println(err)
		os.Exit(1)
	}

	var giai96 = []interface{}{
		uint8(52), // GIAI-96 Header 0011 0100
		p[0],      // 8 bits -> 16 bits
		p[1],      // 8 bits -> 24 bits
		p[2],      // 8 bits -> 32 bits
		p[3],      // 8 bits -> 40 bits
		p[4],      // 8 bits -> 48 bits
		p[5],      // 8 bits -> 56 bits
		p[6],      // 8 bits -> 64 bits
		p[7],      // 8 bits -> 72 bits
		p[8],      // 8 bits -> 80 bits
		p[9],      // 8 bits -> 88 bits
		p[10],     // 8 bits -> 96 bits
	}

	return binutil.Pack(giai96), nil
}

// MakeRuneSliceOfGRAI96 generates GRAI-96
func MakeRuneSliceOfGRAI96(cp string, fv string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	assetType, _ := binutil.GenerateNLengthRandomBinRuneSlice(44-cpSizes[0], uint(math.Pow(float64(10), float64(12-cpSizes[1]))))
	serial, _ := binutil.GenerateNLengthRandomBinRuneSlice(38, 0)

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, assetType...)
	bs = append(bs, serial...)

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := binutil.ParseBinRuneSliceToUint8Slice(bs)
	if err != nil {
		fmt.Println("Something went wrong!")
		fmt.Println(err)
		os.Exit(1)
	}

	var grai96 = []interface{}{
		uint8(51), // GRAI-96 Header 0011 0011
		p[0],      // 8 bits -> 16 bits
		p[1],      // 8 bits -> 24 bits
		p[2],      // 8 bits -> 32 bits
		p[3],      // 8 bits -> 40 bits
		p[4],      // 8 bits -> 48 bits
		p[5],      // 8 bits -> 56 bits
		p[6],      // 8 bits -> 64 bits
		p[7],      // 8 bits -> 72 bits
		p[8],      // 8 bits -> 80 bits
		p[9],      // 8 bits -> 88 bits
		p[10],     // 8 bits -> 96 bits
	}

	return binutil.Pack(grai96), nil
}

// MakeRuneSliceOfSGTIN96 generates SGTIN-96
func MakeRuneSliceOfSGTIN96(cp string, fv string, ir string, s string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	itemReference := GetItemReference(ir, cpSizes)
	serial := GetSerial(s, 38)

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, itemReference...)
	bs = append(bs, serial...)

	//fmt.Println("EPC Header: %v", "00110000")
	//fmt.Println("Filter: %v", string(filter))
	//fmt.Println("Partition: %v", string(partition))
	//fmt.Println("GS1 Company Prefix: %v", string(companyPrefix))
	//fmt.Println("Item Reference: %v", string(itemReference))
	//fmt.Println("Serial: %v", string(serial))

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := binutil.ParseBinRuneSliceToUint8Slice(bs)
	if err != nil {
		fmt.Println("Something went wrong!")
		fmt.Println(err)
		os.Exit(1)
	}

	var sgtin96 = []interface{}{
		uint8(48), // SGTIN-96 Header 0011 0000
		p[0],      // 8 bits -> 16 bits
		p[1],      // 8 bits -> 24 bits
		p[2],      // 8 bits -> 32 bits
		p[3],      // 8 bits -> 40 bits
		p[4],      // 8 bits -> 48 bits
		p[5],      // 8 bits -> 56 bits
		p[6],      // 8 bits -> 64 bits
		p[7],      // 8 bits -> 72 bits
		p[8],      // 8 bits -> 80 bits
		p[9],      // 8 bits -> 88 bits
		p[10],     // 8 bits -> 96 bits
	}

	return binutil.Pack(sgtin96), nil
}

// MakeRuneSliceOfSSCC96 generates SSCC-96
func MakeRuneSliceOfSSCC96(cp string, fv string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	extension, _ := binutil.GenerateNLengthRandomBinRuneSlice(58-cpSizes[0], uint(math.Pow(float64(10), float64(17-cpSizes[1]))))

	// 24 '0's
	reserved := binutil.GenerateNLengthZeroPaddingRuneSlice(24)

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, extension...)
	bs = append(bs, reserved...)

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := binutil.ParseBinRuneSliceToUint8Slice(bs)
	if err != nil {
		fmt.Println("Something went wrong!")
		fmt.Println(err)
		os.Exit(1)
	}

	var sscc96 = []interface{}{
		uint8(49), // SSCC-96 Header 0011 0001
		p[0],      // 8 bits -> 16 bits
		p[1],      // 8 bits -> 24 bits
		p[2],      // 8 bits -> 32 bits
		p[3],      // 8 bits -> 40 bits
		p[4],      // 8 bits -> 48 bits
		p[5],      // 8 bits -> 56 bits
		p[6],      // 8 bits -> 64 bits
		p[7],      // 8 bits -> 72 bits
		p[8],      // 8 bits -> 80 bits
		p[9],      // 8 bits -> 88 bits
		p[10],     // 8 bits -> 96 bits
	}

	return binutil.Pack(sscc96), nil
}
