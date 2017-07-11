// A tool to generate arbitrary UII (aka EPC)
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// kingpin app
	app = kingpin.New("uiigen", "A tool to generate an arbitrary UII (aka EPC).")

	// kingpin generate EPC mode
	epc = app.Command("epc", "Generate an EPC.")
	// EPC scheme
	epcScheme = epc.Flag("scheme", "Scheme for EPC UII.").Short('s').Default("sgtin-96").String()

	// kingpin generate ISO UII mode
	iso = app.Command("iso", "Generate an ISO UII.")

	//hexRunes to hold hex chars
	hexRunes = []rune("abcdef0123456789")
)

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

func GenerateEPC() string {
	var uii []byte
	var pc []byte
	var length uint16
	var epclen uint16

	epcs := []string{"SGTIN-96", "SSCC", "GRAI"}
	i := randomInt(0, len(isos))


	uii, _ := hex.DecodeString("302db319a00000" + RandomHexRunes(10))
	rd, _ := hex.DecodeString(RandomHexRunes(4))

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
