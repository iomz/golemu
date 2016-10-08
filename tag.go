package main

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/iomz/go-llrp"
)

// Tag holds a single virtual tag content
type Tag struct {
	PCBits        uint16
	Length        uint16
	EPCLengthBits uint16
	EPC           []byte
	ReadData      []byte
}

// TagInString to represent Tag struct all in string
type TagInString struct {
	PCBits        string
	Length        string
	EPCLengthBits string
	EPC           string
	ReadData      string
}

// Equal to another Tag by taking one as its argument
// return true if they are the same
func (t Tag) Equal(tt Tag) bool {
	if t.PCBits == tt.PCBits && t.Length == tt.Length && tt.EPCLengthBits == tt.EPCLengthBits && bytes.Equal(t.EPC, tt.EPC) && bytes.Equal(t.ReadData, tt.ReadData) {
		return true
	}
	return false
}

// InString returns Tag structs in TagInString structs
func (t Tag) InString() *TagInString {
	return &TagInString{
		PCBits:        strconv.FormatUint(uint64(t.PCBits), 16),
		Length:        strconv.FormatUint(uint64(t.Length), 10),
		EPCLengthBits: strconv.FormatUint(uint64(t.EPCLengthBits), 10),
		EPC:           hex.EncodeToString(t.EPC),
		ReadData:      hex.EncodeToString(t.ReadData)}
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

// Construct Tag struct from Tag info strings
// TODO: take map instead of []string
func buildTag(record []string) (Tag, error) {
	// If the row is incomplete
	if len(record) != 5 {
		var t Tag
		return t, io.EOF
	}

	pc64, err := strconv.ParseUint(record[0], 16, 16)
	check(err)
	pc := uint16(pc64)
	len64, err := strconv.ParseUint(record[1], 10, 16)
	check(err)
	len := uint16(len64)
	epclen64, err := strconv.ParseUint(record[2], 10, 16)
	check(err)
	epclen := uint16(epclen64)
	epc, err := hex.DecodeString(record[3])
	check(err)
	readData, err := hex.DecodeString(record[4])
	check(err)

	tag := Tag{pc, len, epclen, epc, readData}
	return tag, nil
}

// Read Tag data from the CSV strings and returns a slice of Tag struct pointers
func loadTagsFromCSV(input string) []*Tag {
	r := csv.NewReader(strings.NewReader(input))
	tags := []*Tag{}
	for {
		record, err := r.Read()
		// If reached at the end
		if err == io.EOF {
			break
		}
		check(err)

		// Construct a tag read data
		tag, err := buildTag(record)
		if err != nil {
			continue
		}
		tags = append(tags, &tag)
	}
	return tags
}

// Take one Tag struct and build TagReportData parameter payload in []byte
func buildTagReportDataParameter(tag *Tag) []byte {
	// EPCData
	epcd := llrp.EPCData(tag.Length, tag.EPCLengthBits, tag.EPC)

	// PeakRSSI
	prssi := llrp.PeakRSSI()

	// AirProtocolTagData
	aptd := llrp.C1G2PC(tag.PCBits)

	// OpSpecResult
	osr := llrp.C1G2ReadOpSpecResult(tag.ReadData)

	// Merge them into TagReportData
	return llrp.TagReportData(epcd, prssi, aptd, osr)
}

func buildTagReportDataStack(tags []*Tag) *TagReportDataStack {
	var param []byte
	var trd *TagReportData
	var trds TagReportDataStack
	p := &trds
	si := 0

	// Iterate through tags and divide them into TRD stacks
	for _, tag := range tags {
		if len(p.Stack) != 0 && int(p.Stack[si].TagCount+1) > *maxTag && *maxTag != 0 {
			// When exceeds maxTag per TRD, append another TRD in the stack
			param = buildTagReportDataParameter(tag)
			trd = &TagReportData{Parameter: param, TagCount: 1}
			p.Stack = append(p.Stack, trd)
			si++
		} else {
			param = buildTagReportDataParameter(tag)
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
		if tag.Equal(*t) {
			return index
		}
		index++
	}
	return -1
}

func writeTagsToCSV(tags []*Tag, output string) {
	file, err := os.Create(output)
	check(err)

	w := csv.NewWriter(file)
	for _, tag := range tags {
		record := []string{strconv.FormatUint(uint64(tag.PCBits), 16), strconv.FormatUint(uint64(tag.Length), 10), strconv.FormatUint(uint64(tag.EPCLengthBits), 10), hex.EncodeToString(tag.EPC), hex.EncodeToString(tag.ReadData)}
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
		w.Flush()
		if err := w.Error(); err != nil {
			log.Fatal(err)
		}
	}
	file.Close()
}
