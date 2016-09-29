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

	"github.com/iomz/go-llrp"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Tag struct {
	pcBits        uint16
	length        uint16
	epcLengthBits uint16
	epc           []byte
	readData      []byte
}

func (t Tag) Equal(tt Tag) bool {
	if t.pcBits == tt.pcBits && t.length == tt.length && tt.epcLengthBits == tt.epcLengthBits && bytes.Equal(t.epc, tt.epc) && bytes.Equal(t.readData, tt.readData) {
		return true
	} else {
		return false
	}
}

const (
	BUFSIZE = 512
)

var (
	app                = kingpin.New("gologir", "A mock LLRP-based logical reader for RFID Tags.")
	verbose            = app.Flag("verbose", "Enable verbose mode.").Short('v').Bool()
	initalMessageID    = app.Flag("initialMessageID", "The initial messageID to start from.").Short('m').Default("1000").Int()
	initialKeepaliveID = app.Flag("initialKeepaliveID", "The initial keepaliveID to start from.").Short('k').Default("80000").Int()
	port               = app.Flag("port", "LLRP listening port.").Short('p').Default("5084").Int()
	ip                 = app.Flag("ip", "LLRP listening address.").Short('i').Default("127.0.0.1").IP()

	server = app.Command("server", "Run as a tag stream server.")
	maxTag = server.Flag("maxTag", "The maximum number of TagReportData parameters per ROAccessReport. Pseudo ROReport spec option. 0 for no limit.").Short('t').Default("0").Int()
	file   = server.Flag("file", "The file containing Tag data.").Short('f').Default("tags.csv").String()

	client = app.Command("client", "Run as a client mode.")

	messageID   = *initalMessageID
	keepaliveID = *initialKeepaliveID
	version     = "0.1.0"
)

// Check if error
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Construct Tag struct from Tag info strings
func buildTag(record []string) (Tag, error) {
	// If the row is incomplete
	if len(record) != 4 {
		var t Tag
		return t, io.EOF
	}

	pc64, err := strconv.ParseUint(record[0], 10, 16)
	check(err)
	pc := uint16(pc64)
	len64, err := strconv.ParseUint(record[1], 10, 16)
	check(err)
	len := uint16(len64)
	epclen64, err := strconv.ParseUint(record[2], 10, 16)
	check(err)
	epclen := uint16(epclen64)
	epc, err := hex.DecodeString(record[3])
	check(err)
	readData, err := hex.DecodeString("a896")
	check(err)

	tag := Tag{pc, len, epclen, epc, readData}
	return tag, nil
}

