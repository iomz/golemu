//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package connection

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	// LLRPHeaderSize is the size of the LLRP message header (2 bytes header + 4 bytes length + 4 bytes messageID)
	LLRPHeaderSize = 10
)

// LLRPHeader represents an LLRP message header
type LLRPHeader struct {
	Header    uint16
	Length    uint32
	MessageID uint32
}

// ReadLLRPHeader reads the LLRP header from a connection
func ReadLLRPHeader(conn net.Conn) (*LLRPHeader, error) {
	header := make([]byte, 2)
	length := make([]byte, 4)
	messageID := make([]byte, 4)

	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(conn, length); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(conn, messageID); err != nil {
		return nil, err
	}

	return &LLRPHeader{
		Header:    binary.BigEndian.Uint16(header),
		Length:    binary.BigEndian.Uint32(length),
		MessageID: binary.BigEndian.Uint32(messageID),
	}, nil
}

// ReadLLRPMessage reads a complete LLRP message including header and body
func ReadLLRPMessage(conn net.Conn) (*LLRPHeader, []byte, error) {
	hdr, err := ReadLLRPHeader(conn)
	if err != nil {
		return nil, nil, err
	}

	// guard against malicious or malformed LLRP packets
	if hdr.Length < LLRPHeaderSize {
		return nil, nil, fmt.Errorf("invalid LLRP message length: %d (must be at least %d)", hdr.Length, LLRPHeaderSize)
	}

	var messageValue []byte
	messageSize := hdr.Length - LLRPHeaderSize
	if messageSize > 0 {
		messageValue = make([]byte, messageSize)
		if _, err = io.ReadFull(conn, messageValue); err != nil {
			return nil, nil, err
		}
	}

	return hdr, messageValue, nil
}
