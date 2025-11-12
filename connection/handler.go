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

// Handler handles LLRP connections
type Handler struct {
	currentMessageID  *uint32
	pdu               int
	reportInterval    int
	keepaliveInterval int
	isConnAlive       *atomic.Bool
	tagUpdatedChan    chan llrp.Tags
}

// NewHandler creates a new LLRP handler
func NewHandler(initialMessageID int, pdu, reportInterval, keepaliveInterval int, tagUpdatedChan chan llrp.Tags) *Handler {
	msgID := uint32(initialMessageID)
	return &Handler{
		currentMessageID:  &msgID,
		pdu:               pdu,
		reportInterval:    reportInterval,
		keepaliveInterval: keepaliveInterval,
		isConnAlive:       &atomic.Bool{},
		tagUpdatedChan:    tagUpdatedChan,
	}
}

// HandleRequest handles incoming LLRP requests
func (h *Handler) HandleRequest(conn net.Conn, tags llrp.Tags) {
	defer conn.Close()
	trds := tags.BuildTagReportDataStack(h.pdu)

	for {
		hdr, _, err := ReadLLRPMessage(conn)
		if err == io.EOF {
			log.Info("the client is disconnected, closing LLRP connection")
			conn.Close()
			return
		} else if err != nil {
			log.Infof("closing LLRP connection due to %s", err.Error())
			conn.Close()
			return
		}

		if hdr.Header == llrp.SetReaderConfigHeader || hdr.Header == llrp.KeepaliveAckHeader {
			if hdr.Header == llrp.SetReaderConfigHeader {
				log.Info(">>> SET_READER_CONFIG")
				if _, err := conn.Write(llrp.SetReaderConfigResponse(*h.currentMessageID)); err != nil {
					log.Warnf("error writing SET_READER_CONFIG_RESPONSE: %v", err)
					conn.Close()
					return
				}
				atomic.AddUint32(h.currentMessageID, 1)
				log.Info("<<< SET_READER_CONFIG_RESPONSE")
			} else if hdr.Header == llrp.KeepaliveAckHeader {
				log.Info(">>> KEEP_ALIVE_ACK")
			}

			h.startReportLoop(conn, trds)
		} else {
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

		for {
			h.isConnAlive.Store(true)
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
				roarTicker.Stop()
				if h.keepaliveInterval != 0 {
					keepaliveTicker.Stop()
				}
				break
			}
		}
	}()
}

// SendReaderEventNotification sends a READER_EVENT_NOTIFICATION message
func (h *Handler) SendReaderEventNotification(conn net.Conn) error {
	currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
	if _, err := conn.Write(llrp.ReaderEventNotification(*h.currentMessageID, currentTime)); err != nil {
		return err
	}
	log.Info("<<< READER_EVENT_NOTIFICATION")
	atomic.AddUint32(h.currentMessageID, 1)
	return nil
}

// IsConnAlive returns whether the connection is alive
func (h *Handler) IsConnAlive() bool {
	return h.isConnAlive.Load()
}
