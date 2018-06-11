package main

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

type TagRecord struct {
	PCBits string
	EPC    string
}

// TagInString to represent Tag struct all in string
type TagInString struct {
	PCBits string `json:"PCBits"`
	EPC    string `json:"EPC"`
}

// IsEqual to another Tag by taking one as its argument
// return true if they are the same
func (t Tag) IsEqual(tt Tag) bool {
	if t.PCBits == tt.PCBits && bytes.Equal(t.EPC, tt.EPC) {
		return true
	}
	return false
}

// IsDuplicate to test another Tag by comparing only EPC
// return true if the EPCs are the same
func (t Tag) IsDuplicate(tt Tag) bool {
	if bytes.Equal(t.EPC, tt.EPC) {
		return true
	}
	return false
}

// InString returns Tag structs in TagInString structs
func (t Tag) InString() *TagInString {
	return &TagInString{
		PCBits: strconv.FormatUint(uint64(t.PCBits), 16),
		EPC:    hex.EncodeToString(t.EPC),
	}
}

// MarshalBinary overwrites the marshaller in gob encoding *Tag
func (tp *Tag) MarshalBinary() (_ []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(tp.PCBits)
	enc.Encode(tp.EPC)
	return buf.Bytes(), err
}

// UnmarshalBinary overwrites the unmarshaller in gob decoding *PatriciaTrie
func (tp *Tag) UnmarshalBinary(data []byte) (err error) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(&tp.PCBits); err != nil {
		return
	}
	if err = dec.Decode(&tp.EPC); err != nil {
		return
	}
	return
}

// TagReportData holds an actual parameter in byte and
// how many tags are included in the parameter
type TagReportData struct {
	Parameter []byte
	TagCount  uint
}

// TagReportDataStack is a stack of TagReportData
type TagReportDataStack struct {
	Stack []*TagReportData
}

// TotalTagCounts returns how many tags are included in the TagReportDataStack
func (trds TagReportDataStack) TotalTagCounts() uint {
	ttc := uint(0)
	for _, trd := range trds.Stack {
		ttc += trd.TagCount
	}
	return ttc
}

func makeByteID(s string) ([]byte, error) {
	id, err := binutil.ParseBinRuneSliceToUint8Slice([]rune(s))
	return binutil.Pack([]interface{}{id}), err
}

// Construct Tag struct from a TagRecord
func buildTag(tagRecord *TagRecord) (Tag, error) {
	// PCbits
	pc64, err := strconv.ParseUint(tagRecord.PCBits, 16, 16)
	if err != nil {
		return Tag{}, err
	}
	pc := uint16(pc64)

	// EPC
	epc, err := makeByteID(tagRecord.EPC)
	if err != nil {
		return Tag{}, err
	}

	tag := Tag{pc, epc}
	return tag, nil
}

// Read Tag data from the CSV strings and returns a slice of Tag struct pointers
func loadTagsFromCSV(inputFile string) *[]*Tag {
	// Check inputFile
	fp, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	// Read CSV and store in []*Tag
	tags := []*Tag{}
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
			tag, err := buildTag(tagRecord)
			if err != nil {
				continue
			}
			tags = append(tags, &tag)
		}
	}

	return &tags
}

// Take one Tag struct and build TagReportData parameter payload in []byte
func buildTagReportDataParameter(tag *Tag) []byte {
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

func buildTagReportDataStack(tags []*Tag) *TagReportDataStack {
	var param []byte
	var trd *TagReportData
	var trds TagReportDataStack
	p := &trds // pointer to trds
	si := 0    // stack count

	// Iterate through tags and divide them into TRD stacks
	for _, tag := range tags {
		// When exceeds maxTag per TRD, append another TRD in the stack
		// 100 bytes for the offset for IP frame and ROAR headers
		param = buildTagReportDataParameter(tag)
		if len(p.Stack) != 0 && len(p.Stack[si].Parameter)+len(param)+100 > *pdu {
			trd = &TagReportData{Parameter: param, TagCount: 1}
			p.Stack = append(p.Stack, trd)
			si++
		} else {
			if len(p.Stack) == 0 {
				// First TRD
				trd = &TagReportData{Parameter: param, TagCount: 1}
				p.Stack = []*TagReportData{trd}
			} else {
				// Append TRD to an existing TRD
				p.Stack[si].Parameter = append(p.Stack[si].Parameter, param...)
				p.Stack[si].TagCount++
			}
		}
	}
	return p
}

func getIndexOfTag(tags []*Tag, t *Tag) int {
	index := 0
	for _, tag := range tags {
		if tag.IsDuplicate(*t) {
			return index
		}
		index++
	}
	return -1
}

/*
func writeTagsToCSV(tags []*Tag, output string) {
	file, err := os.Create(output)
	check(err)

	w := csv.NewWriter(file)
	for _, tag := range tags {
		record := []string{strconv.FormatUint(uint64(tag.PCBits), 16), strconv.FormatUint(uint64(tag.Length), 10), strconv.FormatUint(uint64(tag.EPCLengthBits), 10), hex.EncodeToString(tag.EPC)}
		if err := w.Write(record); err != nil {
			logger.Criticalf("Writing record to csv: %v", err.Error())
		}
		w.Flush()
		if err := w.Error(); err != nil {
			logger.Errorf(err.Error())
		}
	}
	file.Close()
}
*/
