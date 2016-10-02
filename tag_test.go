package main

import (
	"bytes"
	"testing"
)

var tagtests = []struct {
	in  []string
	out Tag
}{
	{[]string{"29a9", "16", "80", "dc20420c4c72cf4d76de", "a896"},
		Tag{PCBits: uint16(10665), Length: uint16(16), EPCLengthBits: uint16(80), EPC: []byte{220, 32, 66, 12, 76, 114, 207, 77, 118, 222}, ReadData: []byte{168, 150}}},
	{[]string{"3000", "18", "96", "302DB319A0000040000002B8", "a896"},
		Tag{PCBits: uint16(12288), Length: uint16(18), EPCLengthBits: uint16(96), EPC: []byte{48, 45, 179, 25, 160, 0, 0, 64, 0, 0, 2, 184}, ReadData: []byte{168, 150}}},
	{[]string{"41a2", "22", "128", "c4a301c70d36cb32920b1d31c2dc3482", "a896"},
		Tag{PCBits: uint16(16802), Length: uint16(22), EPCLengthBits: uint16(128), EPC: []byte{196, 163, 1, 199, 13, 54, 203, 50, 146, 11, 29, 49, 194, 220, 52, 130}, ReadData: []byte{168, 150}}},
}

func TestBuildTag(t *testing.T) {
	for _, tt := range tagtests {
		tag, err := buildTag(tt.in)
		check(err)
		if !tag.Equal(tt.out) {
			t.Errorf("%v => %v, want %v", tt.in, tag, tt.out)
		}
	}
}

var csvtests = []struct {
	in  string
	out []Tag
}{
	{`41a2,22,128,c4a301c70d36cb32920b1d31c2dc3482,a896
29a9,16,80,dc20420c4c72cf4d76de,a896
3000,18,96,302DB319A000004000000003,a896
`, []Tag{
		{16802, 22, 128, []byte{196, 163, 1, 199, 13, 54, 203, 50, 146, 11, 29, 49, 194, 220, 52, 130}, []byte{168, 150}},
		{10665, 16, 80, []byte{220, 32, 66, 12, 76, 114, 207, 77, 118, 222}, []byte{168, 150}},
		{12288, 18, 96, []byte{48, 45, 179, 25, 160, 0, 0, 64, 0, 0, 0, 3}, []byte{168, 150}},
	},
	},
}

func TestLoadTagsFromCSV(t *testing.T) {
	for _, tt := range csvtests {
		tags := loadTagsFromCSV(tt.in)
		for i, tag := range tags {
			tag.Equal(tt.out[i])
		}
	}
}

var trdtests = []struct {
	in  []string
	out []byte
}{
	{[]string{"29a9", "16", "80", "dc20420c4c72cf4d76de", "a896"}, []byte{0, 240, 0, 36, 0, 241, 0, 16, 0, 80, 220, 32, 66, 12, 76, 114, 207, 77, 118, 222, 134, 203, 140, 41, 169, 1, 93, 0, 11, 0, 0, 9, 0, 1, 168, 150}},
	{[]string{"3000", "18", "96", "302DB319A0000040000002B8", "a896"}, []byte{0, 240, 0, 33, 141, 48, 45, 179, 25, 160, 0, 0, 64, 0, 0, 2, 184, 134, 203, 140, 48, 0, 1, 93, 0, 11, 0, 0, 9, 0, 1, 168, 150}},
	{[]string{"41a2", "22", "128", "c4a301c70d36cb32920b1d31c2dc3482", "a896"}, []byte{0, 240, 0, 42, 0, 241, 0, 22, 0, 128, 196, 163, 1, 199, 13, 54, 203, 50, 146, 11, 29, 49, 194, 220, 52, 130, 134, 203, 140, 65, 162, 1, 93, 0, 11, 0, 0, 9, 0, 1, 168, 150}},
}

func TestBuildTagReportDataParameter(t *testing.T) {
	for _, tt := range trdtests {
		tag, err := buildTag(tt.in)
		check(err)
		trd := buildTagReportDataParameter(&tag)
		if !bytes.Equal(trd, tt.out) {
			t.Errorf("%v => %v, want %v", tt.in, trd, tt.out)
		}
	}
}

func TestBuildTagReportDataStack(t *testing.T) {
}

func TestGetIndexOfTag(t *testing.T) {
}
