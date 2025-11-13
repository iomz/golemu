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
	// LLRPHeaderSize is the size of the LLRP message header in bytes.
	// The header consists of: 2 bytes (message type) + 4 bytes (message length) + 4 bytes (message ID).
	LLRPHeaderSize = 10
)

// LLRPHeader represents the header portion of an LLRP message.
// It contains the message type, total message length, and unique message identifier.
type LLRPHeader struct {
	Header    uint16 // Message type identifier
	Length    uint32 // Total message length including header
	MessageID uint32 // Unique message identifier
}

// ReadLLRPHeader reads and parses an LLRP message header from the connection.
// It reads exactly 10 bytes and decodes them according to the LLRP protocol specification.
//
// Returns the parsed header or an error if the header cannot be read.
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

// ReadLLRPMessage reads a complete LLRP message including header and body from the connection.
// It validates the message length to prevent malicious or malformed packets, then reads
// the message body based on the length specified in the header.
//
// Returns the parsed header, message body bytes, and an error if the message cannot be read
// or if the message length is invalid.
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
	messageSize := int64(hdr.Length) - int64(LLRPHeaderSize)
	if messageSize > int64(int(^uint(0)>>1)) {
		return nil, nil, fmt.Errorf("invalid LLRP body length: %d exceeds host capacity", hdr.Length)
	}
	if messageSize > 0 {
		messageValue = make([]byte, int(messageSize))
		if _, err = io.ReadFull(conn, messageValue); err != nil {
			return nil, nil, err
		}
	}

	return hdr, messageValue, nil
}
