package watchdog

import (
	"Heis/config"
	"Heis/network/peers"
	"fmt"
)

func HandleLostPeers(elevator *config.Elevator, peers chan peers.PeerUpdate, elevatorsMap map[string]config.Elevator) {

	//kan hende at denne på initialiseres i main
	var lostElevatorsState map[string]config.Elevator = make(map[string]config.Elevator)

	for {
		select {
		case peersUpdate := <-peers:
			if len(peersUpdate.Lost) != 0 {
				for _, lostPeerID := range peersUpdate.Lost {
					lostElevatorsState[lostPeerID] = elevatorsMap[lostPeerID]

					delete(elevatorsMap, lostPeerID)
					fmt.Printf("Heis med ID %s er tapt. Tilstanden er lagret og heisen er fjernet fra elevatorsMap.\n", lostPeerID)
				}
				for _, halls := range lostElevatorsState {
					for floor := 0; floor < config.NumFloors; floor++ {
						for button := 0; button < 2; button++ {
							if elevator.Id == peersUpdate.Peers[0] {
								if halls.Requests[floor][button] == 1 {
									elevator.Requests[floor][button] = 1
								}

							}
							//må se på tilfelle der det er flere som er koblet fra om det blir riktig å gi den ene heisen alle orders
							//enten gå gjennom og bare gi den første heisen på nettet ordrene, eller kjøre den inn i kostfunksjonen på en eller annen måte
							//hvis man sender den inn i kostfunksjonen, hvordan skal den klare å ta disse ordrene men allikevel klare å dele ut til bare de to aktive heisen?
							//gi den til heis 1
							//når alle ordrene til den tapte heisen er overført, kan man slette innholdet i lostElevatorState
						}
					}
				}
				//Overskrive et tomt map på lostPeersmapet
			elevatorsMap[elevator.Id] = *elevator
				lostElevatorsState = make(map[string]config.Elevator)
			}

		}
	}
}
