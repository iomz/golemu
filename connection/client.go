//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package connection

import (
	"io"
	"net"
	"strconv"
	"time"

	"github.com/iomz/go-llrp"
	log "github.com/sirupsen/logrus"
)

// Client handles LLRP client connections
type Client struct {
	ip   string
	port int
}

// NewClient creates a new LLRP client
func NewClient(ip string, port int) *Client {
	return &Client{
		ip:   ip,
		port: port,
	}
}

// Run starts the client and connects to the LLRP server
func (c *Client) Run() int {
	log.Infof("waiting for %s:%d ...", c.ip, c.port)
	conn, err := net.Dial("tcp", c.ip+":"+strconv.Itoa(c.port))
	for err != nil {
		time.Sleep(time.Second)
		conn, err = net.Dial("tcp", c.ip+":"+strconv.Itoa(c.port))
	}
	defer conn.Close()
	log.Infof("established an LLRP connection with %v", conn.RemoteAddr())

	for {
		hdr, messageValue, err := ReadLLRPMessage(conn)
		if err != nil {
			if err == io.EOF {
				log.Info("connection closed by server")
				return 0
			}
			log.Errorf("error reading LLRP message: %v", err)
			return 1
		}

		c.handleMessage(conn, hdr.Header, hdr.MessageID, messageValue)
	}
}

func (c *Client) handleMessage(conn net.Conn, header uint16, messageID uint32, messageValue []byte) {
	switch header {
	case llrp.ReaderEventNotificationHeader:
		log.WithFields(log.Fields{
			"Message ID": messageID,
		}).Info(">>> READER_EVENT_NOTIFICATION")
		conn.Write(llrp.SetReaderConfig(messageID + 1))
	case llrp.KeepaliveHeader:
		log.WithFields(log.Fields{
			"Message ID": messageID,
		}).Info(">>> KEEP_ALIVE")
		conn.Write(llrp.KeepaliveAck(messageID + 1))
	case llrp.SetReaderConfigResponseHeader:
		log.WithFields(log.Fields{
			"Message ID": messageID,
		}).Info(">>> SET_READER_CONFIG_RESPONSE")
	case llrp.ROAccessReportHeader:
		res := llrp.UnmarshalROAccessReportBody(messageValue)
		log.WithFields(log.Fields{
			"Message ID": messageID,
			"#Events":    len(res),
		}).Info(">>> RO_ACCESS_REPORT")
	default:
		log.WithFields(log.Fields{
			"Message ID": messageID,
		}).Warnf("Unknown header: %v", header)
	}
}
