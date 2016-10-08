package main

import (
	"encoding/json"
	"log"

	"github.com/fatih/structs"
	"golang.org/x/net/websocket"
)

// WebsocketMessage to unmarshal JSON message from web clients
type WebsocketMessage struct {
	UpdateType string
	Tag        TagInString
	Tags       []map[string]interface{}
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
		m := WebsocketMessage{}
		if err = json.Unmarshal(clientMessage, &m); err != nil {
			log.Println(err.Error())
		}

		// Handle the command
		// Compose result struct containing proper parameters
		// TODO: separate actions into functions
		switch m.UpdateType {
		case "add":
			tag, err := buildTag([]string{m.Tag.PCBits, m.Tag.Length, m.Tag.EPCLengthBits, m.Tag.EPC, m.Tag.ReadData})
			check(err)
			add := &TagManager{
				action: AddTags,
				tags:   []*Tag{&tag}}
			tagManager <- add
			if add = <-tagManager; len(add.tags) != 0 {
				log.Println(m)
			} else {
				log.Println("failed", m)
				m.UpdateType = "error"
			}
		case "delete":
			tag, err := buildTag([]string{m.Tag.PCBits, m.Tag.Length, m.Tag.EPCLengthBits, m.Tag.EPC, m.Tag.ReadData})
			check(err)
			delete := &TagManager{
				action: DeleteTags,
				tags:   []*Tag{&tag}}
			tagManager <- delete
			if delete = <-tagManager; len(delete.tags) != 0 {
				log.Println(m)
			} else {
				log.Println("failed", m)
				m.UpdateType = "error"
			}
		case "retrieve":
			retrieve := &TagManager{
				action: RetrieveTags,
				tags:   []*Tag{}}
			tagManager <- retrieve
			retrieve = <-tagManager
			var tagList []map[string]interface{}
			for _, tag := range retrieve.tags {
				t := structs.Map(tag.InString())
				tagList = append(tagList, t)
			}
			m = WebsocketMessage{
				UpdateType: "retrieval",
				Tag:        TagInString{},
				Tags:       tagList}
		default:
			log.Println("Unknown UpdateType:", m.UpdateType)
		}

		clientMessage, err = json.Marshal(m)
		check(err)
		for cs := range activeClients {
			if err = websocket.Message.Send(cs.websocket, string(clientMessage)); err != nil {
				// we could not send the message to a peer
				log.Println("Could not send message to ", cs.clientIP, err.Error())
			}
		}
	}
}
