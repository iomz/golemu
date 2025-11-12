//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package connection

import (
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/iomz/go-llrp"
	log "github.com/sirupsen/logrus"
)

// Handler manages LLRP protocol connections and handles incoming requests from LLRP clients.
// It processes SET_READER_CONFIG and KEEP_ALIVE_ACK messages, sends RO_ACCESS_REPORT messages
// at regular intervals, and manages keepalive messages to maintain the connection.
type Handler struct {
	currentMessageID  *uint32
	pdu               int
	reportInterval    int
	keepaliveInterval int
	isConnAlive       *atomic.Bool
	reportLoopStarted *atomic.Bool
	tagUpdatedChan    chan llrp.Tags
}

// NewHandler creates and initializes a new LLRP handler with the specified configuration.
//
// Parameters:
//   - initialMessageID: The starting message ID for LLRP messages
//   - pdu: Maximum Protocol Data Unit size in bytes
//   - reportInterval: Interval in milliseconds between RO_ACCESS_REPORT messages
//   - keepaliveInterval: Interval in seconds for keepalive messages (0 to disable)
//   - tagUpdatedChan: Channel for receiving tag updates
//   - isConnAlive: Atomic boolean flag indicating connection status
func NewHandler(initialMessageID int, pdu, reportInterval, keepaliveInterval int, tagUpdatedChan chan llrp.Tags, isConnAlive *atomic.Bool) *Handler {
	msgID := uint32(initialMessageID)
	return &Handler{
		currentMessageID:  &msgID,
		pdu:               pdu,
		reportInterval:    reportInterval,
		keepaliveInterval: keepaliveInterval,
		isConnAlive:       isConnAlive,
		reportLoopStarted: &atomic.Bool{},
		tagUpdatedChan:    tagUpdatedChan,
	}
}

// HandleRequest processes incoming LLRP requests from a client connection.
// It reads messages from the connection, handles SET_READER_CONFIG and KEEP_ALIVE_ACK messages,
// and starts the report loop when appropriate. The function runs until the connection is closed
// or an error occurs.
//
// Parameters:
//   - conn: The network connection to the LLRP client
//   - tags: Initial set of tags to include in reports
func (h *Handler) HandleRequest(conn net.Conn, tags llrp.Tags) {
	defer conn.Close()
	trds := tags.BuildTagReportDataStack(h.pdu)

	for {
		hdr, _, err := ReadLLRPMessage(conn)
		if err == io.EOF {
			log.Info("the client is disconnected, closing LLRP connection")
			return
		} else if err != nil {
			log.Infof("closing LLRP connection due to %s", err.Error())
			return
		}

		switch hdr.Header {
		case llrp.SetReaderConfigHeader:
			log.Info(">>> SET_READER_CONFIG")
			if _, err := conn.Write(llrp.SetReaderConfigResponse(*h.currentMessageID)); err != nil {
				log.Warnf("error writing SET_READER_CONFIG_RESPONSE: %v", err)
				return
			}
			atomic.AddUint32(h.currentMessageID, 1)
			log.Info("<<< SET_READER_CONFIG_RESPONSE")
			if h.reportLoopStarted.CompareAndSwap(false, true) {
				h.startReportLoop(conn, trds)
			}
		case llrp.KeepaliveAckHeader:
			log.Info(">>> KEEP_ALIVE_ACK")
			if h.reportLoopStarted.CompareAndSwap(false, true) {
				h.startReportLoop(conn, trds)
			}
		default:
			log.Warnf("unknown header: %v", hdr.Header)
			return
		}
	}
}

func (h *Handler) startReportLoop(conn net.Conn, trds llrp.TagReportDataStack) {
	roarTicker := time.NewTicker(time.Duration(h.reportInterval) * time.Millisecond)
	keepaliveTicker := &time.Ticker{}
	if h.keepaliveInterval != 0 {
		keepaliveTicker = time.NewTicker(time.Duration(h.keepaliveInterval) * time.Second)
	}

	go func() {
		defer roarTicker.Stop()
		if h.keepaliveInterval != 0 {
			defer keepaliveTicker.Stop()
		}
		defer h.reportLoopStarted.Store(false)
		// Initial ROAR message
		log.WithFields(log.Fields{
			"Reports":    len(trds),
			"Total tags": trds.TotalTagCounts(),
		}).Info("<<< RO_ACCESS_REPORT")
		for _, trd := range trds {
			roar := llrp.NewROAccessReport(trd.Data, *h.currentMessageID)
			err := roar.Send(conn)
			atomic.AddUint32(h.currentMessageID, 1)
			if err != nil {
				log.Warn(err)
				h.isConnAlive.Store(false)
				return
			}
		}

		// Mark connection as alive before entering the main loop
		h.isConnAlive.Store(true)

		for {
			select {
			case <-roarTicker.C:
				log.WithFields(log.Fields{
					"Reports":    len(trds),
					"Total tags": trds.TotalTagCounts(),
				}).Info("<<< RO_ACCESS_REPORT")
				for _, trd := range trds {
					roar := llrp.NewROAccessReport(trd.Data, *h.currentMessageID)
					err := roar.Send(conn)
					atomic.AddUint32(h.currentMessageID, 1)
					if err != nil {
						log.Warn(err)
						h.isConnAlive.Store(false)
						break
					}
				}
			case <-keepaliveTicker.C:
				log.Info("<<< KEEP_ALIVE")
				if _, err := conn.Write(llrp.Keepalive(*h.currentMessageID)); err != nil {
					log.Warnf("error writing KEEP_ALIVE: %v", err)
					h.isConnAlive.Store(false)
				} else {
					atomic.AddUint32(h.currentMessageID, 1)
				}
			case tags := <-h.tagUpdatedChan:
				log.Debug("TagUpdated")
				trds = tags.BuildTagReportDataStack(h.pdu)
			}
			if !h.isConnAlive.Load() {
				break
			}
		}
	}()
}

// SendReaderEventNotification sends a READER_EVENT_NOTIFICATION message to the client
// to indicate that the reader is ready to accept connections.
// This is typically the first message sent after establishing an LLRP connection.
//
// Returns an error if the message cannot be written to the connection.
func (h *Handler) SendReaderEventNotification(conn net.Conn) error {
	currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
	if _, err := conn.Write(llrp.ReaderEventNotification(*h.currentMessageID, currentTime)); err != nil {
		return err
	}
	log.Info("<<< READER_EVENT_NOTIFICATION")
	atomic.AddUint32(h.currentMessageID, 1)
	return nil
}

// IsConnAlive returns the current connection status.
// It returns true if the connection is active and false otherwise.
func (h *Handler) IsConnAlive() bool {
	return h.isConnAlive.Load()
}
