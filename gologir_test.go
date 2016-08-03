package main

import (
	"bytes"
	"errors"
	"testing"
)

var tagtests = []struct {
	in  []string
	out []byte
}{
	{[]string{"10665", "16", "80", "dc20420c4c72cf4d76de"}, []byte{0, 240, 0, 36, 0, 241, 0, 16, 0, 80, 220, 32, 66, 12, 76, 114, 207, 77, 118, 222, 134, 203, 140, 41, 169, 1, 93, 0, 11, 0, 0, 9, 0, 1, 168, 150}},
	{[]string{"12288", "18", "96", "302DB319A0000040000002B8"}, []byte{0, 240, 0, 33, 141, 48, 45, 179, 25, 160, 0, 0, 64, 0, 0, 2, 184, 134, 203, 140, 48, 0, 1, 93, 0, 11, 0, 0, 9, 0, 1, 168, 150}},
	{[]string{"16802", "22", "128", "c4a301c70d36cb32920b1d31c2dc3482"}, []byte{0, 240, 0, 42, 0, 241, 0, 22, 0, 128, 196, 163, 1, 199, 13, 54, 203, 50, 146, 11, 29, 49, 194, 220, 52, 130, 134, 203, 140, 65, 162, 1, 93, 0, 11, 0, 0, 9, 0, 1, 168, 150}},
}

func TestBuildTagReportDataParameter(t *testing.T) {
	for _, tt := range tagtests {
		tag, err := buildTag(tt.in)
		check(err)
		trd := buildTagReportDataParameter(&tag)
		if !bytes.Equal(trd, tt.out) {
			t.Errorf("%v => %v, want %v", tt.in, trd, tt.out)
		}
	}
}

var csvtests = []struct {
	in		string
	out   []Tag
}{
	{`16802,22,128,c4a301c70d36cb32920b1d31c2dc3482
10665,16,80,dc20420c4c72cf4d76de
12288,18,96,302DB319A000004000000003
`, []Tag{
		{16802,22,128,[]byte{196,163,1,199,13,54,203,50,146,11,29,49,194,220,52,130},[]byte{168,150}},
		{10665,16,80,[]byte{220,32,66,12,76,114,207,77,118,222},[]byte{168,150}},
		{12288,18,96,[]byte{48,45,179,25,160,0,0,64,0,0,0,3},[]byte{168,150}},
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

func TestCheck(t *testing.T) {
	e := errors.New("dummy error")
	check(nil)
	assertCheckPanic(t, check, e)
}

func assertCheckPanic(t *testing.T, f func(error), e error) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f(e)
}

