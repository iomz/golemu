// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package golemu

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/hex"
	"io"
	"os"
	"strconv"

	"github.com/iomz/go-llrp"
	"github.com/iomz/go-llrp/binutil"
)

// Tag holds a single virtual tag content
type Tag struct {
	PCBits uint16
	EPC    []byte
}

// BuildTagReportDataParameter takes one Tag struct and build TagReportData parameter payload in []byte
func (tag *Tag) BuildTagReportDataParameter() []byte {
	// EPCData
	// Calculate the right length fro, epc and pcbits
	epcLengthBits := len(tag.EPC) * 8 // # bytes * 8 = # bits
	length := 4 + 2 + len(tag.EPC)    // header + epcLengthBits + epc
	epcd := llrp.EPCData(uint16(length), uint16(epcLengthBits), tag.EPC)

	// AirProtocolTagData
	aptd := llrp.C1G2PC(tag.PCBits)

	// Merge them into TagReportData
	return llrp.TagReportData(epcd, aptd)
}

// MarshalBinary overwrites the marshaller in gob encoding *Tag
func (tag *Tag) MarshalBinary() (_ []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(tag.PCBits)
	enc.Encode(tag.EPC)
	return buf.Bytes(), err
}

// UnmarshalBinary overwrites the unmarshaller in gob decoding *Tag
func (tag *Tag) UnmarshalBinary(data []byte) (err error) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&tag.PCBits); err != nil {
		return
	}
	if err = dec.Decode(&tag.EPC); err != nil {
		return
	}
	return
}

// IsEqual to another Tag by taking one as its argument
// return true if they are the same
func (tag *Tag) IsEqual(tt *Tag) bool {
	if tag.PCBits == tt.PCBits && bytes.Equal(tag.EPC, tt.EPC) {
		return true
	}
	return false
}

// IsDuplicate to test another Tag by comparing only EPC
// return true if the EPCs are the same
func (tag *Tag) IsDuplicate(tt *Tag) bool {
	if bytes.Equal(tag.EPC, tt.EPC) {
		return true
	}
	return false
}

// ToTagRecord returns a pointer to TagRecord struct of the tag
func (tag *Tag) ToTagRecord() *TagRecord {
	return &TagRecord{
		PCBits: strconv.FormatUint(uint64(tag.PCBits), 16),
		EPC:    hex.EncodeToString(tag.EPC),
	}
}

// NewTag onstructs a Tag struct from a TagRecord
func NewTag(tagRecord *TagRecord) (*Tag, error) {
	// PCbits
	pc64, err := strconv.ParseUint(tagRecord.PCBits, 16, 16)
	if err != nil {
		return &Tag{}, err
	}
	pc := uint16(pc64)

	// EPC
	epc, err := makeByteID(tagRecord.EPC)
	if err != nil {
		return &Tag{}, err
	}

	return &Tag{pc, epc}, nil
}

func makeByteID(s string) ([]byte, error) {
	id, err := binutil.ParseBinRuneSliceToUint8Slice([]rune(s))
	return binutil.Pack([]interface{}{id}), err
}

// LoadTagsFromCSV reads Tag data from the CSV strings and returns a slice of Tag struct pointers
func LoadTagsFromCSV(inputFile string) Tags {
	// Check inputFile
	fp, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	// Read CSV and store in []*Tag
	var tags Tags
	reader := csv.NewReader(fp)
	reader.Comma = ','
	reader.LazyQuotes = true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		if len(record) == 2 {
			tagRecord := &TagRecord{record[0], record[1]} // PCbits, EPC
			// Construct a tag read data
			tag, err := NewTag(tagRecord)
			if err != nil {
				continue
			}
			tags = append(tags, tag)
		}
	}

	return tags
}
