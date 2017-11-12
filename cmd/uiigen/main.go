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

// CheckIfStringInSlice checks if string exists in a string slice
// TODO: fix the way it is, it should be smarter
func CheckIfStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// GenerateNLengthHexString returns random hex rune for n length
func GenerateNLengthHexString(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range b {
		b[i] = hexRunes[rand.Intn(len(hexRunes))]
	}
	return string(b)
}

// GenerateNLengthRandomBinRuneSlice returns n-length random binary string
// max == 0 for no cap limit
func GenerateNLengthRandomBinRuneSlice(n int, max uint) ([]rune, uint) {
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
		binstr, sum = GenerateNLengthRandomBinRuneSlice(n, max)
	}

	return binstr, sum
}

// GenerateNLengthZeroPaddingRuneSlice returns n-length zero padding string
func GenerateNLengthZeroPaddingRuneSlice(n int) []rune {
	binstr := make([]rune, n)

	for i := 0; i < n; i++ {
		binstr[i] = '0'
	}

	return binstr
}

// GenerateRandomInt return random int value with min-max
func GenerateRandomInt(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

// MakeEPC generates EPC code
func MakeEPC() string {
	epcs := []string{"SGTIN-96", "SSCC-96", "GRAI-96", "GIAI-96"}

	if !CheckIfStringInSlice(strings.ToUpper(*epcScheme), epcs) {
		os.Exit(1)
	}

	var uii []byte
	switch strings.ToUpper(*epcScheme) {
	case "SGTIN-96":
		uii, _ = MakeRuneSliceOfSGTIN96(*epcCompanyPrefix, *epcFilterValue, *epcItemReference, *epcSerial)
	case "SSCC-96":
		uii, _ = MakeRuneSliceOfSSCC96(*epcCompanyPrefix, *epcFilterValue)
	case "GRAI-96":
		uii, _ = MakeRuneSliceOfGRAI96(*epcCompanyPrefix, *epcFilterValue)
	case "GIAI-96":
		uii, _ = MakeRuneSliceOfGIAI96(*epcCompanyPrefix, *epcFilterValue)
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
		ParseHexStringToBinString(hex.EncodeToString(uii))
}

// MakeISO returns ISO code
func MakeISO() string {
	var uii []byte
	var pc []byte
	var length uint16
	var epclen uint16

	isos := []string{"17365", "17363"}

	if !CheckIfStringInSlice(*isoScheme, isos) {
		os.Exit(1)
	}

	switch *isoScheme {
	case "17367":
		uii = MakeRuneSliceOfISO17367()
	}
	/*
		if isos[i] == 17365 {
			// ISO 17365
			uii, _ = hex.DecodeString("c4a301c70d36cb32920b1d" + GenerateNLengthHexString(10))
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
			uii, _ = hex.DecodeString("dc20420c4c" + GenerateNLengthHexString(10))
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
		ParseHexStringToBinString(hex.EncodeToString(uii))
}

// Pack the data into (partial) LLRP packet payload.
func Pack(data []interface{}) []byte {
	buf := new(bytes.Buffer)
	for _, v := range data {
		binary.Write(buf, binary.BigEndian, v)
	}
	return buf.Bytes()
}

// ParseBigIntToBinString makes binary string from hex string
func ParseBigIntToBinString(cp *big.Int) string {
	binStr := fmt.Sprintf("%b", cp)
	return binStr
}

// ParseBinRuneSliceToUint8Slice returns uint8 slice from binary string
// Precondition: len(bs) % 8 == 0
func ParseBinRuneSliceToUint8Slice(bs []rune) ([]uint8, error) {
	if len(bs)%8 != 0 {
		return nil, errors.New("non-8 bit length binary string passed to ParseBinRuneSliceToUint8Slice")
	} else if len(bs) < 8 {
		return nil, errors.New("binary string length less than 8 given to ParseBinRuneSliceToUint8Slice")
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

// ParseDecimalStringToBinRuneSlice convert serial to binary rune slice
func ParseDecimalStringToBinRuneSlice(s string) []rune {
	n, _ := strconv.ParseInt(s, 10, 64)
	binStr := ParseBigIntToBinString(big.NewInt(n))
	return []rune(binStr)
}

// ParseHexStringToBinString converts hex string to binary string
func ParseHexStringToBinString(s string) (binString string) {
	for _, c := range s {
		n, _ := strconv.ParseInt(string(c), 16, 32)
		binString = fmt.Sprintf("%s%.4b", binString, n)
	}
	return
}

func main() {
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch parse {
	case epc.FullCommand():
		fmt.Println(MakeEPC())
	case iso.FullCommand():
		fmt.Println(MakeISO())
	}
}
