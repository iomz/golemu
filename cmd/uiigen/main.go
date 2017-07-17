// A tool to generate arbitrary UII (aka EPC)
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
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
	epcScheme = epc.Flag("scheme", "Scheme for EPC UII.").Short('s').Default("SGTIN-96").String()

	// kingpin generate ISO UII mode
	iso = app.Command("iso", "Generate an ISO UII.")

	//hexRunes to hold hex chars
	hexRunes = []rune("abcdef0123456789")
)


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

func randomInt(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

func RandomHexRunes(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range b {
		b[i] = hexRunes[rand.Intn(len(hexRunes))]
	}
	return string(b)
}

func GenerateNLengthBinaryString(n int, max int) ([]rune, int) {
	binstr := make([]rune, n)
	var sum int
	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < n; i++ {
		var b rune
		if max == 0 {
			b = '0'
		} else if rand.Intn(2) == 0 {
			b = '0'
		} else {
			b = '1'
		}
		binstr[i] = b
		if b == '1' {
			sum += int(math.Pow(float64(2), float64(n - i - 1)))
		}
	}

	//fmt.Println("sum: ", sum)

	if max != 0 && max < sum {
		binstr, sum = GenerateNLengthBinaryString(n, max)
	}

	return binstr, sum
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

// SGTIN-96
func GenerateRandomSGTIN96() ([]byte, error) {
	filter, _ := GenerateNLengthBinaryString(3, 7)
	partition, pValue := GenerateNLengthBinaryString(3, 6)

	// SGTIN Partition Table
	cpSizes := make([]int, 2)
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

	companyPrefix, _ := GenerateNLengthBinaryString(cpSizes[0], int(math.Pow(float64(10), float64(cpSizes[1]))))
	itemReference, _ := GenerateNLengthBinaryString(44-cpSizes[0], int(math.Pow(float64(10), float64(13-cpSizes[1]))))
	serial, _ := GenerateNLengthBinaryString(38, int(math.Pow(float64(2), float64(38))))

	bs := append(filter, partition...)
	bs = append(bs, companyPrefix...)
	bs = append(bs, itemReference...)
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

	var sgtin96 = []interface{}{
		uint8(48), // SGTIN-96 Header 0011 0000
		p[0], // 8 bits -> 16 bits
		p[1], // 8 bits -> 24 bits
		p[2], // 8 bits -> 32 bits
		p[3], // 8 bits -> 40 bits
		p[4], // 8 bits -> 48 bits
		p[5], // 8 bits -> 56 bits
		p[6], // 8 bits -> 64 bits
		p[7], // 8 bits -> 72 bits
		p[8], // 8 bits -> 80 bits
		p[9], // 8 bits -> 88 bits
		p[10],// 8 bits -> 96 bits
	}

	return Pack(sgtin96), nil
}

// SSCC-96
func GenerateRandomSSCC96() ([]byte, error) {
	filter, _ := GenerateNLengthBinaryString(3, 7)
	partition, pValue := GenerateNLengthBinaryString(3, 6)

	// SGTIN Partition Table
	cpSizes := make([]int, 2)
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

	companyPrefix, _ := GenerateNLengthBinaryString(cpSizes[0], int(math.Pow(float64(10), float64(cpSizes[1]))))
	extension, _ := GenerateNLengthBinaryString(58-cpSizes[0], int(math.Pow(float64(10), float64(17-cpSizes[1]))))

	// 24 '0's
	//reserved := []rune{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	reserved, _ := GenerateNLengthBinaryString(24, 0)

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
		p[0], // 8 bits -> 16 bits
		p[1], // 8 bits -> 24 bits
		p[2], // 8 bits -> 32 bits
		p[3], // 8 bits -> 40 bits
		p[4], // 8 bits -> 48 bits
		p[5], // 8 bits -> 56 bits
		p[6], // 8 bits -> 64 bits
		p[7], // 8 bits -> 72 bits
		p[8], // 8 bits -> 80 bits
		p[9], // 8 bits -> 88 bits
		p[10],// 8 bits -> 96 bits
	}

	return Pack(sscc96), nil
}

// GRAI-96
func GenerateRandomGRAI96() ([]byte, error) {
	filter, _ := GenerateNLengthBinaryString(3, 7)
	partition, pValue := GenerateNLengthBinaryString(3, 6)

	// SGTIN Partition Table
	cpSizes := make([]int, 2)
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

	companyPrefix, _ := GenerateNLengthBinaryString(cpSizes[0], int(math.Pow(float64(10), float64(cpSizes[1]))))
	assetType, _ := GenerateNLengthBinaryString(44-cpSizes[0], int(math.Pow(float64(10), float64(12-cpSizes[1]))))
	serial, _ := GenerateNLengthBinaryString(38, int(math.Pow(float64(2), float64(38))))

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
		p[0], // 8 bits -> 16 bits
		p[1], // 8 bits -> 24 bits
		p[2], // 8 bits -> 32 bits
		p[3], // 8 bits -> 40 bits
		p[4], // 8 bits -> 48 bits
		p[5], // 8 bits -> 56 bits
		p[6], // 8 bits -> 64 bits
		p[7], // 8 bits -> 72 bits
		p[8], // 8 bits -> 80 bits
		p[9], // 8 bits -> 88 bits
		p[10],// 8 bits -> 96 bits
	}

	return Pack(grai96), nil
}

