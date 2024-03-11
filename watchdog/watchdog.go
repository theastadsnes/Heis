package watchdog

import (
	"Heis/Assigner"
	"Heis/config"
	"Heis/network/peers"
	"fmt"
	"time"
)

func WatchDogLostPeers(elevator *config.Elevator, peers chan peers.PeerUpdate, elevatorsMap map[string]config.Elevator, orderChanTx chan *Assigner.AssignmentResults, ackChanRx chan string) {

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
					//Her har jeg tenkt at vi må oppdatere elevatorsmapet før det sendes i kostfunksjonen igjen fordå nå har jo
					lostElevatorsStates = make(map[string]config.Elevator) //Overskrive et tomt map på lostPeersmapet
					if elevator.Id == peersUpdate.Peers[0] {
						Assigner.AssignHallOrders(orderChanTx, elevatorsMap, ackChanRx)
					}

				}
			} else {
				elevator.IsOnline = false
			}
			// if len(peersUpdate.New) != 0 {

			// 	if elevator.Id == peersUpdate.Peers[0] {
			// 		statemachines.AssignHallOrders(orderChanTx, elevatorsMap)
			// 	}
			// }

		}
	}
}

// Hjelpefunksjon for å sjekke om en liste med strenger inneholder en spesifikk streng/verdi
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
		fmt.Printf("Heis med ID %s er tapt. Tilstanden er lagret og heisen er fjernet fra elevatorsMap.\n", lostPeerID)
	}
}

func transferOrders(elevator *config.Elevator, peersUpdate peers.PeerUpdate, lostElevatorsStates map[string]config.Elevator) {
	for _, lostElev := range lostElevatorsStates {
		for floor := 0; floor < config.NumFloors; floor++ {
			for button := 0; button < 2; button++ {
				if elevator.Id == peersUpdate.Peers[0] { //Velger den første peeren som er online til å ta over ordrene
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

				//hvordan kan man sjekke om man selv er den heisen som har mistet internettforbindelsen

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
