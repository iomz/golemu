// A simple LLRP-based logical reader mock for RFID Tags using go-llrp
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/iomz/go-llrp"
	//"github.com/iomz/go-llrp/binutil"
	"github.com/juju/loggo"
	"golang.org/x/net/websocket"
	"gopkg.in/alecthomas/kingpin.v2"
)

// ManagementAction is a type for TagManager
type ManagementAction int

const (
	// RetrieveTags is a const for retrieving tags
	RetrieveTags ManagementAction = iota
	// AddTags is a const for adding tags
	AddTags
	// DeleteTags is a const for deleting tags
	DeleteTags
)

// TagManager is a struct for tag management channel
type TagManager struct {
	action ManagementAction
	tags   []*Tag
}

var (
	// Current Version
	version = "0.1.0"

	// kingpin app
	app = kingpin.New("golemu", "A mock LLRP-based logical reader emulator for RFID Tags.")
	// kingpin debug mode flag
	debug = app.Flag("debug", "Enable debug mode.").Short('v').Default("false").Bool()
	// kingpin initial MessageID
	initialMessageID = app.Flag("initialMessageID", "The initial messageID to start from.").Default("1000").Int()
	// kingpin initial KeepaliveID
	initialKeepaliveID = app.Flag("initialKeepaliveID", "The initial keepaliveID to start from.").Default("80000").Int()
	// kingpin LLRP listening IP address
	ip = app.Flag("ip", "LLRP listening address.").Short('a').Default("0.0.0.0").IP()
	// kingpin keepalive interval
	keepaliveInterval = app.Flag("keepalive", "LLRP Keepalive interval.").Short('k').Default("0").Int()
	// kingpin LLRP listening port
	port = app.Flag("port", "LLRP listening port.").Short('p').Default("5084").Int()

	// kingpin server command
	server = app.Command("server", "Run as a tag stream server.")
	// kingpin report interval
	reportInterval = server.Flag("reportInterval", "The interval of ROAccessReport in ms. Pseudo ROReport spec option.").Short('i').Default("1000").Int()
	// kingpin web port
	webPort = server.Flag("webPort", "Port listening for web access.").Short('w').Default("3000").Int()
	// kingpin Protocol Data Unit for LLRP
	pdu = server.Flag("pdu", "The maximum size of LLRP PDU.").Short('m').Default("1500").Int()
	// kingpin tag list file
	file = server.Flag("file", "The file containing Tag data.").Short('f').Default("tags.csv").String()

	// kingpin client command
	client = app.Command("client", "Run as a client mode.")

	// loggo
	logger = loggo.GetLogger("")

	// LLRPConn flag
	isLLRPConnAlive = false
	// Current messageID
	messageID = uint32(*initialMessageID)
	// Current KeepaliveID
	keepaliveID = *initialKeepaliveID
	// Current activeClients
	activeClients = make(map[WebsockConn]int) // map containing clients
	// Tag management channel
	tagManager = make(chan *TagManager)
	// notify tag update channel
	notify = make(chan bool)
	// update TagReportDataStack when tag is updated
	tagUpdated = make(chan []*Tag)
)

// Check if error
func check(e error) {
	if e != nil {
		panic(e.Error())
	}
}

// Time
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Debugf("%s took %s", name, elapsed)
}

