package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
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

var messageID = 1000

func check(e error) {
	if e != nil {
		panic(e)
	}
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

func buildROAccessReport(tagReportDataParameter []byte) []byte {
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
	return buf.Bytes()
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

func buildReaderEventNotification() []byte {
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
	return buf.Bytes()
}

func buildLLRPStatusParameter() []byte {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(287), // Rsvd+Type=287
		uint16(8),   // Length
		uint16(0),   // StatusCode=M_Success(0)
		uint16(0),   // ErrorDescriptionByteCount=0
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
}

func buildSetReaderConfigResponse() []byte {
	llrpStatusParameter := buildLLRPStatusParameter()
	setReaderConfigResponseLength :=
		len(llrpStatusParameter) + 10 // Rsvd+Ver+Type+Length+ID+R+Rsvd->80bits=10bytes
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(HEADER_SRCR),                   // Rsvd+Ver+Type=13 (SET_READER_CONFIG_RESPONSE)
		uint32(setReaderConfigResponseLength), // Length
		uint32(0), // ID
		llrpStatusParameter,
	}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		check(err)
	}
	return buf.Bytes()
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
		// Check the connection and disconnect

		for _, tag := range tags {
			fmt.Printf("%+v\n", tag)

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
			roAccessReport :=
				buildROAccessReport(tagReportDataParameter)

			// Send
			conn.Write(roAccessReport)

			// Wait until ACK received
			time.Sleep(time.Millisecond)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, BUFSIZE)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err == io.EOF {
		// Close the connection when you're done with it.
		return
	} else if err != nil {
		fmt.Println("Error reading:", err.Error())
		fmt.Println("reqLen: " + string(reqLen))
		conn.Close()
	}

	header := binary.BigEndian.Uint16(buf[:2])
	if header == HEADER_SRC {
		fmt.Println(">>> SET_READER_CONFIG")
		conn.Write(buildSetReaderConfigResponse())
		// Read virtual tags from a csv file
		tags := readTagsFromCSV("tags.csv")
		// Emit LLRP
		go emit(conn, tags)
	} else {
		fmt.Printf("Unknown header: %v\n", header)
	}
}

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// Send back READER_EVENT_NOTIFICATION
		conn.Write(buildReaderEventNotification())
		time.Sleep(time.Millisecond)

		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
}
