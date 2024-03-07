package watchdog

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/network/peers"
	"Heis/statemachines"
	"fmt"
)

func WatchDog(elevator *config.Elevator, peers chan peers.PeerUpdate, elevatorsMap map[string]config.Elevator, orderChanTx chan *costfunc.AssignmentResults) {

	var lostElevatorsStates map[string]config.Elevator = make(map[string]config.Elevator)

	for {
		select {
		case peersUpdate := <-peers:
			if len(peersUpdate.Lost) != 0 {
				addToLostElevatorsMap(peersUpdate, elevatorsMap, lostElevatorsStates)
				transferOrders(elevator, peersUpdate, lostElevatorsStates)

				if contains(peersUpdate.Peers, elevator.Id) {
					elevatorsMap[elevator.Id] = *elevator
				}
				//Her har jeg tenkt at vi må oppdatere elevatorsmapet før det sendes i kostfunksjonen igjen fordå nå har jo
				lostElevatorsStates = make(map[string]config.Elevator) //Overskrive et tomt map på lostPeersmapet
				statemachines.AssignHallOrders(orderChanTx, elevatorsMap)

			}
			if len(peersUpdate.New) != 0 {
				if contains(peersUpdate.Peers, elevator.Id) {
					elevatorsMap[elevator.Id] = *elevator
				}
				statemachines.AssignHallOrders(orderChanTx, elevatorsMap)
			}

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
					if lostElev.Requests[floor][button] == 1 {
						elevator.Requests[floor][button] = 1
					}
				}
				//hvordan kan man sjekke om man selv er den heisen som har mistet internettforbindelsen

			}
		}
	}
}
