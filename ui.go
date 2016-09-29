package main

import (
	"fmt"
	"log"
	"strings"

	"code.google.com/p/go.net/websocket"
)

// WebsockConn holds connection consists of the websocket and the client ip
type WebsockConn struct {
	websocket *websocket.Conn
	clientIP  string
}

// sockServer to handle messaging between clients
func sockServer(ws *websocket.Conn) {
	var err error
	var clientMessage string
	// use []byte if websocket binary type is blob or arraybuffer
	// var clientMessage []byte

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	client := ws.Request().RemoteAddr
	log.Println("Client connected:", client)
	sockCli := WebsockConn{ws, client}
	activeClients[sockCli] = 0
	log.Println("Number of clients connected ...", len(activeClients))

	// for loop so the websocket stays open otherwise
	// it'll close after one Receieve and Send
	for {
		if err = message.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			// remove the ws client conn from our active clients
			delete(activeClients, sockCli)
			log.Println("Number of clients still connected ...", len(activeClients))
			return
		}

		clientMessage = sockCli.clientIP + " Said: " + clientMessage
		// Handle the command
		if strings.Contains(clientMessage, "add") {
			// add
			fmt.Println("command: add")
			tag, err := buildTag([]string{"10665", "16", "80", "dc20420c4c72cf4d76de"})
			check(err)
			add := &addOp{
				tag:  &tag,
				resp: make(chan bool)}
			adds <- add
			<-add.resp
		} else if strings.Contains(clientMessage, "delete") {
			// delete
			fmt.Println("command: delete")
			tag, err := buildTag([]string{"10665", "16", "80", "dc20420c4c72cf4d76de"})
			check(err)
			delete := &deleteOp{
				tag:  &tag,
				resp: make(chan bool)}
			deletes <- delete
			<-delete.resp
		} else {
			fmt.Println("command: something else")
		}

		for cs := range activeClients {
			if err = message.Send(cs.websocket, clientMessage); err != nil {
				// we could not send the message to a peer
				log.Println("Could not send message to ", cs.clientIP, err.Error())
			}
		}
	}
}

