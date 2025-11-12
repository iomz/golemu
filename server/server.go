//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package server

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/iomz/go-llrp"
	"github.com/iomz/go-llrp/binutil"
	"github.com/iomz/golemu/api"
	"github.com/iomz/golemu/connection"
	"github.com/iomz/golemu/tag"
	log "github.com/sirupsen/logrus"
)

// Server represents the LLRP server
type Server struct {
	ip                string
	port              int
	apiPort           int
	file              string
	pdu               int
	reportInterval    int
	keepaliveInterval int
	initialMessageID  int
	tagManagerChan    chan tag.Manager
	tagUpdatedChan    chan llrp.Tags
	tagService        *tag.ManagerService
	isConnAlive       *atomic.Bool
	llrpHandler       *connection.Handler
}

// NewServer creates a new LLRP server
func NewServer(ip string, port, apiPort, pdu, reportInterval, keepaliveInterval, initialMessageID int, file string) *Server {
	tagManagerChan := make(chan tag.Manager)
	tagUpdatedChan := make(chan llrp.Tags)
	isConnAlive := &atomic.Bool{}
	tagService := tag.NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)
	llrpHandler := connection.NewHandler(initialMessageID, pdu, reportInterval, keepaliveInterval, tagUpdatedChan)

	return &Server{
		ip:                ip,
		port:              port,
		apiPort:           apiPort,
		file:              file,
		pdu:               pdu,
		reportInterval:    reportInterval,
		keepaliveInterval: keepaliveInterval,
		initialMessageID:  initialMessageID,
		tagManagerChan:    tagManagerChan,
		tagUpdatedChan:    tagUpdatedChan,
		tagService:        tagService,
		isConnAlive:       isConnAlive,
		llrpHandler:       llrpHandler,
	}
}

// Run starts the server
func (s *Server) Run() int {
	s.loadTags()

	l, err := net.Listen("tcp", s.ip+":"+strconv.Itoa(s.port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	log.Infof("listening on %v:%v", s.ip, s.port)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start API server
	apiServer := api.NewServer(s.apiPort, s.tagManagerChan)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Errorf("API server error: %v", err)
		}
	}()

	// Start tag manager
	go s.runTagManager(signals)

	// Handle LLRP connections
	log.Info("starting LLRP connection...")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			continue
		}
		log.Info("LLRP connection initiated")

		if err := s.llrpHandler.SendReaderEventNotification(conn); err != nil {
			log.Errorf("error sending READER_EVENT_NOTIFICATION: %v", err)
			conn.Close()
			continue
		}
		go s.llrpHandler.HandleRequest(conn, s.tagService.GetTags())
	}
}

func (s *Server) loadTags() {
	log.WithFields(log.Fields{
		"File": s.file,
	}).Info("loading tags")

	if _, err := os.Stat(s.file); os.IsNotExist(err) {
		log.Warnf("%v doesn't exist, couldn't load tags", s.file)
		return
	}

	var tags llrp.Tags
	err := binutil.Load(s.file, &tags)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("%v tags loaded", len(tags))
	s.tagService.SetTags(tags)
}

func (s *Server) runTagManager(signals chan os.Signal) {
	for {
		select {
		case cmd := <-s.tagManagerChan:
			s.tagService.Process(cmd)
		case sig := <-signals:
			log.Fatalf("%v", sig)
		}
	}
}