// Read Tag data from the CSV strings and returns a slice of Tag struct pointers
func loadTagsFromCSV(input string) []*Tag {
	r := csv.NewReader(strings.NewReader(input))
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

// Take one Tag struct and build TagReportData parameter payload in []byte
func buildTagReportDataParameter(tag *Tag) []byte {
	// EPCData
	epcd := llrp.EPCData(tag.length, tag.epcLengthBits, tag.epc)

	// PeakRSSI
	prssi := llrp.PeakRSSI()

	// AirProtocolTagData
	aptd := llrp.C1G2PC(tag.pcBits)

	// OpSpecResult
	osr := llrp.C1G2ReadOpSpecResult(tag.readData)

	// Merge them into TagReportData
	trd := llrp.TagReportData(epcd, prssi, aptd, osr)

	return trd
}

// Iterate through the Tags and write ROAccessReport message to the socket
func emit(conn net.Conn, tags []*Tag) {
	var trds []*[]byte
	tagCount := 0
	trdIndex := 0

	// Iterate through tags and divide them into TRD stacks
	for _, tag := range tags {
		tagCount += 1
		// TODO: Need to set ceiling for too large payload?
		if tagCount > *maxTag && *maxTag != 0 {
			trd := buildTagReportDataParameter(tag)
			trds = append(trds, &trd)
			trdIndex += 1
			tagCount = 1
		} else {
			trd := buildTagReportDataParameter(tag)
			if len(trds) == 0 {
				trds = append(trds, &trd)
			} else {
				*(trds[trdIndex]) = append(*(trds[trdIndex]), trd...)
			}
		}
	}

	t := time.NewTicker(1 * time.Second)
	count := 0
	for { // Infinite loop
		for _, trd := range trds {
			// Append TagReportData to ROAccessReport
			roar := llrp.ROAccessReport(*trd, messageID)
			messageID += 1

			// Send
			conn.Write(roar)

			// Wait until ACK received
			time.Sleep(time.Millisecond)
		}
		select {
		case <-t.C:
			count += 1
		}
		if count >= 10 {
			conn.Write(llrp.Keepalive())
			buf := make([]byte, BUFSIZE)
			reqLen, err := conn.Read(buf)
			if err == io.EOF {
				// Close the connection when you're done with it.
				return
			} else if err != nil {
				log.Println("Error reading:", err.Error())
				log.Println("reqLen: " + string(reqLen))
				conn.Close()
			}
			header := binary.BigEndian.Uint16(buf[:2])
			if header != llrp.H_KeepaliveAck {
				log.Printf("Unknown header: %v\n", header)
				return
			}
			count = 0
		}
	} // Infinite loop
}

// Handles incoming requests.
func handleRequest(conn net.Conn, tags []*Tag) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, BUFSIZE)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err == io.EOF {
		// Close the connection when you're done with it.
		return
	} else if err != nil {
		log.Println("Error reading:", err.Error())
		log.Println("reqLen: " + string(reqLen))
		conn.Close()
	}

	header := binary.BigEndian.Uint16(buf[:2])
	if header == llrp.H_SetReaderConfig {
		log.Println(">>> SET_READER_CONFIG")
		conn.Write(llrp.SetReaderConfigResponse())
		// Emit LLRP
		go emit(conn, tags)
	} else if header == llrp.H_KeepaliveAck {
		go emit(conn, tags)
	} else {
		log.Printf("Unknown header: %v\n", header)
		fmt.Println("Message: %v", buf)
		return
	}
}

// server mode
func runServer() int {
	// Read virtual tags from a csv file
	log.Printf("Loading virtual Tags from \"%v\"\n", *file)
	csv_in, err := ioutil.ReadFile(*file)
	check(err)
	tags := loadTagsFromCSV(string(csv_in))

	// Listen for incoming connections.
	l, err := net.Listen("tcp", ip.String()+":"+strconv.Itoa(*port))
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on " + ip.String() + ":" + strconv.Itoa(*port))

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// Send back READER_EVENT_NOTIFICATION
		currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
		conn.Write(llrp.ReaderEventNotification(messageID, currentTime))
		messageID += 1
		time.Sleep(time.Millisecond)

		// Handle connections in a new goroutine.
		go handleRequest(conn, tags)
	}

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
	return 0
}

// client mode
func runClient() int {
	// Establish a connection to the llrp client
	conn, err := net.Dial("tcp", ip.String()+":"+strconv.Itoa(*port))
	check(err)

	buf := make([]byte, BUFSIZE)
	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err == io.EOF {
			// Close the connection when you're done with it.
			return 0
		} else if err != nil {
			fmt.Println("Error reading:", err.Error())
			fmt.Println("reqLen: " + string(reqLen))
			conn.Close()
			break
		}

		header := binary.BigEndian.Uint16(buf[:2])
		if header == llrp.H_ReaderEventNotification {
			fmt.Println(">>> READER_EVENT_NOTIFICATION")
			conn.Write(llrp.SetReaderConfig(messageID))
		} else if header == llrp.H_SetReaderConfigResponse {
			fmt.Println(">>> SET_READER_CONFIG_RESPONSE")
		} else if header == llrp.H_ROAccessReport {
			fmt.Println(">>> RO_ACCESS_REPORT")
			fmt.Printf("Packet size: %v\n", reqLen)
			fmt.Printf("% x\n", buf[:reqLen])
		} else {
			fmt.Printf("Unknown header: %v\n", header)
		}
	}
	return 0
}

func main() {
	app.Version(version)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case server.FullCommand():
		os.Exit(runServer())
	case client.FullCommand():
		os.Exit(runClient())
	}
}