// decapsulate the ROAccessReport and extract IDs
func decapsulateROAccessReport(roarLength uint32, buf []byte) int {
	count := 0
	defer timeTrack(time.Now(), fmt.Sprintf("unpacking %v bytes", len(buf)))
	trds := buf[4 : roarLength-6] // TRD stack
	trdLength := uint16(0)        // First TRD size
	offset := uint32(0)           // the start of TRD
	//logger.Debugf("len(trds): %v\n", len(trds))
	//for trdLength != 0 && int(offset) != len(trds) {
	for {
		if uint32(10+offset) < roarLength {
			trdLength = binary.BigEndian.Uint16(trds[offset+2 : offset+4])
		} else {
			break
		}
		var id, pc []byte
		if trds[offset+4] == 141 { // EPC-96
			id = trds[offset+5 : offset+17]
			if trds[offset+17] == 140 { // C1G2-PC parameter
				pc = trds[offset+18 : offset+20]
			}
			count++
			//logger.Debugf("EPC: %v, (%x)\n", id, pc)
		} else if binary.BigEndian.Uint16(trds[offset+4:offset+6]) == 241 { // EPCData
			epcDataLength := binary.BigEndian.Uint16(trds[offset+6 : offset+8])  // length
			epcLengthBits := binary.BigEndian.Uint16(trds[offset+8 : offset+10]) // EPCLengthBits
			epcLengthBytes := uint32(epcLengthBits / 8)
			/*
				// ID length in byte = Length - (6 + 10 + 16 + 16)/8
				//id = trds[offset+6 : offset+epcDataSize-6]
				// trim the last 1 byte if it's not a multiple of a word
				//id = id[0 : epcLengthBits/8]
			*/
			id = trds[offset+10 : offset+10+epcLengthBytes]
			if 4+epcDataLength < trdLength && trds[offset+10+epcLengthBytes] == 140 { // C1G2-PC parameter
				pc = trds[offset+10+epcLengthBytes+1 : offset+10+epcLengthBytes+3]
			}
			_ = id
			_ = pc
			count++
			//logger.Debugf("EPC: %v, (%x)\n", id, pc)
		}
		offset += uint32(trdLength) // move the offset at the end of this TRD
		//logger.Debugf("offset: %v, roarLength: %v\n", offset, roarLength)
		//logger.Debugf("trdLength: %v, len(trds): %v\n", trdLength, len(trds))
	}
	return count
}

// Iterate through the Tags and write ROAccessReport message to the socket
func sendROAccessReport(conn net.Conn, trds *TagReportDataStack) error {
	perms := rand.Perm(len(trds.Stack))
	//buf := make([]byte, 512)
	for _, i := range perms {
		trd := trds.Stack[i]
		// Append TagReportData to ROAccessReport
		roar := llrp.ROAccessReport(trd.Parameter, messageID)
		atomic.AddUint32(&messageID, 1)
		runtime.Gosched()

		// Send
		_, err := conn.Write(roar)
		if err != nil {
			return err
		}
		//time.Sleep(time.Millisecond)
	}

	return nil
}

// Handles incoming requests.
func handleRequest(conn net.Conn, tags []*Tag) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, *pdu)
	trds := buildTagReportDataStack(tags)

	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err == io.EOF {
			// Close the connection when you're done with it.
			logger.Infof("The client is disconnected, closing LLRP connection")
			conn.Close()
			return
		} else if err != nil {
			logger.Errorf(err.Error())
			logger.Infof("Closing LLRP connection")
			conn.Close()
			return
		}

		// Respond according to the LLRP packet header
		header := binary.BigEndian.Uint16(buf[:2])
		if header == llrp.SetReaderConfigHeader || header == llrp.KeepaliveAckHeader {
			if header == llrp.SetReaderConfigHeader {
				// SRC received, start ROAR
				logger.Infof(">>> SET_READER_CONFIG")
				conn.Write(llrp.SetReaderConfigResponse())
			} else if header == llrp.KeepaliveAckHeader {
				// KA receieved, continue ROAR
				logger.Infof(">>> KEEP_ALIVE_ACK")
			}

			// Tick ROAR and Keepalive interval
			roarTicker := time.NewTicker(time.Duration(*reportInterval) * time.Millisecond)
			keepaliveTicker := &time.Ticker{}
			if *keepaliveInterval != 0 {
				keepaliveTicker = time.NewTicker(time.Duration(*keepaliveInterval) * time.Second)
			}
			go func() {
				for { // Infinite loop
					isLLRPConnAlive = true
					logger.Debugf("[LLRP handler select]: %v", trds)
					select {
					// ROAccessReport interval tick
					case <-roarTicker.C:
						logger.Tracef("### roarTicker.C")
						logger.Infof("<<< RO_ACCESS_REPORT (# reports: %v, # total tags: %v)", len(trds.Stack), trds.TotalTagCounts())
						err := sendROAccessReport(conn, trds)
						if err != nil {
							logger.Errorf(err.Error())
							isLLRPConnAlive = false
						}
					// Keepalive interval tick
					case <-keepaliveTicker.C:
						logger.Tracef("### keepaliveTicker.C")
						logger.Infof("<<< KEEP_ALIVE")
						conn.Write(llrp.Keepalive())
						isLLRPConnAlive = false
					// When the tag queue is updated
					case tags := <-tagUpdated:
						logger.Tracef("### TagUpdated")
						trds = buildTagReportDataStack(tags)
					}
					if !isLLRPConnAlive {
						roarTicker.Stop()
						if *keepaliveInterval != 0 {
							keepaliveTicker.Stop()
						}
						break
					}
				}
			}()
		} else {
			// Unknown LLRP packet received, reset the connection
			logger.Warningf("Unknown header: %v, reqlen: %v", header, reqLen)
			logger.Warningf("Message: %v", buf)
			return
		}
	}
}

