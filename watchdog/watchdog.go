package watchdog

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/network/peers"
	"Heis/singleElev/elevio"
	"Heis/statemachines"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func WatchDogLostPeers(elevator *config.Elevator, peers chan peers.PeerUpdate, elevatorsMap map[string]config.Elevator, orderChanTx chan *costfunc.AssignmentResults, ackChanRx chan string) {

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
						statemachines.AssignHallOrders(orderChanTx, elevatorsMap, ackChanRx)
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
					if lostElev.Requests[floor][button] == 1 {
						elevator.Requests[floor][button] = 1

					}

				}
				for _, id := range peersUpdate.Lost {
					if elevator.Id == id {
						if lostElev.Requests[floor][button] == 1 {
							elevator.Requests[floor][button] = 0

						}
					}
				}

				//hvordan kan man sjekke om man selv er den heisen som har mistet internettforbindelsen

			}
		}
	}
}

func WriteToBackup(elevator *config.Elevator) {
	filename := "cabOrder.txt"
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	caborders := make([]int, config.NumFloors)

	for floors, _ := range elevator.Requests {
		caborders[floors] = elevator.Requests[floors][2]
	}

	cabordersString := strings.Trim(fmt.Sprint(caborders), "[]")
	_, err = f.WriteString(cabordersString)
	defer f.Close()
}

func ReadFromBackup(buttons chan elevio.ButtonEvent) {
	filename := "cabOrder.txt"
	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	caborders := make([]bool, 0)
	if err == nil {
		cabOrders := strings.Split(string(f), " ")
		for _, order := range cabOrders {
			result, _ := strconv.ParseBool(order)
			caborders = append(caborders, result)
		}
	}
	time.Sleep(20 * time.Millisecond)
	for floor, order := range caborders {
		if order {
			backupOrder := elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
			buttons <- backupOrder
			time.Sleep(20 * time.Millisecond)
		}
	}
}
