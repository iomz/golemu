package main

import (
	"bytes"
	"testing"
)

var tagtests = []struct {
	in  *TagRecord
	out Tag
}{
	{&TagRecord{"21a2", "1100101101010100110010110101100010101101111110011100111100001000"},
		Tag{PCBits: uint16(8610), EPC: []byte{203, 84, 203, 88, 173, 249, 207, 8}}},
	{&TagRecord{"29a2", "11001011010101001101001001011000100111001111000100010111000111000010000010000010"},
		Tag{PCBits: uint16(10658), EPC: []byte{203, 84, 210, 88, 156, 241, 23, 28, 32, 130}}},
	{&TagRecord{"29a9", "11011100001000100100000100010001010111000011000011000011000111001011000111011110"},
		Tag{PCBits: uint16(10665), EPC: []byte{220, 34, 65, 17, 92, 48, 195, 28, 177, 222}}},
	{&TagRecord{"3000", "001100000111000000111110101001110111100100000000000000010100000000000000000000000000000000000001"},
		Tag{PCBits: uint16(12288), EPC: []byte{48, 112, 62, 167, 121, 0, 1, 64, 0, 0, 0, 1}}},
	{&TagRecord{"3000", "001100010110110110110101101100101110010101000010001011000010011110000101000000000000000000000000"},
		Tag{PCBits: uint16(12288), EPC: []byte{49, 109, 181, 178, 229, 66, 44, 39, 133, 0, 0, 0}}},
	{&TagRecord{"31a2", "110010110101010011011001000111010011000110110011110101111001110011000101000010110000110101100000"},
		Tag{PCBits: uint16(12706), EPC: []byte{203, 84, 217, 29, 49, 179, 215, 156, 197, 11, 13, 96}}},
	{&TagRecord{"39a2", "1100101101010100110110100011110100101110000001010000010001001101110000101100101110001101101100011101011000001000"},
		Tag{PCBits: uint16(14754), EPC: []byte{203, 84, 218, 61, 46, 5, 4, 77, 194, 203, 141, 177, 214, 8}}},
}

func TestBuildTag(t *testing.T) {
	for _, tt := range tagtests {
		tag, err := buildTag(tt.in)
		check(err)
		if !tag.IsEqual(tt.out) {
			t.Errorf("%v => %v, want %v", tt.in, tag, tt.out)
		}
	}
}

/*
var csvtests = []struct {
	in  string
	out []Tag
}{
	{`41a2,22,128,c4a301c70d36cb32920b1d31c2dc3482
29a9,16,80,dc20420c4c72cf4d76de
3000,18,96,302DB319A000004000000003
`, []Tag{
		{16802, 22, 128, []byte{196, 163, 1, 199, 13, 54, 203, 50, 146, 11, 29, 49, 194, 220, 52, 130}},
		{10665, 16, 80, []byte{220, 32, 66, 12, 76, 114, 207, 77, 118, 222}},
		{12288, 18, 96, []byte{48, 45, 179, 25, 160, 0, 0, 64, 0, 0, 0, 3}},
	},
	},
}

func TestLoadTagsFromCSV(t *testing.T) {
	for _, tt := range csvtests {
		tags := loadTagsFromCSV(tt.in)
		for i, tag := range tags {
			tag.IsEqual(tt.out[i])
		}
	}
}
*/

var trdtests = []struct {
	in  *TagRecord
	out []byte
}{}

func TestBuildTagReportDataParameter(t *testing.T) {
	for _, tt := range trdtests {
		tag, err := buildTag(tt.in)
		check(err)
		param := buildTagReportDataParameter(&tag)
		if !bytes.Equal(param, tt.out) {
			t.Errorf("%v => \n%v, want \n%v", tt.in, param, tt.out)
		}
	}
}

/*
func TestBuildTagReportDataStack(t *testing.T) {
	csvIn, _ := ioutil.ReadFile("tags.csv")
	tags := loadTagsFromCSV(string(csvIn))
	trds := buildTagReportDataStack(tags)
	t.Logf("TotalTagCounts: %v\n", trds.TotalTagCounts())
	t.Logf("TotalTagReportData: %v\n", len(trds.Stack))
}
*/

func TestGetIndexOfTag(t *testing.T) {
}

/*
func TestTag_IsEqual(t *testing.T) {
	type fields struct {
		PCBits        uint16
		Length        uint16
		EPCLengthBits uint16
		EPC           []byte
		ReadData      []byte
	}
	type args struct {
		tt Tag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t := Tag{
				PCBits:        tt.fields.PCBits,
				Length:        tt.fields.Length,
				EPCLengthBits: tt.fields.EPCLengthBits,
				EPC:           tt.fields.EPC,
				ReadData:      tt.fields.ReadData,
			}
			if got := t.IsEqual(tt.args.tt); got != tt.want {
				t.Errorf("Tag.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTag_IsDuplicate(t *testing.T) {
	type fields struct {
		PCBits        uint16
		Length        uint16
		EPCLengthBits uint16
		EPC           []byte
		ReadData      []byte
	}
	type args struct {
		tt Tag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t := Tag{
				PCBits:        tt.fields.PCBits,
				Length:        tt.fields.Length,
				EPCLengthBits: tt.fields.EPCLengthBits,
				EPC:           tt.fields.EPC,
				ReadData:      tt.fields.ReadData,
			}
			if got := t.IsDuplicate(tt.args.tt); got != tt.want {
				t.Errorf("Tag.IsDuplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTag_InString(t *testing.T) {
	type fields struct {
		PCBits        uint16
		Length        uint16
		EPCLengthBits uint16
		EPC           []byte
		ReadData      []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   *TagInString
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t := Tag{
				PCBits:        tt.fields.PCBits,
				Length:        tt.fields.Length,
				EPCLengthBits: tt.fields.EPCLengthBits,
				EPC:           tt.fields.EPC,
				ReadData:      tt.fields.ReadData,
			}
			if got := t.InString(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tag.InString() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