// APIPostTag redirects the tag addition request
func APIPostTag(c *gin.Context) {
	var json []TagInString
	c.BindWith(&json, binding.JSON)
	if res := ReqAddTag("add", json); res == "error" {
		c.String(http.StatusAlreadyReported, "The tag already exists!\n")
	} else {
		c.String(http.StatusAccepted, "Post requested!\n")
	}
}

// APIDeleteTag redirects the tag deletion request
func APIDeleteTag(c *gin.Context) {
	var json []TagInString
	c.BindWith(&json, binding.JSON)
	if res := ReqDeleteTag("delete", json); res == "error" {
		c.String(http.StatusNoContent, "The tag doesn't exist!\n")
	} else {
		c.String(http.StatusAccepted, "Delete requested!\n")
	}
}

// server mode
func runServer() int {
	// Read virtual tags from a csv file
	logger.Infof("Loading virtual Tags from \"%v\"", *file)

	if _, err := os.Stat(*file); os.IsNotExist(err) {
		_, err := os.Create(*file)
		check(err)
		logger.Infof("%v created.", *file)
	}

	// Prepare the tags
	/*
		tags := new([]*Tag)
		if _, err := os.Stat("tags.gob"); os.IsNotExist(err) {
			tags = loadTagsFromCSV(*file)
			binutil.Save("tags.gob", tags)
		} else {
			if err := binutil.Load("tags.gob", tags); err != nil {
				panic(err)
			}
		}
	*/
	tags := loadTagsFromCSV(*file)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", ip.String()+":"+strconv.Itoa(*port))
	check(err)

	// Close the listener when the application closes.
	defer l.Close()
	logger.Infof("Listening on %v:%v", ip, *port)

	// Channel for communicating virtual tag updates and signals
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Handle websocket and static file hosting with gin
	go func() {
		r := gin.Default()
		r.Use(static.Serve("/", static.LocalFile("./public", true)))
		r.GET("/ws", func(c *gin.Context) {
			handler := websocket.Handler(SockServer)
			handler.ServeHTTP(c.Writer, c.Request)
		})
		v1 := r.Group("api/v1")
		v1.POST("/tags", APIPostTag)
		v1.DELETE("/tags", APIDeleteTag)
		r.Run(":" + strconv.Itoa(*webPort))
	}()

	go func() {
		for {
			select {
			case cmd := <-tagManager:
				// Tag management
				res := []*Tag{}
				switch cmd.action {
				case AddTags:
					for _, t := range cmd.tags {
						if i := getIndexOfTag(*tags, t); i < 0 {
							*tags = append(*tags, t)
							res = append(res, t)
							// Write to file
							//writeTagsToCSV(*tags, *file)
							if isLLRPConnAlive {
								tagUpdated <- *tags
							}
						}
					}
				case DeleteTags:
					for _, t := range cmd.tags {
						if i := getIndexOfTag(*tags, t); i >= 0 {
							*tags = append((*tags)[:i], (*tags)[i+1:]...)
							res = append(res, t)
							// Write to file
							//writeTagsToCSV(tags, *file)
							if isLLRPConnAlive {
								tagUpdated <- *tags
							}
						}
					}
				case RetrieveTags:
					res = *tags
				}
				cmd.tags = res
				tagManager <- cmd
			case signal := <-signals:
				// Handle SIGINT and SIGTERM.
				logger.Infof("%v", signal)
				os.Exit(0)
			}
		}
	}()

	// Handle LLRP connection
	for {
		// Accept an incoming connection.
		logger.Infof("LLRP connection initiated")
		conn, err := l.Accept()
		if err != nil {
			logger.Errorf(err.Error())
			os.Exit(2)
		}

		// Send back READER_EVENT_NOTIFICATION
		currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
		conn.Write(llrp.ReaderEventNotification(messageID, currentTime))
		logger.Infof("<<< READER_EVENT_NOTIFICATION")
		atomic.AddUint32(&messageID, 1)
		runtime.Gosched()
		time.Sleep(time.Millisecond)

		// Handle connections in a new goroutine.
		go handleRequest(conn, *tags)
	}
}

