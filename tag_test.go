package main

import (
	"bytes"
	"io/ioutil"
	"reflect"
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
		if !tag.IsEqual(tt.out) {
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
			tag.IsEqual(tt.out[i])
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
		param := buildTagReportDataParameter(&tag)
		if !bytes.Equal(param, tt.out) {
			t.Errorf("%v => %v, want %v", tt.in, param, tt.out)
		}
	}
}

func TestBuildTagReportDataStack(t *testing.T) {
	csvIn, _ := ioutil.ReadFile("tags.csv")
	tags := loadTagsFromCSV(string(csvIn))
	trds := buildTagReportDataStack(tags)
	t.Logf("TotalTagCounts: %v\n", trds.TotalTagCounts())
	t.Logf("TotalTagReportData: %v\n", len(trds.Stack))
}

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

func TestTagReportDataStack_TotalTagCounts(t *testing.T) {
	type fields struct {
		Stack []*TagReportData
	}
	tests := []struct {
		name   string
		fields fields
		want   uint
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trds := TagReportDataStack{
				Stack: tt.fields.Stack,
			}
			if got := trds.TotalTagCounts(); got != tt.want {
				t.Errorf("TagReportDataStack.TotalTagCounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTag(t *testing.T) {
	type args struct {
		record []string
	}
	tests := []struct {
		name    string
		args    args
		want    Tag
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTag(tt.args.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadTagsFromCSV(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want []*Tag
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loadTagsFromCSV(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadTagsFromCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTagReportDataParameter(t *testing.T) {
	type args struct {
		tag *Tag
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTagReportDataParameter(tt.args.tag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTagReportDataParameter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTagReportDataStack(t *testing.T) {
	type args struct {
		tags []*Tag
	}
	tests := []struct {
		name string
		args args
		want *TagReportDataStack
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTagReportDataStack(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTagReportDataStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIndexOfTag(t *testing.T) {
	type args struct {
		tags []*Tag
		t    *Tag
	}
	tests := []struct {
		name string
		args args
		want int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIndexOfTag(tt.args.tags, tt.args.t); got != tt.want {
				t.Errorf("getIndexOfTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeTagsToCSV(t *testing.T) {
	type args struct {
		tags   []*Tag
		output string
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeTagsToCSV(tt.args.tags, tt.args.output)
		})
	}
}
