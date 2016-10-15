package main

import (
	"encoding/json"

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

// ReqAddTag handles a tag addition request
func ReqAddTag(ut string, req []TagInString) string {
	var tags []*Tag
	for _, t := range req {
		tag, err := buildTag([]string{t.PCBits, t.Length, t.EPCLengthBits, t.EPC, t.ReadData})
		check(err)
		tags = append(tags, &tag)
	}
	add := &TagManager{
		action: AddTags,
		tags:   tags}
	tagManager <- add
	if add = <-tagManager; len(add.tags) != 0 {
		logger.Debugf("%v, %v", ut, req)
		return ut
	}
	logger.Errorf("failed %v, %v", ut, req)
	return "error"
}

// ReqDeleteTag handles a tag deletion request
func ReqDeleteTag(ut string, req []TagInString) string {
	var tags []*Tag
	for _, t := range req {
		tag, err := buildTag([]string{t.PCBits, t.Length, t.EPCLengthBits, t.EPC, t.ReadData})
		check(err)
		tags = append(tags, &tag)
	}
	delete := &TagManager{
		action: DeleteTags,
		tags:   tags}
	tagManager <- delete
	if delete = <-tagManager; len(delete.tags) != 0 {
		logger.Debugf("%v %v", ut, req)
		return ut
	}
	logger.Errorf("failed %v %v", ut, req)
	return "error"
}

// ReqRetrieveTag handles a tag retrieval request
func ReqRetrieveTag() []map[string]interface{} {
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
	return tagList
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
			logger.Warningf("Websocket could not be closed", err.Error())
		}
	}()

	client := ws.Request().RemoteAddr
	logger.Debugf("Client connected:", client)
	clientSock := WebsockConn{ws, client}
	activeClients[clientSock] = 0
	logger.Debugf("Number of clients connected ...", len(activeClients))

	// for loop so the websocket stays open otherwise
	// it'll close after one Receieve and Send
	for {
		if err = websocket.Message.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			logger.Errorf("Websocket Disconnected waiting %v", err.Error())
			// remove the ws client conn from our active clients
			delete(activeClients, clientSock)
			logger.Debugf("Number of clients still connected ... %v", len(activeClients))
			return
		}

		//clientMessage = clientSock.clientIP + " Said: " + clientMessage

		// Parse the JSON
		m := WebsocketMessage{}
		if err = json.Unmarshal(clientMessage, &m); err != nil {
			logger.Errorf(err.Error())
		}

		// Handle the command
		// Compose result struct containing proper parameters
		// TODO: separate actions into functions
		switch m.UpdateType {
		case "add":
			m.UpdateType = ReqAddTag(m.UpdateType, []TagInString{m.Tag})
		case "delete":
			m.UpdateType = ReqDeleteTag(m.UpdateType, []TagInString{m.Tag})
		case "retrieve":
			tagList := ReqRetrieveTag()
			m = WebsocketMessage{
				UpdateType: "retrieval",
				Tag:        TagInString{},
				Tags:       tagList}
		default:
			logger.Warningf("Unknown UpdateType: %v", m.UpdateType)
		}

		clientMessage, err = json.Marshal(m)
		check(err)
		for cs := range activeClients {
			if err = websocket.Message.Send(cs.websocket, string(clientMessage)); err != nil {
				// we could not send the message to a peer
				logger.Warningf("Could not send message to ", cs.clientIP, err.Error())
			}
		}
	}
}
