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

	fmt.Println("start disconection test to peer " + peer.Address)
	peerJSON, _ := json.Marshal(peer)
	var list []int
	list = append(list, 0)
	for _, neighboor := range system.Peers {
		if neighboor.Address != system.Self.Address && neighboor.Address != peer.Address {
			url := "http://" + neighboor.Address + "/disconnectionTest"
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(peerJSON))
			if err != nil {
				list = append(list, 0)
			} else {
				defer resp.Body.Close()
				var strResp string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&strResp)
				if strResp == "ok" {
					list = append(list, 1)
				} else {
					list = append(list, 0)
				}
			}
		}
	}
	total := 0
	for _, el := range list {
		total = total + el
	}
	if len(list)*2.0/3.0 >= total {
		delete(system.Peers, peer.Address)
		fmt.Printf("Peer diconnected")
	}
}
