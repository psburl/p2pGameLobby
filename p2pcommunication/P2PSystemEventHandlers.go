package p2pcommunication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func peerJoinEventHandler(system *P2PSystem, peer Peer) {
	if !system.knownPeer(peer) {
		fmt.Printf("# Connected to: %s \n", peer.Address)
		system.Peers[peer.Address] = peer
		go system.sendJoin(peer)
	}
}

func getCurrentPeersEventHandler(system *P2PSystem) {
	system.currentPeers <- system.Peers
}

func userMessageEventHandler(system *P2PSystem, message P2PMessage) {
	fmt.Printf("%s (self) says: %s\n", message.SourcePeer.Name, message.Message)
	for _, peer := range system.Peers {
		go system.sendMessage(peer, message)
	}
}

func receiveMessageEventHandler(system *P2PSystem, message P2PMessage) {
	fmt.Printf("%s writes: %s\n", message.SourcePeer.Name, message.Message)
}

func peerLeftEventHandler(system *P2PSystem, peer Peer) {
	go startElection(system, peer)
}

func startElection(system *P2PSystem, peer Peer) {

	if system.knownPeer(peer) == false {
		return
	}

	fmt.Println("start disconection test to peer " + peer.Address)
	peerJSON, _ := json.Marshal(peer)
	// starts with one 0 representing that current peer could not connect to destination
	neighboorsResponses := []int{0}

	for _, neighboor := range system.Peers {
		if neighboor.Address != system.Self.Address && neighboor.Address != peer.Address {

			url := "http://" + neighboor.Address + "/disconnectionTest"
			response, err := http.Post(url, "application/json", bytes.NewBuffer(peerJSON))

			if err != nil {
				neighboorsResponses = append(neighboorsResponses, 0)
				system.ReceivePeerLeft(neighboor)
			} else {
				defer response.Body.Close()
				var neighboorResponse string
				decoder := json.NewDecoder(response.Body)
				err = decoder.Decode(&neighboorResponse)
				if neighboorResponse == "ok" {
					neighboorsResponses = append(neighboorsResponses, 1)
				} else {
					neighboorsResponses = append(neighboorsResponses, 0)
				}
			}
		}
	}
	totalSuccessResponses := 0
	for _, responseByNeighboor := range neighboorsResponses {
		totalSuccessResponses = totalSuccessResponses + responseByNeighboor
	}
	if len(neighboorsResponses)*2.0/3.0 >= totalSuccessResponses {
		if system.knownPeer(peer) {
			delete(system.Peers, peer.Address)
			fmt.Println("Peer " + peer.Address + " diconnected")
		}
	}
}
