/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/network/bcast"
	"Heis/network/localip"
	"Heis/network/peers"
	"Heis/network/statehandler"
	"Heis/singleElev/elevio"
	"Heis/singleElev/fsm"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	//Define the number of floors in the building
	numFloors := 4

	// Initialize elevator I/O
	elevio.Init("localhost:15000", numFloors)

	// Create channels for elevator I/O events
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Create a timer for the door
	doorTimer := time.NewTimer(time.Duration(3) * time.Second)

	// Start polling for elevator I/O events
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	//Når vi skal initialisere et elevator objekt bruker vi LocalIP() til å finne IP adressen til den pcen koden kjøres fra. Denne vil bli lagt til i structen.
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("%s-%d", localIP, os.Getpid())
	}

	elevator := config.InitElevState(id)
	// Start the finite state machine for elevator control
	go fsm.Fsm(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors)

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate) //Kanal som sender/mottar uppdateringer i Peer structen, om det er noen som kobles fra nettet eller noen nye tilkoblinger
	peerTxEnable := make(chan bool)             //Kanal som kan brukes for å vise at man ikke er tilgjengelig, selvom man kanskje er på nettet
	stateTx := make(chan *config.Elevator)      //Gjøre denne om til å sende Elevator state
	stateRx := make(chan *config.Elevator)
	cabOrder := make(chan *elevio.ButtonEvent)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Transmitter(16569, stateTx)
	go bcast.Receiver(16569, stateRx)
	go statehandler.HandlePeerUpdates(peerUpdateCh, stateRx)
	go statehandler.Send(stateTx, &elevator)
	go assigner.Assigner(stateRx, drv_buttons, cabOrder)

	select {}

}
