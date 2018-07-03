// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/structs"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/iomz/go-llrp"
	"github.com/iomz/golemu"
	"golang.org/x/net/websocket"
	"gopkg.in/alecthomas/kingpin.v2"
)

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

	// LLRPConn flag
	isLLRPConnAlive = false
	// Current messageID
	messageID = uint32(*initialMessageID)
	// Current KeepaliveID
	keepaliveID = *initialKeepaliveID
	// Current activeClients
	activeClients = make(map[WebsockConn]int) // map containing clients
	// Tag management channel
	tagManagerChannel = make(chan golemu.TagManager)
	// notify tag update channel
	notify = make(chan bool)
	// update TagReportDataStack when tag is updated
	tagUpdated = make(chan []*golemu.Tag)
)

// WebsocketMessage to unmarshal JSON message from web clients
type WebsocketMessage struct {
	UpdateType string
	Tag        golemu.TagRecord
	Tags       []map[string]interface{}
}

// WebsockConn holds connection consists of the websocket and the client ip
type WebsockConn struct {
	websocket *websocket.Conn
	clientIP  string
}

// APIPostTag redirects the tag addition request
func APIPostTag(c *gin.Context) {
	var json []golemu.TagRecord
	c.BindWith(&json, binding.JSON)
	if res := ReqAddTag("add", json); res == "error" {
		c.String(http.StatusAlreadyReported, "The tag already exists!\n")
	} else {
		c.String(http.StatusAccepted, "Post requested!\n")
	}
}

// APIDeleteTag redirects the tag deletion request
func APIDeleteTag(c *gin.Context) {
	var json []golemu.TagRecord
	c.BindWith(&json, binding.JSON)
	if res := ReqDeleteTag("delete", json); res == "error" {
		c.String(http.StatusNoContent, "The tag doesn't exist!\n")
	} else {
		c.String(http.StatusAccepted, "Delete requested!\n")
	}
}

// Broadcast a message vi websocket
func Broadcast(clientMessage []byte) {
	for cs := range activeClients {
		if err := websocket.Message.Send(cs.websocket, string(clientMessage)); err != nil {
			// we could not send the message to a peer
			log.Printf("could not send message to %v", cs.clientIP)
			log.Print(err)
		}
	}
}

// ReqAddTag handles a tag addition request
func ReqAddTag(ut string, req []golemu.TagRecord) string {
	// TODO: success/fail notification per tag
	failed := false
	for _, t := range req {
		tag, err := golemu.NewTag(&golemu.TagRecord{
			PCBits: t.PCBits,
			EPC:    t.EPC,
		})
		if err != nil {
			log.Fatal(err)
		}

		add := golemu.TagManager{
			Action: golemu.AddTags,
			Tags:   []*golemu.Tag{tag},
		}
		tagManagerChannel <- add

		if add = <-tagManagerChannel; len(add.Tags) != 0 {
			m := WebsocketMessage{
				UpdateType: "add",
				Tag:        t,
				Tags:       []map[string]interface{}{}}
			clientMessage, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			Broadcast(clientMessage)
		} else {
			failed = true
		}
	}

	if failed {
		log.Printf("failed %v %v", ut, req)
		return "error"
	}
	log.Printf("%v %v", ut, req)
	return ut
}

// ReqDeleteTag handles a tag deletion request
func ReqDeleteTag(ut string, req []golemu.TagRecord) string {
	// TODO: success/fail notification per tag
	failed := false
	for _, t := range req {
		tag, err := golemu.NewTag(&golemu.TagRecord{
			PCBits: t.PCBits,
			EPC:    t.EPC,
		})
		if err != nil {
			panic(err)
		}

		delete := golemu.TagManager{
			Action: golemu.DeleteTags,
			Tags:   []*golemu.Tag{tag},
		}
		tagManagerChannel <- delete

		if delete = <-tagManagerChannel; len(delete.Tags) != 0 {
			m := WebsocketMessage{
				UpdateType: "delete",
				Tag:        t,
				Tags:       []map[string]interface{}{}}
			clientMessage, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			Broadcast(clientMessage)
		} else {
			failed = true
		}
	}
	if failed {
		log.Printf("failed %v %v", ut, req)
		return "error"
	}
	log.Printf("%v %v", ut, req)
	return ut
}

