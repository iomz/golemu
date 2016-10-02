package main

import (
	"encoding/json"
	"log"

	"github.com/fatih/structs"
	"golang.org/x/net/websocket"
)

// WebsockMessage to unmarshal JSON message from web clients
type WebsockMessage struct {
	UpdateType string
	Tag        TagInString
}

// WebsockConn holds connection consists of the websocket and the client ip
type WebsockConn struct {
	websocket *websocket.Conn
	clientIP  string
}

// SockServer to handle messaging between clients
func SockServer(ws *websocket.Conn) {
	var err error
	//var clientMessage string
	// use []byte if websocket binary type is blob or arraybuffer
	var clientMessage []byte

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	client := ws.Request().RemoteAddr
	log.Println("Client connected:", client)
	clientSock := WebsockConn{ws, client}
	activeClients[clientSock] = 0
	log.Println("Number of clients connected ...", len(activeClients))

	// for loop so the websocket stays open otherwise
	// it'll close after one Receieve and Send
	for {
		if err = websocket.Message.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			// remove the ws client conn from our active clients
			delete(activeClients, clientSock)
			log.Println("Number of clients still connected ...", len(activeClients))
			return
		}

		//clientMessage = clientSock.clientIP + " Said: " + clientMessage

		// Parse the JSON
		m := WebsockMessage{}
		if err = json.Unmarshal(clientMessage, &m); err != nil {
			log.Println(err.Error())
		}

		// Handle the command
		// Compose result struct containing proper parameters
		// TODO: separate actions into functions
		result := false
		switch m.UpdateType {
		case "add":
			tag, err := buildTag([]string{m.Tag.PCBits, m.Tag.Length, m.Tag.EPCLengthBits, m.Tag.EPC, m.Tag.ReadData})
			check(err)
			add := &addOp{
				tag:  &tag,
				resp: make(chan bool)}
			adds <- add
			if result = <-add.resp; result {
				log.Println(m)
			} else {
				log.Println("failed", m)
			}
		case "delete":
			tag, err := buildTag([]string{m.Tag.PCBits, m.Tag.Length, m.Tag.EPCLengthBits, m.Tag.EPC, m.Tag.ReadData})
			check(err)
			delete := &deleteOp{
				tag:  &tag,
				resp: make(chan bool)}
			deletes <- delete
			if result = <-delete.resp; result {
				log.Println(m)
			} else {
				log.Println("failed", m)
			}
		case "retrieve":
			retrieve := &retrieveOp{
				tags: make(chan []*Tag)}
			retrieves <- retrieve
			tags := <-retrieve.tags
			var tagList []map[string]interface{}
			for _, tag := range tags {
				t := structs.Map(tag.InString())
				tagList = append(tagList, t)
			}
			clientMessage, err = json.Marshal(tagList)
			check(err)
		default:
			log.Println("Unknown UpdateType:", m.UpdateType)
		}

		for cs := range activeClients {
			if err = websocket.Message.Send(cs.websocket, string(clientMessage)); err != nil {
				// we could not send the message to a peer
				log.Println("Could not send message to ", cs.clientIP, err.Error())
			}
		}
	}
}