// GIAI-96
func GenerateRandomGIAI96() ([]byte, error) {
	filter, _ := GenerateNLengthBinaryString(3, 7)
	partition, pValue := GenerateNLengthBinaryString(3, 6)

	// SGTIN Partition Table
	cpSizes := make([]int, 2)
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

	companyPrefix, _ := GenerateNLengthBinaryString(cpSizes[0], int(math.Pow(float64(10), float64(cpSizes[1]))))
	indivisualAssetReference, _ := GenerateNLengthBinaryString(82-cpSizes[0], int(math.Pow(float64(10), float64(25-cpSizes[1]))))

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
		uint8(51), // GRAI-96 Header 0011 0011
		p[0], // 8 bits -> 16 bits
		p[1], // 8 bits -> 24 bits
		p[2], // 8 bits -> 32 bits
		p[3], // 8 bits -> 40 bits
		p[4], // 8 bits -> 48 bits
		p[5], // 8 bits -> 56 bits
		p[6], // 8 bits -> 64 bits
		p[7], // 8 bits -> 72 bits
		p[8], // 8 bits -> 80 bits
		p[9], // 8 bits -> 88 bits
		p[10],// 8 bits -> 96 bits
	}

	return Pack(giai96), nil
}

func GenerateRandomEPCData(std string) ([]byte, error) {
	var epcData []byte

	switch std {
		case "SGTIN-96":
			epcData, _ = GenerateRandomSGTIN96()
		case "SSCC-96":
			epcData, _ = GenerateRandomSSCC96()
		case "GRAI-96":
			epcData, _ = GenerateRandomGRAI96()
		case "GIAI-96":
			epcData, _ = GenerateRandomGIAI96()
	}

	return epcData, nil
}

func GenerateEPC() string {
	epcs := []string{"SGTIN-96", "SSCC-96", "GRAI-96", "GIAI-96"}

	if !stringInSlice(strings.ToUpper(*epcScheme), epcs) {
		os.Exit(1)
	}
	fmt.Println(strings.ToUpper(*epcScheme))

	uii, _ := GenerateRandomEPCData(strings.ToUpper(*epcScheme))
	rd, _ := hex.DecodeString(RandomHexRunes(4))

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
		hex.EncodeToString(uii) + "," +
		hex.EncodeToString(rd)
}

func GenerateISO() string {
	var uii []byte
	var pc []byte
	var length uint16
	var epclen uint16

	isos := []uint16{17365, 17363}
	i := randomInt(0, len(isos))

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
	rd, _ := hex.DecodeString(RandomHexRunes(4))

	return hex.EncodeToString(pc) + "," +
		strconv.FormatUint(uint64(length), 10) + "," +
		strconv.FormatUint(uint64(epclen), 10) + "," +
		hex.EncodeToString(uii) + "," +
		hex.EncodeToString(rd)
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