// ReqRetrieveTag handles a tag retrieval request
func ReqRetrieveTag() []map[string]interface{} {
	retrieve := golemu.TagManager{
		Action: golemu.RetrieveTags,
		Tags:   []*golemu.Tag{},
	}
	tagManagerChannel <- retrieve
	retrieve = <-tagManagerChannel
	var tagList []map[string]interface{}
	for _, tag := range retrieve.Tags {
		t := structs.Map(tag.InString())
		tagList = append(tagList, t)
	}
	log.Printf("retrieve: %v", tagList)
	return tagList
}

// SockServer to handle messaging between clients
func SockServer(ws *websocket.Conn) {
	var err error
	//var clientMessage string
	// use []byte if websocket binary type is blob or arraybuffer
	var clientMessage []byte

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Print(err)
		}
	}()

	client := ws.Request().RemoteAddr
	log.Printf("client connected: %v", client)
	clientSock := WebsockConn{ws, client}
	activeClients[clientSock] = 0
	log.Printf("number of clients connected: %v", len(activeClients))

	// for loop so the websocket stays open otherwise
	// it'll close after one Receieve and Send
	for {
		if err = websocket.Message.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Printf("websocket Disconnected waiting %v", err.Error())
			// remove the ws client conn from our active clients
			delete(activeClients, clientSock)
			log.Printf("number of clients still connected ... %v", len(activeClients))
			return
		}

		//clientMessage = clientSock.clientIP + " Said: " + clientMessage

		// Parse the JSON
		m := WebsocketMessage{}
		if err = json.Unmarshal(clientMessage, &m); err != nil {
			log.Print(err)
		}

		// Handle the command
		// Compose result struct containing proper parameters
		// TODO: separate actions into functions
		switch m.UpdateType {
		case "add":
			m.UpdateType = ReqAddTag(m.UpdateType, []golemu.TagRecord{m.Tag})
		case "delete":
			m.UpdateType = ReqDeleteTag(m.UpdateType, []golemu.TagRecord{m.Tag})
		case "retrieve":
			tagList := ReqRetrieveTag()
			m = WebsocketMessage{
				UpdateType: "retrieval",
				Tag:        golemu.TagRecord{},
				Tags:       tagList}
			clientMessage, err = json.Marshal(m)
			if err != nil {
				panic(err)
			}
			Broadcast(clientMessage)
		default:
			log.Printf("unknown UpdateType: %v", m.UpdateType)
		}
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, tags []*golemu.Tag) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, *pdu)
	trds := golemu.BuildTagReportDataStack(tags, *pdu)

	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err == io.EOF {
			// Close the connection when you're done with it.
			log.Println("the client is disconnected, closing LLRP connection")
			conn.Close()
			return
		} else if err != nil {
			log.Println("closing LLRP connection")
			log.Print(err)
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
				log.Println(">>> KEEP_ALIVE_ACK")
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
					select {
					// ROAccessReport interval tick
					case <-roarTicker.C:
						log.Printf("<<< RO_ACCESS_REPORT (# reports: %v, # total tags: %v)", len(trds.Stack), trds.TotalTagCounts())
						err := golemu.SendROAccessReport(conn, trds, &messageID)
						if err != nil {
							log.Print(err)
							isLLRPConnAlive = false
						}
					// Keepalive interval tick
					case <-keepaliveTicker.C:
						log.Println("<<< KEEP_ALIVE")
						conn.Write(llrp.Keepalive())
						isLLRPConnAlive = false
					// When the tag queue is updated
					case tags := <-tagUpdated:
						log.Println("### TagUpdated")
						trds = golemu.BuildTagReportDataStack(tags, *pdu)
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
			log.Printf("unknown header: %v, reqlen: %v", header, reqLen)
			log.Printf("message: %v", buf)
			return
		}
	}
}

