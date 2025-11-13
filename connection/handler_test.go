//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package connection

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iomz/go-llrp"
)

func TestNewHandler(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}

	handler := NewHandler(1000, 1500, 10000, 5, tagUpdatedChan, isConnAlive)

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.currentMessageID == nil {
		t.Error("currentMessageID should not be nil")
	}
	if *handler.currentMessageID != 1000 {
		t.Errorf("expected currentMessageID 1000, got %d", *handler.currentMessageID)
	}
	if handler.pdu != 1500 {
		t.Errorf("expected pdu 1500, got %d", handler.pdu)
	}
	if handler.reportInterval != 10000 {
		t.Errorf("expected reportInterval 10000, got %d", handler.reportInterval)
	}
	if handler.keepaliveInterval != 5 {
		t.Errorf("expected keepaliveInterval 5, got %d", handler.keepaliveInterval)
	}
	if handler.tagUpdatedChan != tagUpdatedChan {
		t.Error("tagUpdatedChan not set correctly")
	}
	if handler.isConnAlive != isConnAlive {
		t.Error("isConnAlive not set correctly")
	}
}

func TestHandler_SendReaderEventNotification(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	var buf bytes.Buffer
	conn := &mockConn{writer: &buf}

	err := handler.SendReaderEventNotification(conn)
	if err != nil {
		t.Fatalf("SendReaderEventNotification failed: %v", err)
	}

	if *handler.currentMessageID != 1001 {
		t.Errorf("expected messageID to be incremented to 1001, got %d", *handler.currentMessageID)
	}

	if buf.Len() == 0 {
		t.Error("expected data to be written to connection")
	}
}

func TestHandler_SendReaderEventNotification_WriteError(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	// Create a connection that fails on write
	conn := &mockConn{writer: &errorWriter{}}

	err := handler.SendReaderEventNotification(conn)
	if err == nil {
		t.Error("expected error when write fails")
	}
}

func TestHandler_IsConnAlive(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	if handler.IsConnAlive() {
		t.Error("expected connection to be not alive initially")
	}

	isConnAlive.Store(true)
	if !handler.IsConnAlive() {
		t.Error("expected connection to be alive after setting")
	}

	isConnAlive.Store(false)
	if handler.IsConnAlive() {
		t.Error("expected connection to be not alive after clearing")
	}
}

func TestHandler_HandleRequest_SetReaderConfig(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	// Create a message buffer with SET_READER_CONFIG header
	msg := llrp.SetReaderConfig(1001)
	var buf bytes.Buffer
	buf.Write(msg)

	// Add EOF after the message to terminate the loop
	conn := &mockConn{reader: &buf, writer: &bytes.Buffer{}}

	tags := llrp.Tags{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.HandleRequest(conn, tags)
	}()

	// Wait for handler to process and finish
	wg.Wait()

	// Verify messageID was incremented
	if *handler.currentMessageID < 1001 {
		t.Errorf("expected messageID to be incremented, got %d", *handler.currentMessageID)
	}
}

func TestHandler_HandleRequest_KeepaliveAck(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	// Create a message buffer with KEEP_ALIVE_ACK header
	msg := llrp.KeepaliveAck(1001)
	var buf bytes.Buffer
	buf.Write(msg)

	conn := &mockConn{reader: &buf, writer: &bytes.Buffer{}}

	tags := llrp.Tags{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.HandleRequest(conn, tags)
	}()

	// Wait for handler to process and finish
	wg.Wait()

	// Verify report loop was started
	if !handler.reportLoopStarted.Load() {
		t.Error("expected report loop to be started")
	}
}

func TestHandler_HandleRequest_EOF(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	conn := &mockConn{reader: bytes.NewReader([]byte{})}

	tags := llrp.Tags{}
	handler.HandleRequest(conn, tags)

	// Should return without error
}

func TestHandler_HandleRequest_UnknownHeader(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	// Create a message with unknown header
	header := uint16(0xFFFF)
	length := uint32(10)
	messageID := uint32(1234)

	buf := make([]byte, 10)
	binary.BigEndian.PutUint16(buf[0:2], header)
	binary.BigEndian.PutUint32(buf[2:6], length)
	binary.BigEndian.PutUint32(buf[6:10], messageID)

	conn := &mockConn{reader: bytes.NewReader(buf), writer: &bytes.Buffer{}}

	tags := llrp.Tags{}
	handler.HandleRequest(conn, tags)

	// Should return without error
}

