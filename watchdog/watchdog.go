package watchdog

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/network/peers"
	"fmt"
)

func Watchdog(elevator *config.Elevator, peers chan peers.PeerUpdate, elevatorsMap map[string]config.Elevator, orderChanTx chan *config.AssignmentResults, ackChanRx chan string) {
	firstActivePeer := 0
	var lostElevatorsStates map[string]config.Elevator = make(map[string]config.Elevator)

	for {
		select {
		case peersUpdate := <-peers:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peersUpdate.Peers)
			fmt.Printf("  New:      %q\n", peersUpdate.New)
			fmt.Printf("  Lost:     %q\n", peersUpdate.Lost)
			if elevatorInActivePeers(peersUpdate.Peers, elevator.Id) {
				elevator.IsOnline = true

				if len(peersUpdate.Lost) != 0 {
					addToLostElevatorsMap(peersUpdate, elevatorsMap, lostElevatorsStates)
					transferOrders(elevator, peersUpdate, lostElevatorsStates)

					elevatorsMap[elevator.Id] = *elevator

					lostElevatorsStates = make(map[string]config.Elevator)

					if elevator.Id == peersUpdate.Peers[firstActivePeer] {
						fmt.Println("trying to assign hall orders")
						assigner.AssignHallOrders(orderChanTx, elevatorsMap, ackChanRx)
					}
				}
			} else {
				elevator.IsOnline = false
			}

		}
	}
}

func elevatorInActivePeers(activePeers []string, elevatorId string) bool {
	for _, id := range activePeers {
		if id == elevatorId {
			return true
		}
	}
	return false
}

func addToLostElevatorsMap(peersUpdate peers.PeerUpdate, elevatorsMap, lostElevatorsStates map[string]config.Elevator) {
	for _, lostPeerID := range peersUpdate.Lost {
		lostElevatorsStates[lostPeerID] = elevatorsMap[lostPeerID]
		delete(elevatorsMap, lostPeerID)
	}
}

func transferOrders(elevator *config.Elevator, peersUpdate peers.PeerUpdate, lostElevatorsStates map[string]config.Elevator) {
	firstActivePeer := 0

	for _, lostElev := range lostElevatorsStates {
		for floor := 0; floor < config.NumFloors; floor++ {
			for button := 0; button < 2; button++ {
				if elevator.Id == peersUpdate.Peers[firstActivePeer] {
					if lostElev.Requests[floor][button] {
						elevator.Requests[floor][button] = true
					}
				} /*
					for _, id := range peersUpdate.Lost {
						if elevator.Id == id {
							if lostElev.Requests[floor][button] {
								elevator.Requests[floor][button] = false

							}
						}
					}*/

			}
		}
	}
}