// server mode
func runServer() int {
	// Read virtual tags from a csv file
	log.Printf("loading virtual Tags from \"%v\"", *file)

	if _, err := os.Stat(*file); os.IsNotExist(err) {
		_, err := os.Create(*file)
		if err != nil {
			panic(err)
		}
		log.Printf("%v created.", *file)
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
	tags := golemu.LoadTagsFromCSV(*file)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", ip.String()+":"+strconv.Itoa(*port))
	if err != nil {
		panic(err)
	}

	// Close the listener when the application closes.
	defer l.Close()
	log.Printf("listening on %v:%v", ip, *port)

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
			case cmd := <-tagManagerChannel:
				// Tag management
				res := []*golemu.Tag{}
				switch cmd.Action {
				case golemu.AddTags:
					for _, t := range cmd.Tags {
						if i := golemu.GetIndexOfTag(*tags, t); i < 0 {
							*tags = append(*tags, t)
							res = append(res, t)
							// Write to file
							//writeTagsToCSV(*tags, *file)
							if isLLRPConnAlive {
								tagUpdated <- *tags
							}
						}
					}
				case golemu.DeleteTags:
					for _, t := range cmd.Tags {
						if i := golemu.GetIndexOfTag(*tags, t); i >= 0 {
							*tags = append((*tags)[:i], (*tags)[i+1:]...)
							res = append(res, t)
							// Write to file
							//writeTagsToCSV(tags, *file)
							if isLLRPConnAlive {
								tagUpdated <- *tags
							}
						}
					}
				case golemu.RetrieveTags:
					res = *tags
				}
				cmd.Tags = res
				tagManagerChannel <- cmd
			case signal := <-signals:
				// Handle SIGINT and SIGTERM.
				log.Fatalf("%v", signal)
			}
		}
	}()

	// Handle LLRP connection
	log.Println("starting LLRP connection...")
	for {
		// Accept an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		log.Println("LLRP connection initiated")

		// Send back READER_EVENT_NOTIFICATION
		currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
		conn.Write(llrp.ReaderEventNotification(messageID, currentTime))
		log.Println("<<< READER_EVENT_NOTIFICATION")
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
	if err != nil {
		panic(err)
	}

	header := make([]byte, 2)
	length := make([]byte, 4)
	for {
		_, err = io.ReadFull(conn, header)
		if err != nil {
			log.Fatal(err)
		}
		//length := binary.BigEndian.Uint32(prefix)

		h := binary.BigEndian.Uint16(header)
		if h == llrp.ReaderEventNotificationHeader {
			_, err = io.ReadFull(conn, length)
			message := make([]byte, binary.BigEndian.Uint32(length)-6)
			_, err = io.ReadFull(conn, message)
			log.Println(">>> READER_EVENT_NOTIFICATION")
			conn.Write(llrp.SetReaderConfig(messageID))
		} else if h == llrp.KeepaliveHeader {
			_, err = io.ReadFull(conn, length)
			message := make([]byte, binary.BigEndian.Uint32(length)-6)
			_, err = io.ReadFull(conn, message)
			log.Println(">>> KEEP_ALIVE")
			conn.Write(llrp.KeepaliveAck())
		} else if h == llrp.SetReaderConfigResponseHeader {
			_, err = io.ReadFull(conn, length)
			message := make([]byte, binary.BigEndian.Uint32(length)-6)
			_, err = io.ReadFull(conn, message)
			log.Println(">>> SET_READER_CONFIG_RESPONSE")
		} else if h == llrp.ROAccessReportHeader {
			_, err = io.ReadFull(conn, length)
			l := binary.BigEndian.Uint32(length)
			message := make([]byte, l-6)
			_, err = io.ReadFull(conn, message)
			log.Println(">>> RO_ACCESS_REPORT")
			golemu.DecapsulateROAccessReport(l, message)
		} else {
			log.Fatalf("Unknown header: %v", h)
		}
	}
}

func main() {
	app.Version(version)
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		//loggo.ConfigureLoggers("TRACE")
		gin.SetMode(gin.DebugMode)
	} else {
		//loggo.ConfigureLoggers("INFO")
		gin.SetMode(gin.ReleaseMode)
	}

	switch parse {
	case server.FullCommand():
		os.Exit(runServer())
	case client.FullCommand():
		os.Exit(runClient())
	}
}
