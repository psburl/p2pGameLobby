package p2pcommunication

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func createJoinHandler(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		peerJoining := Peer{}
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&peerJoining)
		if err != nil {
			log.Printf("Error unmarshalling from: %v", err)
		}
		system.peerJoin <- peerJoining
		system.getCurrentPeers <- true
		encoder := json.NewEncoder(writer)
		encoder.Encode(<-system.currentPeers)
	}
}

func createMessageHandler(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		receivedMessage := P2PMessage{}
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&receivedMessage)
		if err != nil {
			log.Printf("Error unmarshalling from: %v", err)
		}
		system.receivedMsg <- receivedMessage
	}
}

func getKnownPeersHandler(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		var peers []Peer
		for _, peer := range system.Peers {
			peers = append(peers, peer)
		}
		json, _ := json.Marshal(peers)
		fmt.Fprint(writer, string(json))
	}
}

func disconnectionTest(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		disconnectedPeer := Peer{}
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&disconnectedPeer)
		URL := "http://" + disconnectedPeer.Address + "/ping"
		_, err = http.Get(URL)
		if err != nil {
			fmt.Fprintf(writer, "error")
			_, connected := system.Peers[disconnectedPeer.Address]
			if connected {
				system.ReceivePeerLeft(disconnectedPeer)
			}
		}
		fmt.Fprintf(writer, "ok")
	}
}

func ping(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "")
	}
}
