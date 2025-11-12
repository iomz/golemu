//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package connection

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestLLRPHeaderSize(t *testing.T) {
	if LLRPHeaderSize != 10 {
		t.Errorf("expected LLRPHeaderSize to be 10, got %d", LLRPHeaderSize)
	}
}

func TestReadLLRPHeader(t *testing.T) {
	// Create test data
	header := uint16(0x1234)
	length := uint32(20)
	messageID := uint32(0x567890AB)

	// Create buffer with test data
	buf := make([]byte, 10)
	binary.BigEndian.PutUint16(buf[0:2], header)
	binary.BigEndian.PutUint32(buf[2:6], length)
	binary.BigEndian.PutUint32(buf[6:10], messageID)

	// Create a connection that reads from buffer
	conn := &mockConn{reader: bytes.NewReader(buf)}

	hdr, err := ReadLLRPHeader(conn)
	if err != nil {
		t.Fatalf("ReadLLRPHeader failed: %v", err)
	}

	if hdr.Header != header {
		t.Errorf("expected header %x, got %x", header, hdr.Header)
	}
	if hdr.Length != length {
		t.Errorf("expected length %d, got %d", length, hdr.Length)
	}
	if hdr.MessageID != messageID {
		t.Errorf("expected messageID %x, got %x", messageID, hdr.MessageID)
	}
}

func TestReadLLRPHeader_IncompleteData(t *testing.T) {
	// Create buffer with incomplete data (only 5 bytes)
	buf := make([]byte, 5)
	conn := &mockConn{reader: bytes.NewReader(buf)}

	_, err := ReadLLRPHeader(conn)
	if err == nil {
		t.Error("expected error for incomplete header data")
	}
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Errorf("expected EOF or UnexpectedEOF, got %v", err)
	}
}

func TestReadLLRPMessage_WithBody(t *testing.T) {
	header := uint16(0x1234)
	length := uint32(15) // 10 bytes header + 5 bytes body
	messageID := uint32(0x567890AB)
	body := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// Create buffer with header and body
	buf := make([]byte, 15)
	binary.BigEndian.PutUint16(buf[0:2], header)
	binary.BigEndian.PutUint32(buf[2:6], length)
	binary.BigEndian.PutUint32(buf[6:10], messageID)
	copy(buf[10:15], body)

	conn := &mockConn{reader: bytes.NewReader(buf)}

	hdr, msgBody, err := ReadLLRPMessage(conn)
	if err != nil {
		t.Fatalf("ReadLLRPMessage failed: %v", err)
	}

	if hdr.Header != header {
		t.Errorf("expected header %x, got %x", header, hdr.Header)
	}
	if len(msgBody) != 5 {
		t.Errorf("expected body length 5, got %d", len(msgBody))
	}
	if !bytes.Equal(msgBody, body) {
		t.Errorf("expected body %v, got %v", body, msgBody)
	}
}

func TestReadLLRPMessage_NoBody(t *testing.T) {
	header := uint16(0x1234)
	length := uint32(10) // Only header, no body
	messageID := uint32(0x567890AB)

	buf := make([]byte, 10)
	binary.BigEndian.PutUint16(buf[0:2], header)
	binary.BigEndian.PutUint32(buf[2:6], length)
	binary.BigEndian.PutUint32(buf[6:10], messageID)

	conn := &mockConn{reader: bytes.NewReader(buf)}

	hdr, msgBody, err := ReadLLRPMessage(conn)
	if err != nil {
		t.Fatalf("ReadLLRPMessage failed: %v", err)
	}

	if hdr.Header != header {
		t.Errorf("expected header %x, got %x", header, hdr.Header)
	}
	if len(msgBody) != 0 {
		t.Errorf("expected empty body, got %d bytes", len(msgBody))
	}
}

func TestReadLLRPMessage_IncompleteBody(t *testing.T) {
	header := uint16(0x1234)
	length := uint32(15) // Says 15 bytes total, but we only provide 12
	messageID := uint32(0x567890AB)

	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], header)
	binary.BigEndian.PutUint32(buf[2:6], length)
	binary.BigEndian.PutUint32(buf[6:10], messageID)
	buf[10] = 0x01
	buf[11] = 0x02

	conn := &mockConn{reader: bytes.NewReader(buf)}

	_, _, err := ReadLLRPMessage(conn)
	if err == nil {
		t.Error("expected error for incomplete body data")
	}
}

// mockConn is a simple mock implementation of net.Conn for testing
// It is thread-safe for concurrent writes to support testing goroutines.
type mockConn struct {
	reader io.Reader
	writer io.Writer
	mu     sync.Mutex // Protects concurrent writes to the writer
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.reader == nil {
		return 0, io.EOF
	}
	return m.reader.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writer == nil {
		return len(b), nil
	}
	return m.writer.Write(b)
}

// Len returns the length of the underlying buffer if it's a *bytes.Buffer.
// This method is thread-safe and should be used instead of accessing the buffer directly.
func (m *mockConn) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if buf, ok := m.writer.(*bytes.Buffer); ok {
		return buf.Len()
	}
	return 0
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &mockAddr{}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &mockAddr{}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type mockAddr struct{}

func (m *mockAddr) Network() string { return "tcp" }
func (m *mockAddr) String() string  { return "127.0.0.1:1234" }
