// A tool to generate arbitrary UII (aka EPC)
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// kingpin app
	app = kingpin.New("uiigen", "A tool to generate an arbitrary UII (aka EPC).")

	// kingpin generate EPC mode
	epc = app.Command("epc", "Generate an EPC.")
	// EPC scheme
	epcScheme        = epc.Flag("type", "EPC UII type.").Short('t').Default("SGTIN-96").String()
	epcCompanyPrefix = epc.Flag("companyPrefix", "Company Prefix for EPC UII.").Short('c').Default("").String()
	epcFilterValue   = epc.Flag("filterValue", "Filter Value for EPC UII.").Short('f').Default("").String()
	epcItemReference = epc.Flag("itemReference", "Item Reference Value for EPC UII.").Short('i').Default("").String()
	epcSerial        = epc.Flag("serial", "Serial value for EPC UII.").Short('s').Default("").String()

	// kingpin generate ISO UII mode
	iso = app.Command("iso", "Generate an ISO UII.")
	// ISO scheme
	isoScheme = iso.Flag("scheme", "Scheme for ISO UII.").Short('s').Default("17367").String()

	//hexRunes to hold hex chars
	hexRunes = []rune("abcdef0123456789")
)

// Check if string exists in a string slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Pack the data into (partial) LLRP packet payload.
func Pack(data []interface{}) []byte {
	buf := new(bytes.Buffer)
	for _, v := range data {
		binary.Write(buf, binary.BigEndian, v)
	}
	return buf.Bytes()
}

// Return random int value with max
func randomInt(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

// Return random hex rune for n length
func RandomHexRunes(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range b {
		b[i] = hexRunes[rand.Intn(len(hexRunes))]
	}
	return string(b)
}

// Make binary string from hex string
func bigIntToBin(cp *big.Int) string {
	binStr := fmt.Sprintf("%b", cp)
	return binStr
}

// Convert hex string to binary string
func hexStringToBinString(s string) (binString string) {
	for _, c := range s {
		n, _ := strconv.ParseInt(string(c), 16, 32)
		binString = fmt.Sprintf("%s%.4b", binString, n)
	}
	return
}

// Generate zero padding string
func GenerateNLengthZeroBinaryString(n int) []rune {
	binstr := make([]rune, n)

	for i := 0; i < n; i++ {
		binstr[i] = '0'
	}

	return binstr
}

// max == 0 for no cap limit
func GenerateNLengthBinaryString(n int, max uint) ([]rune, uint) {
	binstr := make([]rune, n)
	var sum uint
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
			sum += uint(math.Pow(float64(2), float64(n-i-1)))
		}
	}

	if max != uint(0) && max < sum {
		binstr, sum = GenerateNLengthBinaryString(n, max)
	}

	return binstr, sum
}

// Precondition: len(bs) % 8 == 0
func BinaryString2uint8(bs []rune) ([]uint8, error) {
	if len(bs)%8 != 0 {
		return nil, errors.New("non-8 bit length binary string passed to BinaryString2uint8")
	} else if len(bs) < 8 {
		return nil, errors.New("binary string length less than 8 given to BinaryString2uint8")
	}

	bsSize := len(bs) / 8
	uints := make([]uint8, bsSize)

	for j := 0; j < bsSize; j++ {
		uintRep := uint8(0)
		for i := 0; i < 8; i++ {
			if bs[j*8-i+7] == '1' {
				uintRep += uint8(math.Pow(float64(2), float64(i)))
			}
		}
		uints[j] = uintRep
	}

	return uints, nil
}

func SerialToBinary(s string) []rune {
	n, _ := strconv.ParseInt(s, 10, 64)
	binStr := bigIntToBin(big.NewInt(n))
	return []rune(binStr)
}

func GetFilterValue(fv string) (filter []rune) {
	if fv != "" {
		n, _ := strconv.ParseInt(fv, 10, 32)
		filter = []rune(fmt.Sprintf("%.3b", n))
	} else {
		filter, _ = GenerateNLengthBinaryString(3, 7)
	}
	return
}

