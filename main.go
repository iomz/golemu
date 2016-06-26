package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Tag struct {
	pcBits        string
	length        int64
	epcLengthBits int64
	epc           []byte
	readData      []byte
}

const (
	CONN_HOST   = "0.0.0.0"
	CONN_PORT   = "5084"
	CONN_TYPE   = "tcp"
	BUFSIZE     = 512
	HEADER_ROAR = 1085
	HEADER_REN  = 1087
	HEADER_SRC  = 1027
	HEADER_SRCR = 1037
)

var llrpHost string
var llrpPort int
var messageID = 1000

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	const (
		defaultHost = "127.0.0.1"
		hostUsage   = "llrp client hostname"
		defaultPort = 5084
		portUsage   = "port used for llrp connection"
	)
	flag.StringVar(&llrpHost, "host", defaultHost, hostUsage)
	flag.StringVar(&llrpHost, "h",
		defaultHost, hostUsage+" (shorthand)")
	flag.IntVar(&llrpPort, "port", defaultPort, portUsage)
	flag.IntVar(&llrpPort, "p",
		defaultPort, portUsage+" (shorthand)")
}

func buildPeakRSSIParameter() []byte {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint8(134), // 1+uint7(Type=6)
		uint8(203), // PeakRSSI
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildC1G2PCParameter(hexpc string) []byte {
	buf := new(bytes.Buffer)
	intpc, _ := strconv.ParseInt(hexpc, 10, 32)
	var data = []interface{}{
		uint8(140),    // 1+uint7(Type=12)
		uint16(intpc), // PC bits
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildEPCDataParameter(length int64, epcLengthBits int64, epc []byte) []byte {
	var data []interface{}
	if epcLengthBits == 96 {
		data = []interface{}{
			uint8(141), // 1+uint7(Type=13)
			epc,        // 96-bit EPCData string
		}
	} else {
		data = []interface{}{
			uint16(241),           // uint8(0)+uint8(Type=241)
			uint16(length),        // Length
			uint16(epcLengthBits), // EPCLengthBits
			epc, // EPCData string
		}
	}
	buf := new(bytes.Buffer)
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildC1G2ReadOpSpecResultParameter(readData []byte) []byte {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(349), // Rsvd+Type=
		uint16(11),  // Length
		uint8(0),    // Result
		uint16(9),   // OpSpecID
		uint16(1),   // ReadDataWordCount
		readData,    // ReadData
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildTagReportDataParameter(epcDataParameter []byte,
	peakRSSIParameter []byte,
	airProtocolTagDataParameter []byte,
	opSpecResultParameter []byte) []byte {
	tagReportDataLength := len(epcDataParameter) +
		len(peakRSSIParameter) + len(airProtocolTagDataParameter) +
		len(opSpecResultParameter) + 4 // Rsvd+Type+length->32bits=4bytes
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(240),                 // Rsvd+Type=240 (TagReportData parameter)
		uint16(tagReportDataLength), // Length
		epcDataParameter,
		peakRSSIParameter,
		airProtocolTagDataParameter,
		opSpecResultParameter,
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func bufferROAccessReport(tagReportDataParameter []byte) *bytes.Buffer {
	roAccessReportLength :=
		len(tagReportDataParameter) + 10 // Rsvd+Ver+Type+Length+ID->80bits=10bytes
	messageID += 1
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(HEADER_ROAR),          // Rsvd+Ver+Type=61 (RO_ACCESS_REPORT)
		uint32(roAccessReportLength), // Message length
		uint32(messageID),            // Message ID
		tagReportDataParameter,
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf
}

func buildUTCTimeStampParameter() []byte {
	buf := new(bytes.Buffer)
	currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
	var data = []interface{}{
		uint16(128), // Rsvd+Type=128
		uint16(12),  // Length
		currentTime, // Microseconds
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildConnectionAttemptEventParameter() []byte {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(256), // Rsvd+Type=256
		uint16(6),   // Length
		uint16(0),   // Status(Success=0)
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildReaderEventNotificationDataParameter() []byte {
	utcTimeStampParameter := buildUTCTimeStampParameter()
	connectionAttemptEventParameter := buildConnectionAttemptEventParameter()
	readerEventNotificationDataLength := len(utcTimeStampParameter) +
		len(connectionAttemptEventParameter) + 4 // Rsvd+Type+length=32bits=4bytes
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(246),                               // Rsvd+Type=246 (ReaderEventNotificationData parameter)
		uint16(readerEventNotificationDataLength), // Length
		utcTimeStampParameter,
		connectionAttemptEventParameter,
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func bufferReaderEventNotification() *bytes.Buffer {
	readerEventNotificationDataParameter := buildReaderEventNotificationDataParameter()
	readerEventNotificationLength :=
		len(readerEventNotificationDataParameter) + 10 // Rsvd+Ver+Type+Length+ID->80bits=10bytes
	messageID += 1
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(HEADER_REN),                    // Rsvd+Ver+Type=63 (READER_EVENT_NOTIFICATION)
		uint32(readerEventNotificationLength), // Length
		uint32(messageID),                     // ID
		readerEventNotificationDataParameter,
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf
}

func buildKeepaliveSpecParameter() []byte {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(220),   // Rsvd+Type=220
		uint16(9),     // Length
		uint8(1),      // KeepaliveTriggerType=Periodic(1)
		uint32(10000), // TimeInterval=10000
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func bufferSetReaderConfig() *bytes.Buffer {
	keepaliveSpecParameter := buildKeepaliveSpecParameter()
	setReaderConfigLength :=
		len(keepaliveSpecParameter) + 11 // Rsvd+Ver+Type+Length+ID+R+Rsvd->88bits=11bytes
	messageID += 1
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(HEADER_SRC),            // Rsvd+Ver+Type=3 (SET_READER_CONFIG)
		uint32(setReaderConfigLength), // Length
		uint32(messageID),             // ID
		uint8(0),                      // RestoreFactorySetting(no=0)+Rsvd
		keepaliveSpecParameter,
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf
}

func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}

func buildTag(record []string) (Tag, error) {
	// If the row is incomplete
	if len(record) != 4 {
		var t Tag
		return t, io.EOF
	}

	pcBits := record[0]
	length, err := strconv.ParseInt(record[1], 10, 16)
	check(err)
	epcLengthBits, err := strconv.ParseInt(record[2], 10, 16)
	check(err)
	epc, err := hex.DecodeString(record[3])
	check(err)
	readData, err := hex.DecodeString("a896")
	check(err)

	tag := Tag{pcBits, length, epcLengthBits, epc, readData}
	fmt.Printf("%+v\n", tag)
	return tag, nil
}

func readTagsFromCSV(csvfile string) []*Tag {
	csv_in, err := ioutil.ReadFile(csvfile)
	check(err)
	r := csv.NewReader(strings.NewReader(string(csv_in)))

	tags := []*Tag{}
	for {
		record, err := r.Read()
		// If reached at the end
		if err == io.EOF {
			break
		}
		check(err)

		// Construct a tag read data
		tag, err := buildTag(record)
		if err != nil {
			continue
		}
		tags = append(tags, &tag)
	}
	return tags
}

func emit(conn net.Conn, tags []*Tag) {
	for {
		for _, tag := range tags {
			// PeakRSSIParameter
			peakRSSIParameter :=
				buildPeakRSSIParameter()

			// AirProtocolTagDataParameter
			airProtocolTagDataParameter :=
				buildC1G2PCParameter(tag.pcBits)

			// EPCDataParameter
			epcDataParameter :=
				buildEPCDataParameter(tag.length, tag.epcLengthBits, tag.epc)

			// OpSpecResultParameter
			opSpecResultParameter :=
				buildC1G2ReadOpSpecResultParameter(tag.readData)

			// Merge them into TagReportData
			tagReportDataParameter :=
				buildTagReportDataParameter(epcDataParameter,
					peakRSSIParameter, airProtocolTagDataParameter,
					opSpecResultParameter)

			/*

				TODO: Here, maybe stack more TRDs to ROAR

			*/

			// Append TagReportData to ROAccessReport
			roAccessReportBuffer :=
				bufferROAccessReport(tagReportDataParameter)

			// Send
			fmt.Fprint(conn, roAccessReportBuffer)
			fmt.Printf("%v\n", roAccessReportBuffer.Len())
			fmt.Printf("% x\n", roAccessReportBuffer.Bytes())

			// Wait until ACK received
			time.Sleep(time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
		break
	}
}

func main() {
	flag.Parse()

	// Establish a connection to the llrp client
	conn, err := net.Dial("tcp",
		llrpHost+":"+strconv.Itoa(llrpPort))
	check(err)

	fmt.Fprint(conn, bufferReaderEventNotification())
	time.Sleep(time.Millisecond)
	fmt.Fprint(conn, bufferSetReaderConfig())

	// Read virtual tags from a csv file
	tags := readTagsFromCSV("tags.csv")

	// Emit LLRP
	go emit(conn, tags)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	conn.Close()
}
