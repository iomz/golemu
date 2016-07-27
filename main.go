package main

import (
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

	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/iomz/go-llrp"
)

type Tag struct {
	pcBits        string
	length        int64
	epcLengthBits int64
	epc           []byte
	readData      []byte
}

const (
	VERSION     = "0.0.1"
	BUFSIZE     = 512
)

var (
	app                = kingpin.New("gologir", "A mock LLRP-based logical reader for RFID Tags.")
	verbose            = app.Flag("verbose", "Enable verbose mode.").Bool()
	initalMessageID    = app.Flag("initialMessageID", "The initial messageID to start from.").Default("1000").Int()
	initialKeepaliveID = app.Flag("initialKeepaliveID", "The initial keepaliveID to start from.").Default("80000").Int()
	port               = app.Flag("port", "LLRP listening port.").Default("5084").Int()
	ip                 = app.Flag("ip", "LLRP listening address.").Default("127.0.0.1").IP()

	stream = app.Command("stream", "Run as a tag streamer.")
	//sub  = stream.Arg("sub", "Sub option.")

	client = app.Command("client", "Run as a client mode.")

	messageID   = *initalMessageID
	keepaliveID = *initialKeepaliveID
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func BuildTag(record []string) (Tag, error) {
	// If the row is incomplete
	if len(record) != 4 {
		var t Tag
		return t, io.EOF
	}

	pcBits := record[0]
	length, err := strconv.ParseInt(record[1], 10, 16)
	Check(err)
	epcLengthBits, err := strconv.ParseInt(record[2], 10, 16)
	Check(err)
	epc, err := hex.DecodeString(record[3])
	Check(err)
	readData, err := hex.DecodeString("a896")
	Check(err)

	tag := Tag{pcBits, length, epcLengthBits, epc, readData}
	return tag, nil
}

func readTagsFromCSV(csvfile string) []*Tag {
	csv_in, err := ioutil.ReadFile(csvfile)
	Check(err)
	r := csv.NewReader(strings.NewReader(string(csv_in)))

	tags := []*Tag{}
	for {
		record, err := r.Read()
		// If reached at the end
		if err == io.EOF {
			break
		}
		Check(err)

		// Construct a tag read data
		tag, err := BuildTag(record)
		if err != nil {
			continue
		}
		tags = append(tags, &tag)
	}
	return tags
}

func emit(conn net.Conn, tags []*Tag) {
	t := time.NewTicker(1 * time.Second)
	count := 0
	for {
		for _, tag := range tags {
			// Log the tag for debug
			//log.Printf("%+v\n", tag)

			// PeakRSSI
			peakRSSI := llrp.PeakRSSI()

			// AirProtocolTagData
			airProtocolTagData := llrp.C1G2PC(tag.pcBits)

			// EPCData
			epcData := llrp.EPCData(tag.length, tag.epcLengthBits, tag.epc)

			// OpSpecResult
			opSpecResult := llrp.C1G2ReadOpSpecResult(tag.readData)

			// Merge them into TagReportData
			tagReportData :=
				llrp.TagReportData(epcData,
					peakRSSI, airProtocolTagData,
					opSpecResult)

			/*

			   TODO: Here, maybe stack more TRDs to ROAR

			*/

			// Append TagReportData to ROAccessReport
			roAccessReport :=
				llrp.ROAccessReport(tagReportData, messageID)
			messageID += 1

			// Send
			conn.Write(roAccessReport)

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
		log.Println("Error reading:", err.Error())
		log.Println("reqLen: " + string(reqLen))
		conn.Close()
	}

	header := binary.BigEndian.Uint16(buf[:2])
	if header == llrp.H_SetReaderConfig {
		log.Println(">>> SET_READER_CONFIG")
		conn.Write(llrp.SetReaderConfigResponse())
		// Read virtual tags from a csv file
		tags := readTagsFromCSV("tags.csv")
		// Emit LLRP
		go emit(conn, tags)
	} else if header == llrp.H_KeepaliveAck {
		tags := readTagsFromCSV("tags.csv")
		go emit(conn, tags)
	} else {
		log.Printf("Unknown header: %v\n", header)
		fmt.Println("Message: %v", buf)
		return
	}
}

// stream mode
func runStream() {
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
		conn.Write(llrp.ReaderEventNotification(messageID))
		messageID += 1
		time.Sleep(time.Millisecond)

		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
}

// client mode
func runClient() {
	// Establish a connection to the llrp client
	conn, err := net.Dial("tcp", ip.String()+":"+strconv.Itoa(*port))
	Check(err)

	buf := make([]byte, BUFSIZE)
	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err == io.EOF {
			// Close the connection when you're done with it.
			return
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
}

func main() {
	app.Version(VERSION)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case stream.FullCommand():
		runStream()
	case client.FullCommand():
		runClient()
	}
}