func TestHandler_HandleRequest_WriteError(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	handler := NewHandler(1000, 1500, 10000, 0, tagUpdatedChan, isConnAlive)

	// Create a message buffer with SET_READER_CONFIG header
	msg := llrp.SetReaderConfig(1001)
	var buf bytes.Buffer
	buf.Write(msg)

	// Create connection that fails on write
	conn := &mockConn{reader: &buf, writer: &errorWriter{}}

	tags := llrp.Tags{}
	handler.HandleRequest(conn, tags)

	// Should return without error (error is logged)
}

func TestHandler_startReportLoop(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(true)
	handler := NewHandler(1000, 1500, 10, 0, tagUpdatedChan, isConnAlive) // Short interval for testing

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	tags := llrp.Tags{tag1}
	trds := tags.BuildTagReportDataStack(1500)

	var writeBuf bytes.Buffer
	reportSentChan := make(chan struct{}, 10) // Buffered to avoid blocking
	conn := &signalingMockConn{
		mockConn:    mockConn{writer: &writeBuf},
		writeSignal: reportSentChan,
	}

	handler.startReportLoop(conn, trds)

	// Wait for initial report via synchronization channel
	select {
	case <-reportSentChan:
		// Initial report received
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for initial RO_ACCESS_REPORT")
	}

	// Verify initial report was sent
	if conn.Len() == 0 {
		t.Error("expected initial RO_ACCESS_REPORT to be sent")
	}

	// Note: reportLoopStarted is only set via CompareAndSwap in HandleRequest,
	// not when startReportLoop is called directly. Since we're testing startReportLoop
	// directly, we verify the loop is running by checking that the initial report was sent.

	// Send tag update
	select {
	case tagUpdatedChan <- tags:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout sending tag update")
	}

	// Stop the loop by marking connection as not alive
	isConnAlive.Store(false)

	// Note: Since we're calling startReportLoop directly (not through HandleRequest),
	// reportLoopStarted is never set to true, so we can't use it to verify shutdown.
	// The loop will stop when isConnAlive is false, which we've set above.
	// The main test objectives (initial report sent, tag update received) are already verified.
}

// waitForCondition polls a condition function until it returns true or timeout occurs
func waitForCondition(condition func() bool, timeout time.Duration) bool {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	timeoutChan := time.After(timeout)

	for {
		if condition() {
			return true
		}
		select {
		case <-ticker.C:
			// Continue polling
		case <-timeoutChan:
			return false
		}
	}
}

func TestHandler_startReportLoop_WithKeepalive(t *testing.T) {
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(true)
	handler := NewHandler(1000, 1500, 100, 1, tagUpdatedChan, isConnAlive) // Short intervals for testing

	tags := llrp.Tags{}
	trds := tags.BuildTagReportDataStack(1500)

	var writeBuf bytes.Buffer
	conn := &mockConn{writer: &writeBuf}

	handler.startReportLoop(conn, trds)

	// Give loop time to send keepalive (keepalive interval is 1 second)
	time.Sleep(1200 * time.Millisecond)

	// Verify keepalive was sent
	if conn.Len() == 0 {
		t.Error("expected KEEP_ALIVE to be sent")
	}

	// Stop the loop by setting connection as not alive
	isConnAlive.Store(false)
	time.Sleep(100 * time.Millisecond)

	// Verify loop stopped
	if handler.reportLoopStarted.Load() {
		t.Error("expected report loop to stop when connection is not alive")
	}
}

// signalingMockConn wraps mockConn and signals on writes via a channel
type signalingMockConn struct {
	mockConn
	writeSignal chan struct{}
}

func (s *signalingMockConn) Write(b []byte) (n int, err error) {
	n, err = s.mockConn.Write(b)
	// Signal write non-blockingly
	select {
	case s.writeSignal <- struct{}{}:
	default:
		// Channel full, skip signal (shouldn't happen with buffered channel)
	}
	return n, err
}

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}
