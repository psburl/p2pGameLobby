package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	p2pc "./p2pcommunication"
)

func main() {
	port := flag.String("p", "8000", "Listen on port number")
	name := flag.String("n", "anonymous", "Nickname")
	peer := flag.String("j", "", "Other peer to join")
	flag.Parse()

	myPeer := p2pc.Peer{}
	myPeer.Name = *name
	myPeer.Address = getLocalIpv4() + ":" + *port

	system := p2pc.NewP2PSystem(myPeer)
	system.Start()

	if len(*peer) != 0 {
		knownPeer := p2pc.Peer{"", resolveOtherPeerIpv4(*peer)}
		system.ReceivePeerJoin(knownPeer)
	}

	system.StartStdinListener(system.Self)
}

func resolveOtherPeerIpv4(address string) string {
	if strings.Contains(address, "localhost") {
		address = strings.Replace(address, "localhost", getLocalIpv4(), -1)
	}
	return address
}

func getLocalIpv4() string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return fmt.Sprintf("%s", ipv4)
		}
	}
	return "localhost"
}
