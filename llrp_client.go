package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
)

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
var messageID = 10000

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

func buildSetReaderConfig() []byte {
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
	return buf.Bytes()
}

func main() {
	flag.Parse()

	// Establish a connection to the llrp client
	conn, err := net.Dial("tcp",
		llrpHost+":"+strconv.Itoa(llrpPort))
	check(err)

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
		if header == HEADER_REN {
			fmt.Println(">>> READER_EVENT_NOTIFICATION")
			conn.Write(buildSetReaderConfig())
		} else if header == HEADER_SRCR {
			fmt.Println(">>> SET_READER_CONFIG_RESPONSE")
		} else if header == HEADER_ROAR {
			fmt.Println(">>> RO_ACCESS_REPORT")
			fmt.Printf("Packet size: %v\n", reqLen)
			fmt.Printf("% x\n", buf[:reqLen])
		} else {
			fmt.Printf("Unknown header: %v\n", header)
		}
	}
}
