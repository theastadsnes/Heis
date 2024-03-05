package statehandler

import (
	"Heis/config"
	"Heis/network/peers"
	"fmt"
	"time"
)

func Send(stateTx chan *config.Elevator, elevator *config.Elevator) {
	for {
		stateTx <- elevator
		time.Sleep(1 * time.Millisecond)
	}
}

func HandlePeerUpdates(peerUpdateCh <-chan peers.PeerUpdate, helloRx <-chan *config.Elevator) {
	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			//case received := <-helloRx:
			//fmt.Printf("Received:  %#v\n", received)
		}
	}

}
