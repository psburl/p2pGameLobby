package p2pcommunication

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Peers is the map of Peers with a address as key
type Peers map[string]Peer

// P2PSystem contains the complete p2p system
type P2PSystem struct {
	Self            Peer
	Peers           Peers
	receivedMsg     chan (P2PMessage)
	peerJoin        chan (Peer)
	peerLeft        chan (Peer)
	currentPeers    chan (Peers)
	getCurrentPeers chan (bool)
	userMsg         chan (P2PMessage)
}

// NewP2PSystem initializes a new P2PSystem and return a *P2PSystem
func NewP2PSystem(self Peer) *P2PSystem {
	system := new(P2PSystem)
	system.Self = self
	system.Peers = make(Peers)
	system.peerJoin = make(chan (Peer))
	system.currentPeers = make(chan (Peers))
	system.getCurrentPeers = make(chan (bool))
	system.userMsg = make(chan (P2PMessage))
	system.receivedMsg = make(chan (P2PMessage))
	system.peerLeft = make(chan (Peer))
	return system
}

// Start starts p2pcommunication
func (system *P2PSystem) Start() {
	go system.startListenerSelectLoop()
	go system.startWebListener()
	fmt.Printf("# \"%s\" listening on %s \n", system.Self.Name, system.Self.Address)
}

// StartStdinListener peer listening
func (system *P2PSystem) StartStdinListener(sender Peer) {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, _ := reader.ReadString('\n')
		message := line[:len(line)-1]
		system.userMsg <- P2PMessage{message, sender}
	}
}

// ReceivePeerJoin Receives the join of new peer
func (system *P2PSystem) ReceivePeerJoin(peer Peer) {
	system.peerJoin <- peer

}

func (system *P2PSystem) startListenerSelectLoop() {
	for {
		select {
		case peer := <-system.peerJoin:
			if !system.knownPeer(peer) {
				fmt.Printf("# Connected to: %s \n", peer.Address)
				system.Peers[peer.Address] = peer
				go system.sendJoin(peer)
			}

		case <-system.getCurrentPeers:
			system.currentPeers <- system.Peers

		case peer := <-system.peerLeft:
			go system.startElection(peer)

		case messageMsg := <-system.receivedMsg:
			fmt.Printf("%s writes: %s\n", messageMsg.SourcePeer.Name, messageMsg.Message)

		case messageMsg := <-system.userMsg:
			fmt.Printf("%s (self) says: %s\n", messageMsg.SourcePeer.Name, messageMsg.Message)
			for _, peer := range system.Peers {
				go system.sendMessage(peer, messageMsg)
			}
		}
	}
}

func (system *P2PSystem) startElection(peer Peer) {

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

func (system *P2PSystem) sendJoin(peer Peer) {
	URL := "http://" + peer.Address + "/join"
	qs, _ := json.Marshal(system.Self)
	resp, err := http.Post(URL, "application/json", bytes.NewBuffer(qs))
	if err != nil {
		system.peerLeft <- peer
		return
	}

	system.peerJoin <- peer
	defer resp.Body.Close()
	otherPeers := Peers{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&otherPeers)
	for _, peer := range otherPeers {
		system.peerJoin <- peer
	}
}

func createJoinHandler(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		joiner := Peer{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&joiner)
		if err != nil {
			log.Printf("Error unmarshalling from: %v", err)
		}
		system.peerJoin <- joiner
		system.getCurrentPeers <- true
		enc := json.NewEncoder(w)
		enc.Encode(<-system.currentPeers)
	}
}

func createMessageHandler(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cm := P2PMessage{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&cm)
		if err != nil {
			log.Printf("Error unmarshalling from: %v", err)
		}
		system.receivedMsg <- cm
	}
}

func getKnownPeersHandler(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var list []Peer
		for _, peer := range system.Peers {
			list = append(list, peer)
		}
		json, _ := json.Marshal(list)
		fmt.Fprint(w, string(json))
	}
}

func disconnectionTest(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dcPeer := Peer{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&dcPeer)
		URL := "http://" + dcPeer.Address + "/ping"
		_, err = http.Get(URL)
		if err != nil {
			fmt.Fprintf(w, "error")
		}
		fmt.Fprintf(w, "ok")
	}
}

func ping(system *P2PSystem) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "")
	}
}

func (system *P2PSystem) knownPeer(peer Peer) bool {
	if peer.Address == system.Self.Address {
		return true
	}
	_, knownPeer := system.Peers[peer.Address]
	return knownPeer
}

func (system *P2PSystem) sendMessage(peer Peer, msg P2PMessage) {
	URL := "http://" + peer.Address + "/message"
	qs, _ := json.Marshal(msg)
	_, err := http.Post(URL, "application/json", bytes.NewBuffer(qs))
	if err != nil {
		system.peerLeft <- peer
		return
	}
}

func (system *P2PSystem) startWebListener() {
	http.HandleFunc("/message", createMessageHandler(system))
	http.HandleFunc("/join", createJoinHandler(system))
	http.HandleFunc("/peers", getKnownPeersHandler(system))
	http.HandleFunc("/ping", ping(system))
	http.HandleFunc("/disconnectionText", disconnectionTest(system))
	log.Fatal(http.ListenAndServe(system.Self.Address, nil))
}
