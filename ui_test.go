package main

import (
	"reflect"
	"testing"

	"golang.org/x/net/websocket"
)

func TestReqAddTag(t *testing.T) {
	type args struct {
		ut  string
		req []TagInString
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
			if got := ReqAddTag(tt.args.ut, tt.args.req); got != tt.want {
				t.Errorf("ReqAddTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReqDeleteTag(t *testing.T) {
	type args struct {
		ut  string
		req []TagInString
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
			if got := ReqDeleteTag(tt.args.ut, tt.args.req); got != tt.want {
				t.Errorf("ReqDeleteTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReqRetrieveTag(t *testing.T) {
	tests := []struct {
		name string
		want []map[string]interface{}
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReqRetrieveTag(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReqRetrieveTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBroadcast(t *testing.T) {
	type args struct {
		clientMessage []byte
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Broadcast(tt.args.clientMessage)
		})
	}
}

func TestSockServer(t *testing.T) {
	type args struct {
		ws *websocket.Conn
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SockServer(tt.args.ws)
		})
	}
}