func GetPartitionAndCompanyPrefix(cp string) (partition []rune, companyPrefix []rune, cpSizes []int) {
	var pValue uint
	// If company prefix is already supplied
	if cp != "" {
		companyPrefix = SerialToBinary(cp)
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
			leftPadding := GenerateNLengthZeroBinaryString(cpSizes[0] - len(companyPrefix))
			companyPrefix = append(leftPadding, companyPrefix...)
		}
	} else {
		// SGTIN Partition Table
		partition, pValue = GenerateNLengthBinaryString(3, 6)
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
		companyPrefix, _ = GenerateNLengthBinaryString(cpSizes[0], uint(math.Pow(float64(10), float64(cpSizes[1]))))
	}
	return
}

func GetItemReference(ir string, cpSizes []int) (itemReference []rune) {
	if ir != "" {
		itemReference = SerialToBinary(ir)
		// If the itemReference is short, pad zeroes to the left
		if 44-cpSizes[0] > len(itemReference) {
			leftPadding := GenerateNLengthZeroBinaryString(44 - cpSizes[0] - len(itemReference))
			itemReference = append(leftPadding, itemReference...)
		} else if 44-cpSizes[0] > len(itemReference) {
			// if the resulting itemReference is bigger than it is supposed to be...
			panic("invalid itemreference!")
		}
	} else {
		itemReference, _ = GenerateNLengthBinaryString(44-cpSizes[0], uint(math.Pow(float64(10), float64(13-cpSizes[1]))))
	}
	return
}

func GetSerial(s string, serialLength int) (serial []rune) {
	if s != "" {
		serial = SerialToBinary(s)
		if serialLength > len(serial) {
			leftPadding := GenerateNLengthZeroBinaryString(serialLength - len(serial))
			serial = append(leftPadding, serial...)
		}
	} else {
			serial, _ := GenerateNLengthBinaryString(serialLength, uint(math.Pow(float64(2), float64(serialLength))))
			_ = serial
	}
	return serial
}

// SGTIN-96
func GenerateRandomSGTIN96(cp string, fv string, ir string, s string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	itemReference := GetItemReference(ir, cpSizes)
	serial := GetSerial(s, 38)

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, itemReference...)
	bs = append(bs, serial...)

	fmt.Println("EPC Header: %s", "00110000")
	fmt.Println("Filter: %s", string(filter))
	fmt.Println("Partition: %s", string(partition))
	fmt.Println("GS1 Company Prefix: %s", string(companyPrefix))
	fmt.Println("Item Reference: %s", string(itemReference))
	fmt.Println("Serial: %s", string(serial))

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := BinaryString2uint8(bs)
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

	return Pack(sgtin96), nil
}

// SSCC-96
func GenerateRandomSSCC96(cp string, fv string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	extension, _ := GenerateNLengthBinaryString(58-cpSizes[0], uint(math.Pow(float64(10), float64(17-cpSizes[1]))))

	// 24 '0's
	reserved := GenerateNLengthZeroBinaryString(24)

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, extension...)
	bs = append(bs, reserved...)

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := BinaryString2uint8(bs)
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

	return Pack(sscc96), nil
}

// GRAI-96
func GenerateRandomGRAI96(cp string, fv string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	assetType, _ := GenerateNLengthBinaryString(44-cpSizes[0], uint(math.Pow(float64(10), float64(12-cpSizes[1]))))
	serial, _ := GenerateNLengthBinaryString(38, 0)

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, assetType...)
	bs = append(bs, serial...)

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := BinaryString2uint8(bs)
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

	return Pack(grai96), nil
}