// client mode
func runClient() int {
	// Establish a connection to the llrp client
	conn, err := net.Dial("tcp", ip.String()+":"+strconv.Itoa(*port))
	check(err)

	header := make([]byte, 2)
	length := make([]byte, 4)
	count := 0
	for {
		_, err = io.ReadFull(conn, header)
		//length := binary.BigEndian.Uint32(prefix)

		h := binary.BigEndian.Uint16(header)
		if h == llrp.ReaderEventNotificationHeader {
			_, err = io.ReadFull(conn, length)
			message := make([]byte, binary.BigEndian.Uint32(length)-6)
			_, err = io.ReadFull(conn, message)
			logger.Infof(">>> READER_EVENT_NOTIFICATION")
			conn.Write(llrp.SetReaderConfig(messageID))
		} else if h == llrp.KeepaliveHeader {
			_, err = io.ReadFull(conn, length)
			message := make([]byte, binary.BigEndian.Uint32(length)-6)
			_, err = io.ReadFull(conn, message)
			logger.Infof(">>> KEEP_ALIVE")
			conn.Write(llrp.KeepaliveAck())
		} else if h == llrp.SetReaderConfigResponseHeader {
			_, err = io.ReadFull(conn, length)
			message := make([]byte, binary.BigEndian.Uint32(length)-6)
			_, err = io.ReadFull(conn, message)
			logger.Infof(">>> SET_READER_CONFIG_RESPONSE")
		} else if h == llrp.ROAccessReportHeader {
			_, err = io.ReadFull(conn, length)
			l := binary.BigEndian.Uint32(length)
			message := make([]byte, l-6)
			_, err = io.ReadFull(conn, message)
			logger.Infof(">>> RO_ACCESS_REPORT")
			count += decapsulateROAccessReport(l, message)
			logger.Debugf("%v", count)
		} else {
			logger.Warningf("Unknown header: %v", h)
			return 1
		}
	}
	return 0
}

func main() {
	app.Version(version)
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		loggo.ConfigureLoggers("TRACE")
		gin.SetMode(gin.DebugMode)
	} else {
		loggo.ConfigureLoggers("INFO")
		gin.SetMode(gin.ReleaseMode)
	}

	switch parse {
	case server.FullCommand():
		os.Exit(runServer())
	case client.FullCommand():
		os.Exit(runClient())
	}
}
