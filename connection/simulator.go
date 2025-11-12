//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package connection

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/iomz/go-llrp"
	"github.com/iomz/go-llrp/binutil"
	log "github.com/sirupsen/logrus"
)

// Simulator handles LLRP simulation mode
type Simulator struct {
	ip               string
	port             int
	pdu              int
	reportInterval   int
	simulationDir    string
	currentMessageID *uint32
}

// NewSimulator creates a new simulator
func NewSimulator(ip string, port, pdu, reportInterval int, simulationDir string, initialMessageID int) *Simulator {
	msgID := uint32(initialMessageID)
	return &Simulator{
		ip:               ip,
		port:             port,
		pdu:              pdu,
		reportInterval:   reportInterval,
		simulationDir:    simulationDir,
		currentMessageID: &msgID,
	}
}

// Run starts the simulator
func (s *Simulator) Run() int {
	simulationFiles, err := s.loadSimulationFiles()
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", s.ip+":"+strconv.Itoa(s.port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	log.Infof("listening on %v:%v", s.ip, s.port)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		log.Infof("received signal %v, shutting down...", sig)
		os.Exit(0)
	}()

	log.Info("waiting for LLRP connection...")
	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Infof("initiated LLRP connection with %v", conn.RemoteAddr())

	// Send READER_EVENT_NOTIFICATION
	currentTime := uint64(time.Now().UTC().Nanosecond() / 1000)
	if _, err := conn.Write(llrp.ReaderEventNotification(*s.currentMessageID, currentTime)); err != nil {
		log.Fatalf("error sending READER_EVENT_NOTIFICATION: %v", err)
	}
	log.Info("<<< READER_EVENT_NOTIFICATION")
	atomic.AddUint32(s.currentMessageID, 1)

	eventCycle := 0
	tags, err := s.loadTagsForNextEventCycle(simulationFiles, &eventCycle)
	if err != nil {
		log.Fatal(err)
	}
	eventCycle++
	trds := tags.BuildTagReportDataStack(s.pdu)
	roarTicker := time.NewTicker(time.Duration(s.reportInterval) * time.Millisecond)

	for {
		hdr, _, err := ReadLLRPMessage(conn)
		if err != nil {
			log.Fatalf("error reading LLRP message: %v", err)
		}

		if hdr.Header == llrp.SetReaderConfigHeader {
			if _, err := conn.Write(llrp.SetReaderConfigResponse(*s.currentMessageID)); err != nil {
				log.Fatalf("error writing SET_READER_CONFIG_RESPONSE: %v", err)
			}
			atomic.AddUint32(s.currentMessageID, 1)

			s.startSimulationLoop(conn, simulationFiles, &eventCycle, trds, roarTicker)
		} else {
			log.Warnf(">>> header: %v", hdr.Header)
		}
	}
	return 0
}

func (s *Simulator) loadSimulationFiles() ([]string, error) {
	dir, err := filepath.Abs(s.simulationDir)
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	simulationFiles := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gob") {
			simulationFiles = append(simulationFiles, path.Join(dir, f.Name()))
		}
	}
	if len(simulationFiles) == 0 {
		return nil, fmt.Errorf("no event cycle file found in %s", s.simulationDir)
	}
	return simulationFiles, nil
}

func (s *Simulator) loadTagsForNextEventCycle(simulationFiles []string, eventCycle *int) (llrp.Tags, error) {
	tags := llrp.Tags{}
	if len(simulationFiles) <= *eventCycle {
		log.Debugf("Total iteration: %v, current event cycle: %v", len(simulationFiles), eventCycle)
		log.Infof("Resetting event cycle from %v to 0", *eventCycle)
		*eventCycle = 0
	}
	err := binutil.Load(simulationFiles[*eventCycle], &tags)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func (s *Simulator) startSimulationLoop(conn net.Conn, simulationFiles []string, eventCycle *int, trds llrp.TagReportDataStack, roarTicker *time.Ticker) chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			<-roarTicker.C
			tags, err := s.loadTagsForNextEventCycle(simulationFiles, eventCycle)
			if err != nil {
				log.Warn(err)
				continue
			}
			*eventCycle++
			trds = tags.BuildTagReportDataStack(s.pdu)

			log.Infof("<<< Simulated Event Cycle %v, %v tags, %v roars", *eventCycle-1, len(tags), len(trds))
			for _, trd := range trds {
				roar := llrp.NewROAccessReport(trd.Data, *s.currentMessageID)
				if err := roar.Send(conn); err != nil {
					log.Errorf("error sending RO_ACCESS_REPORT: %v", err)
					return
				}
				atomic.AddUint32(s.currentMessageID, 1)
			}
		}
	}()
	return done
}
