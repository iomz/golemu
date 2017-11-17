package binutil

import (
	"math/big"
	"reflect"
	"testing"
)

func TestGenerateNLengthHexString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"n = 0", args{0}, ""},
		{"n = 1", args{1}, "a"},
		{"n = 2", args{2}, "ab"},
		{"n = 32", args{32}, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
		{"n = 64", args{64}, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateNLengthHexString(tt.args.n); len(got) != len(tt.want) {
				t.Errorf("GenerateNLengthHexString() = %v, want %v length", got, len(tt.want))
			}
		})
	}
}

func TestGenerateNLengthRandomBinRuneSlice(t *testing.T) {
	type args struct {
		n   int
		max uint
	}
	tests := []struct {
		name  string
		args  args
		want  []rune
		want1 uint
	}{
		{"n = 0", args{0, 0}, []rune(""), 0},
		{"n = 2", args{2, 0}, []rune("11"), 3},
		{"n = 64, max = 16", args{64, 16}, []rune("0000000000000000000000000000000000000000000000000000000000010000"), 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GenerateNLengthRandomBinRuneSlice(tt.args.n, tt.args.max)
			if len(got) != len(tt.want) {
				t.Errorf("GenerateNLengthHexString() = %v, want %v length", got, len(tt.want))
			}
			if got1 > tt.want1 {
				t.Errorf("GenerateNLengthRandomBinRuneSlice() got1 = %v, want less than %v", got1, tt.want1)
			}
		})
	}
}

func TestGenerateNLengthZeroPaddingRuneSlice(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want []rune
	}{
		{"n = 0", args{0}, []rune("")},
		{"n = 2", args{2}, []rune("00")},
		{"n = 32", args{32}, []rune("00000000000000000000000000000000")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateNLengthZeroPaddingRuneSlice(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateNLengthZeroPaddingRuneSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRandomInt(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
	}{
		{"min = 0, max = 100", args{0, 100}},
		{"min = 35, max = 40", args{35, 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateRandomInt(tt.args.min, tt.args.max)
			if got < tt.args.min {
				t.Errorf("GenerateRandomInt() = %v, want > %v", got, tt.args.min)
			} else if got > tt.args.max {
				t.Errorf("GenerateRandomInt() = %v, want < %v", got, tt.args.max)
			}
		})
	}
}

func TestPack(t *testing.T) {
	type args struct {
		data []interface{}
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
			if got := Pack(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBigIntToBinString(t *testing.T) {
	type args struct {
		cp *big.Int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseBigIntToBinString(tt.args.cp); got != tt.want {
				t.Errorf("ParseBigIntToBinString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBinRuneSliceToUint8Slice(t *testing.T) {
	type args struct {
		bs []rune
	}
	tests := []struct {
		name    string
		args    args
		want    []uint8
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBinRuneSliceToUint8Slice(tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBinRuneSliceToUint8Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBinRuneSliceToUint8Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDecimalStringToBinRuneSlice(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []rune
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseDecimalStringToBinRuneSlice(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDecimalStringToBinRuneSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHexStringToBinString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name          string
		args          args
		wantBinString string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBinString := ParseHexStringToBinString(tt.args.s); gotBinString != tt.wantBinString {
				t.Errorf("ParseHexStringToBinString() = %v, want %v", gotBinString, tt.wantBinString)
			}
		})
	}
}