// GIAI-96
func GenerateRandomGIAI96(cp string, fv string) ([]byte, error) {
	filter := GetFilterValue(fv)
	partition, companyPrefix, cpSizes := GetPartitionAndCompanyPrefix(cp)
	// TODO: cap overflow
	indivisualAssetReference, _ := GenerateNLengthBinaryString(82-cpSizes[0], uint(math.Pow(float64(10), float64(25-cpSizes[1]))))

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, indivisualAssetReference...)

	if len(bs) != 88 {
		fmt.Println("Something went wrong!")
		fmt.Println("len(bs): ", len(bs))
		os.Exit(1)
	}

	p, err := BinaryString2uint8(bs)
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

	return Pack(giai96), nil
}

//
func GenerateEPC() string {
	epcs := []string{"SGTIN-96", "SSCC-96", "GRAI-96", "GIAI-96"}

	if !stringInSlice(strings.ToUpper(*epcScheme), epcs) {
		os.Exit(1)
	}

	var uii []byte
	switch strings.ToUpper(*epcScheme) {
	case "SGTIN-96":
		uii, _ = GenerateRandomSGTIN96(*epcCompanyPrefix, *epcFilterValue, *epcItemReference, *epcSerial)
	case "SSCC-96":
		uii, _ = GenerateRandomSSCC96(*epcCompanyPrefix, *epcFilterValue)
	case "GRAI-96":
		uii, _ = GenerateRandomGRAI96(*epcCompanyPrefix, *epcFilterValue)
	case "GIAI-96":
		uii, _ = GenerateRandomGIAI96(*epcCompanyPrefix, *epcFilterValue)
	}

	// TODO: update pc when length changed (for non-96-bit codes)
	pc := Pack([]interface{}{
		uint8(48), // L4-0=11000(6words=96bits), UMI=0, XI=0
		uint8(0),  // RFU=0
	})

	length := uint16(18)
	epclen := uint16(96)

	return hex.EncodeToString(pc) + "," +
		strconv.FormatUint(uint64(length), 10) + "," +
		strconv.FormatUint(uint64(epclen), 10) + "," +
		hex.EncodeToString(uii) + "\n" +
		hexStringToBinString(hex.EncodeToString(uii))
}

// Generate a random 17367 code
func GenerateRandom17367() []byte {

	return []byte{}
}

//
func GenerateISO() string {
	var uii []byte
	var pc []byte
	var length uint16
	var epclen uint16

	isos := []string{"17365", "17363"}

	if !stringInSlice(*isoScheme, isos) {
		os.Exit(1)
	}

	switch *isoScheme {
	case "17367":
		uii = GenerateRandom17367()
	}
	/*
		if isos[i] == 17365 {
			// ISO 17365
			uii, _ = hex.DecodeString("c4a301c70d36cb32920b1d" + RandomHexRunes(10))
			pc = Pack([]interface{}{
				// 16802 : ISO 17365
				uint8(65),  // L4-0=01000(5words=80bits), UMI=0, XI=0, T=1 : ISO 17365
				uint8(162), // AFI=10100010 : 17365
			})
			// 22, 128 : ISO 17365
			length = uint16(22)
			epclen = uint16(128)
		} else if isos[i] == 17363 {
			// ISO 17363
			uii, _ = hex.DecodeString("dc20420c4c" + RandomHexRunes(10))
			pc = Pack([]interface{}{
				// 10665 : ISO 17363
				uint8(41),  // L4-0=00101(8words=128bits), UMI=0, XI=0, T=1 : ISO 17363
				uint8(169), // AFI=10101001 : 17363
			})
			// 16, 80 : ISO 17363
			length = uint16(16)
			epclen = uint16(80)
		}
	*/

	return hex.EncodeToString(pc) + "," +
		strconv.FormatUint(uint64(length), 10) + "," +
		strconv.FormatUint(uint64(epclen), 10) + "," +
		hex.EncodeToString(uii) + "\n" +
		hexStringToBinString(hex.EncodeToString(uii))
}

func main() {
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch parse {
	case epc.FullCommand():
		fmt.Println(GenerateEPC())
	case iso.FullCommand():
		fmt.Println(GenerateISO())
	}
}
