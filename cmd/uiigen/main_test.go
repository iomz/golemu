package main

import "testing"

func TestCheckIfStringInSlice(t *testing.T) {
	type args struct {
		a    string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckIfStringInSlice(tt.args.a, tt.args.list); got != tt.want {
				t.Errorf("CheckIfStringInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeEPC(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeEPC(); got != tt.want {
				t.Errorf("MakeEPC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeISO(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeISO(); got != tt.want {
				t.Errorf("MakeISO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
