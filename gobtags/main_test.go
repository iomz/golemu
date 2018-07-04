// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/iomz/golemu"
)

func Test_makeByteID(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"0 255", args{"0000000011111111"}, []byte{0, 255}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeByteID(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeByteID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeByteID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeUint16PC(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    uint16
		wantErr bool
	}{
		{"12288", args{"3000"}, uint16(12288), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeUint16PC(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeUint16PC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeUint16PC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readTagsFromCSV(t *testing.T) {
	type args struct {
		inputFile string
	}
	tests := []struct {
		name string
		args args
		want golemu.Tags
	}{
		{
			"2 tags",
			args{os.Getenv("GOPATH") + "/src/github.com/iomz/golemu/testdata/tags.csv"},
			golemu.Tags{
				&golemu.Tag{8610, []byte{203, 84, 216, 81, 46, 49, 227, 24}},
				&golemu.Tag{12288, []byte{52, 112, 249, 106, 163, 0, 0, 0, 0, 0, 1, 54}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readTagsFromCSV(tt.args.inputFile)
			for i, tag := range *got {
				if ok := tag.IsEqual(tt.want[i]); !ok {
					t.Errorf("readTagsFromCSV() =\n %v, want \n%v", *tag, tt.want[i])
				}
			}
		})
	}
}
