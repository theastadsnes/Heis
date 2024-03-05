/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/costfunc"
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
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("%s-%d", localIP, os.Getpid())
	}

	numFloors := 4
	elevator := config.InitElevState(id)
	// Initialize elevator I/O
	elevio.Init("localhost:15657", numFloors)

	// Create channels for elevator I/O events
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	buttons2 := make(chan elevio.ButtonEvent)
	// Create a timer for the door
	doorTimer := time.NewTimer(time.Duration(3) * time.Second)

	peerUpdateCh := make(chan peers.PeerUpdate) //Kanal som sender/mottar uppdateringer i Peer structen, om det er noen som kobles fra nettet eller noen nye tilkoblinger
	peerTxEnable := make(chan bool)             //Kanal som kan brukes for å vise at man ikke er tilgjengelig, selvom man kanskje er på nettet
	stateTx := make(chan *config.Elevator)      //Gjøre denne om til å sende Elevator state
	stateRx := make(chan *config.Elevator)
	cabOrder := make(chan *elevio.ButtonEvent)
	orderChanTx := make(chan *costfunc.AssignmentResults)
	orderChanRx := make(chan *costfunc.AssignmentResults)

	// Start polling for elevator I/O events
	go elevio.PollButtons(drv_buttons)
	go elevio.PollButtons(buttons2)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Start the finite state machine for elevator control
	go fsm.Fsm(&elevator, buttons2, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors, orderChanRx)

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Transmitter(16569, stateTx)
	go bcast.Receiver(16569, stateRx)
	go bcast.Transmitter(16570, orderChanTx)
	go bcast.Receiver(16570, orderChanRx)
	go statehandler.HandlePeerUpdates(peerUpdateCh, stateRx)
	go statehandler.Send(stateTx, &elevator)
	go assigner.Assigner(stateRx, drv_buttons, cabOrder, orderChanTx, &elevator)

	//go assigner.AssigningHallOrders(orderChanRx, &elevator, cabOrder)
	select {}

}
