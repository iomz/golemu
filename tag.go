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
	pcBits        uint16
	length        uint16
	epcLengthBits uint16
	epc           []byte
	readData      []byte
}

type addOp struct {
	tag  *Tag
	resp chan bool
}

type deleteOp struct {
	tag  *Tag
	resp chan bool
}

// Equal to another Tag by taking one as its argument
// return true if they are the same
func (t Tag) Equal(tt Tag) bool {
	if t.pcBits == tt.pcBits && t.length == tt.length && tt.epcLengthBits == tt.epcLengthBits && bytes.Equal(t.epc, tt.epc) && bytes.Equal(t.readData, tt.readData) {
		return true
	}
	return false
}

// Construct Tag struct from Tag info strings
func buildTag(record []string) (Tag, error) {
	// If the row is incomplete
	if len(record) != 4 {
		var t Tag
		return t, io.EOF
	}

	pc64, err := strconv.ParseUint(record[0], 10, 16)
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
	readData, err := hex.DecodeString("a896")
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
	epcd := llrp.EPCData(tag.length, tag.epcLengthBits, tag.epc)

	// PeakRSSI
	prssi := llrp.PeakRSSI()

	// AirProtocolTagData
	aptd := llrp.C1G2PC(tag.pcBits)

	// OpSpecResult
	osr := llrp.C1G2ReadOpSpecResult(tag.readData)

	// Merge them into TagReportData
	trd := llrp.TagReportData(epcd, prssi, aptd, osr)

	return trd
}

func buildTagReportDataStack(tags []*Tag) []*[]byte {
	var trds []*[]byte
	tagCount := 0
	trdIndex := 0

	// Iterate through tags and divide them into TRD stacks
	for _, tag := range tags {
		tagCount++
		// TODO: Need to set ceiling for too large payload?
		if tagCount > *maxTag && *maxTag != 0 {
			trd := buildTagReportDataParameter(tag)
			trds = append(trds, &trd)
			trdIndex++
			tagCount = 1
		} else {
			trd := buildTagReportDataParameter(tag)
			if len(trds) == 0 {
				trds = append(trds, &trd)
			} else {
				*(trds[trdIndex]) = append(*(trds[trdIndex]), trd...)
			}
		}
	}

	return trds
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
		record := []string{strconv.FormatUint(uint64(tag.pcBits), 10), strconv.FormatUint(uint64(tag.length), 10), strconv.FormatUint(uint64(tag.epcLengthBits), 10), hex.EncodeToString(tag.epc)}
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
