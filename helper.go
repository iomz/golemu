// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package golemu

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/iomz/go-llrp"
)

// Time
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

// DecapsulateROAccessReport takes IDs and PCs from the ROAccessReport in buf
func DecapsulateROAccessReport(roarLength uint32, buf []byte) int {
	count := 0
	defer timeTrack(time.Now(), fmt.Sprintf("unpacking %v bytes", len(buf)))
	trds := buf[4 : roarLength-6] // TRD stack
	trdLength := uint16(0)        // First TRD size
	offset := uint32(0)           // the start of TRD
	//logger.Debugf("len(trds): %v\n", len(trds))
	//for trdLength != 0 && int(offset) != len(trds) {
	for {
		if uint32(10+offset) < roarLength {
			trdLength = binary.BigEndian.Uint16(trds[offset+2 : offset+4])
		} else {
			break
		}
		var id, pc []byte
		if trds[offset+4] == 141 { // EPC-96
			id = trds[offset+5 : offset+17]
			if trds[offset+17] == 140 { // C1G2-PC parameter
				pc = trds[offset+18 : offset+20]
			}
			count++
			//logger.Debugf("EPC: %v, (%x)\n", id, pc)
		} else if binary.BigEndian.Uint16(trds[offset+4:offset+6]) == 241 { // EPCData
			epcDataLength := binary.BigEndian.Uint16(trds[offset+6 : offset+8])  // length
			epcLengthBits := binary.BigEndian.Uint16(trds[offset+8 : offset+10]) // EPCLengthBits
			epcLengthBytes := uint32(epcLengthBits / 8)
			/*
				// ID length in byte = Length - (6 + 10 + 16 + 16)/8
				//id = trds[offset+6 : offset+epcDataSize-6]
				// trim the last 1 byte if it's not a multiple of a word
				//id = id[0 : epcLengthBits/8]
			*/
			id = trds[offset+10 : offset+10+epcLengthBytes]
			if 4+epcDataLength < trdLength && trds[offset+10+epcLengthBytes] == 140 { // C1G2-PC parameter
				pc = trds[offset+10+epcLengthBytes+1 : offset+10+epcLengthBytes+3]
			}
			_ = id
			_ = pc
			count++
			//logger.Debugf("EPC: %v, (%x)\n", id, pc)
		}
		offset += uint32(trdLength) // move the offset at the end of this TRD
		//logger.Debugf("offset: %v, roarLength: %v\n", offset, roarLength)
		//logger.Debugf("trdLength: %v, len(trds): %v\n", trdLength, len(trds))
	}
	return count
}

// SendROAccessReport iterates through the Tags and write ROAccessReport message to the socket
func SendROAccessReport(conn net.Conn, trds *TagReportDataStack, messageID *uint32) error {
	perms := rand.Perm(len(trds.Stack))
	//buf := make([]byte, 512)
	for _, i := range perms {
		trd := trds.Stack[i]
		// Append TagReportData to ROAccessReport
		roar := llrp.ROAccessReport(trd.Parameter, *messageID)
		atomic.AddUint32(messageID, 1)
		runtime.Gosched()

		// Send
		_, err := conn.Write(roar)
		if err != nil {
			return err
		}
		//time.Sleep(time.Millisecond)
	}

	return nil
}
