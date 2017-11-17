package main

import (
	"reflect"
	"testing"
)

func TestGetFilterValue(t *testing.T) {
	type args struct {
		fv string
	}
	tests := []struct {
		name       string
		args       args
		wantFilter []rune
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFilter := GetFilterValue(tt.args.fv); !reflect.DeepEqual(gotFilter, tt.wantFilter) {
				t.Errorf("GetFilterValue() = %v, want %v", gotFilter, tt.wantFilter)
			}
		})
	}
}

func TestGetItemReference(t *testing.T) {
	type args struct {
		ir      string
		cpSizes []int
	}
	tests := []struct {
		name              string
		args              args
		wantItemReference []rune
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotItemReference := GetItemReference(tt.args.ir, tt.args.cpSizes); !reflect.DeepEqual(gotItemReference, tt.wantItemReference) {
				t.Errorf("GetItemReference() = %v, want %v", gotItemReference, tt.wantItemReference)
			}
		})
	}
}

func TestGetPartitionAndCompanyPrefix(t *testing.T) {
	type args struct {
		cp string
	}
	tests := []struct {
		name              string
		args              args
		wantPartition     []rune
		wantCompanyPrefix []rune
		wantCpSizes       []int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPartition, gotCompanyPrefix, gotCpSizes := GetPartitionAndCompanyPrefix(tt.args.cp)
			if !reflect.DeepEqual(gotPartition, tt.wantPartition) {
				t.Errorf("GetPartitionAndCompanyPrefix() gotPartition = %v, want %v", gotPartition, tt.wantPartition)
			}
			if !reflect.DeepEqual(gotCompanyPrefix, tt.wantCompanyPrefix) {
				t.Errorf("GetPartitionAndCompanyPrefix() gotCompanyPrefix = %v, want %v", gotCompanyPrefix, tt.wantCompanyPrefix)
			}
			if !reflect.DeepEqual(gotCpSizes, tt.wantCpSizes) {
				t.Errorf("GetPartitionAndCompanyPrefix() gotCpSizes = %v, want %v", gotCpSizes, tt.wantCpSizes)
			}
		})
	}
}

func TestGetSerial(t *testing.T) {
	type args struct {
		s            string
		serialLength int
	}
	tests := []struct {
		name       string
		args       args
		wantSerial []rune
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSerial := GetSerial(tt.args.s, tt.args.serialLength); !reflect.DeepEqual(gotSerial, tt.wantSerial) {
				t.Errorf("GetSerial() = %v, want %v", gotSerial, tt.wantSerial)
			}
		})
	}
}

func TestMakeRuneSliceOfGIAI96(t *testing.T) {
	type args struct {
		cp string
		fv string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeRuneSliceOfGIAI96(tt.args.cp, tt.args.fv)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeRuneSliceOfGIAI96() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeRuneSliceOfGIAI96() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeRuneSliceOfGRAI96(t *testing.T) {
	type args struct {
		cp string
		fv string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeRuneSliceOfGRAI96(tt.args.cp, tt.args.fv)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeRuneSliceOfGRAI96() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeRuneSliceOfGRAI96() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeRuneSliceOfSGTIN96(t *testing.T) {
	type args struct {
		cp string
		fv string
		ir string
		s  string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeRuneSliceOfSGTIN96(tt.args.cp, tt.args.fv, tt.args.ir, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeRuneSliceOfSGTIN96() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeRuneSliceOfSGTIN96() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeRuneSliceOfSSCC96(t *testing.T) {
	type args struct {
		cp string
		fv string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeRuneSliceOfSSCC96(tt.args.cp, tt.args.fv)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeRuneSliceOfSSCC96() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeRuneSliceOfSSCC96() = %v, want %v", got, tt.want)
			}
		})
	}
}
