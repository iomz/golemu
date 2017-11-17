// A tool to generate arbitrary UII (aka EPC)
package main

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"os"

	"github.com/iomz/go-llrp/binutil"
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
	pc := binutil.Pack([]interface{}{
		uint8(48), // L4-0=11000(6words=96bits), UMI=0, XI=0
		uint8(0),  // RFU=0
	})

	length := uint16(18)
	epclen := uint16(96)

	return hex.EncodeToString(pc) + "," +
		strconv.FormatUint(uint64(length), 10) + "," +
		strconv.FormatUint(uint64(epclen), 10) + "," +
		hex.EncodeToString(uii) + "\n" +
		binutil.ParseHexStringToBinString(hex.EncodeToString(uii))
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
		binutil.ParseHexStringToBinString(hex.EncodeToString(uii))
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
