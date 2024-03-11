package watchdog

import (
	"Heis/Assigner"
	"Heis/Network/peers"
	"Heis/config"
	"fmt"
	"time"
)

func Watchdog(elevator *config.Elevator, peers chan peers.PeerUpdate, elevatorsMap map[string]config.Elevator, orderChanTx chan *config.AssignmentResults, ackChanRx chan string) {

	var lostElevatorsStates map[string]config.Elevator = make(map[string]config.Elevator)

	for {
		select {
		case peersUpdate := <-peers:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peersUpdate.Peers)
			fmt.Printf("  New:      %q\n", peersUpdate.New)
			fmt.Printf("  Lost:     %q\n", peersUpdate.Lost)
			if len(peersUpdate.Peers) != 0 {
				elevator.IsOnline = true

				if len(peersUpdate.Lost) != 0 {
					addToLostElevatorsMap(peersUpdate, elevatorsMap, lostElevatorsStates)
					transferOrders(elevator, peersUpdate, lostElevatorsStates)

					if contains(peersUpdate.Peers, elevator.Id) {
						elevatorsMap[elevator.Id] = *elevator
					}
					lostElevatorsStates = make(map[string]config.Elevator)
					if elevator.Id == peersUpdate.Peers[0] {
						Assigner.AssignHallOrders(orderChanTx, elevatorsMap, ackChanRx)
					}
				}
			} else {
				elevator.IsOnline = false
			}

		}
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
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
	for _, lostElev := range lostElevatorsStates {
		for floor := 0; floor < config.NumFloors; floor++ {
			for button := 0; button < 2; button++ {
				if elevator.Id == peersUpdate.Peers[0] {
					if lostElev.Requests[floor][button] {
						elevator.Requests[floor][button] = true

					}

				}
				for _, id := range peersUpdate.Lost {
					if elevator.Id == id {
						if lostElev.Requests[floor][button] {
							elevator.Requests[floor][button] = false

						}
					}
				}

			}
		}
	}
}

func SendElevatorStates(stateTx chan *config.Elevator, elevator *config.Elevator) {
	for {
		stateTx <- elevator
		time.Sleep(200 * time.Millisecond)
	}
}
