package main

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"net"
	//"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/iomz/go-llrp"
	"golang.org/x/net/websocket"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	// BUFSIZE is a general size for a buffer
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

	messageID     = uint32(*initalMessageID)
	keepaliveID   = *initialKeepaliveID
	version       = "0.1.0"
	activeClients = make(map[WebsockConn]int) // map containing clients
	adds          = make(chan *addOp)
	deletes       = make(chan *deleteOp)
	notify        = make(chan bool)
)

func init() {
}

// Iterate through the Tags and write ROAccessReport message to the socket
func sendROAccessReport(conn net.Conn, trds []*[]byte) error {
	for _, trd := range trds {
		// Append TagReportData to ROAccessReport
		roar := llrp.ROAccessReport(*trd, messageID)
		atomic.AddUint32(&messageID, 1)
		runtime.Gosched()

		// Send
		_, err := conn.Write(roar)
		if err != nil {
			return err
		}

		// Wait until ACK received
		time.Sleep(time.Millisecond)
	}

	return nil
}

// Handles incoming requests.
func handleRequest(conn net.Conn, tags []*Tag, tagUpdated chan []*Tag) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, BUFSIZE)

	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err == io.EOF {
			// Close the connection when you're done with it.
			log.Printf("Closing LLRP connection")
			conn.Close()
			return
		} else if err != nil {
			log.Println("Error:", err.Error())
			log.Printf("reqLen = %v\n", reqLen)
			log.Printf("Closing LLRP connection")
			conn.Close()
			return
		}

		// Respond according to the LLRP packet header
		header := binary.BigEndian.Uint16(buf[:2])
		if header == llrp.SetReaderConfigHeader || header == llrp.KeepaliveAckHeader {
			if header == llrp.SetReaderConfigHeader {
				// SRC received, start ROAR
				log.Println(">>> SET_READER_CONFIG")
				conn.Write(llrp.SetReaderConfigResponse())
			} else if header == llrp.KeepaliveAckHeader {
				// KA receieved, continue ROAR
				log.Println(">>> KeepaliveAck")
			}
			trds := buildTagReportDataStack(tags)
			// TODO: ROAR and Keepalive interval
			roarTicker := time.NewTicker(1 * time.Second)
			KeepaliveTicker := time.NewTicker(10 * time.Second)
			for { // Infinite loop
				isAlive := true
				select {
				// ROAccessReport interval tick
				case <-roarTicker.C:
					log.Println("<<< ROAccessReport")
					err := sendROAccessReport(conn, trds)
					if err != nil {
						log.Println("Error:", err.Error())
						isAlive = false
					}
				// Keepalive interval tick
				case <-KeepaliveTicker.C:
					log.Println("<<< Keepalive")
					conn.Write(llrp.Keepalive())
					isAlive = false
				// When the tag queue is updated
				// TODO: Fix to execute without LLRP conection
				case tags := <-tagUpdated:
					trds = buildTagReportDataStack(tags)
					log.Println(trds)
				}
				if !isAlive {
					roarTicker.Stop()
					KeepaliveTicker.Stop()
					break
				}
			}
		} else {
			// Unknown LLRP packet received, reset the connection
			log.Printf("Unknown header: %v\n", header)
			log.Printf("Message: %v\n", buf)
			return
		}
	}
}

// server mode
func runServer() int {
	// Read virtual tags from a csv file
	log.Printf("Loading virtual Tags from \"%v\"\n", *file)
	csvIn, err := ioutil.ReadFile(*file)
	check(err)
	tags := loadTagsFromCSV(string(csvIn))

	// Listen for incoming connections.
	l, err := net.Listen("tcp", ip.String()+":"+strconv.Itoa(*port))
	check(err)

	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on " + ip.String() + ":" + strconv.Itoa(*port))

	// Channel for communicating virtual tag updates
	tagUpdated := make(chan []*Tag)

	// Handle websocket and static file hosting with gin
	go func() {
		r := gin.Default()
		r.Use(static.Serve("/", static.LocalFile("./public", true)))
		r.GET("/ws", func(c *gin.Context) {
			handler := websocket.Handler(SockServer)
			handler.ServeHTTP(c.Writer, c.Request)
		})
		r.Run(":8080")
	}()

	// Tag management
	go func() {
		for {
			select {
			case add := <-adds:
				if getIndexOfTag(tags, add.tag) < 0 {
					tags = append(tags, add.tag)
					writeTagsToCSV(tags, *file)
					add.resp <- true
					tagUpdated <- tags
				} else {
					add.resp <- false
				}
			case delete := <-deletes:
				indexToDelete := getIndexOfTag(tags, delete.tag)
				if indexToDelete >= 0 {
					tags = append(tags[:indexToDelete], tags[indexToDelete+1:]...)
					writeTagsToCSV(tags, *file)
					delete.resp <- true
					tagUpdated <- tags
				} else {
					delete.resp <- false
				}
			}
		}
	}()

	// Handle LLRP connection
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error:", err.Error())
				os.Exit(2)
			}

			// Send back READER_EVENT_NOTIFICATION
			currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
			conn.Write(llrp.ReaderEventNotification(messageID, currentTime))
			atomic.AddUint32(&messageID, 1)
			runtime.Gosched()
			time.Sleep(time.Millisecond)

			// Handle connections in a new goroutine.
			go handleRequest(conn, tags, tagUpdated)
		}
	}()

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
			log.Println("Error:", err.Error())
			log.Printf("reqLen = %v\n", reqLen)
			conn.Close()
			break
		}

		header := binary.BigEndian.Uint16(buf[:2])
		if header == llrp.ReaderEventNotificationHeader {
			log.Println(">>> READER_EVENT_NOTIFICATION")
			conn.Write(llrp.SetReaderConfig(messageID))
		} else if header == llrp.SetReaderConfigResponseHeader {
			log.Println(">>> SET_READER_CONFIG_RESPONSE")
		} else if header == llrp.ROAccessReportHeader {
			log.Println(">>> RO_ACCESS_REPORT")
			log.Printf("Packet size: %v\n", reqLen)
			log.Printf("% x\n", buf[:reqLen])
		} else {
			log.Printf("Unknown header: %v\n", header)
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
