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
			peerJoinEventHandler(system, peer)
		case <-system.getCurrentPeers:
			getCurrentPeersEventHandler(system)
		case message := <-system.userMsg:
			userMessageEventHandler(system, message)
		case message := <-system.receivedMsg:
			receiveMessageEventHandler(system, message)
		case peer := <-system.peerLeft:
			peerLeftEventHandler(system, peer)
		}
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
