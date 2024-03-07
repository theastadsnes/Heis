/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/network/bcast"
	"Heis/network/peers"

	"Heis/network/statehandler"
	"Heis/singleElev/elevio"
	"Heis/singleElev/fsm"
	"Heis/watchdog"
	"time"
)

func main() {

	// Initialize
	id := config.InitId()
	numFloors := 4
	elevator := config.InitElevState(id)
	// Initialize elevator I/O
	elevio.Init("localhost:15657", numFloors)
	elevatorsMap := make(map[string]config.Elevator)

	// Create channels for elevator I/O events
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	// Create a timer for the door
	doorTimer := time.NewTimer(time.Duration(3) * time.Second)

	peerUpdateCh := make(chan peers.PeerUpdate) //Kanal som sender/mottar uppdateringer i Peer structen, om det er noen som kobles fra nettet eller noen nye tilkoblinger
	peerTxEnable := make(chan bool)             //Kanal som kan brukes for å vise at man ikke er tilgjengelig, selvom man kanskje er på nettet
	stateTx := make(chan *config.Elevator)      //Gjøre denne om til å sende Elevator state
	stateRx := make(chan *config.Elevator)
	orderChanTx := make(chan *costfunc.AssignmentResults)
	orderChanRx := make(chan *costfunc.AssignmentResults)

	// Start polling for elevator I/O events
	go elevio.PollButtons(drv_buttons)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Transmitter(16569, stateTx)
	go bcast.Receiver(16569, stateRx)
	go bcast.Transmitter(16570, orderChanTx)
	go bcast.Receiver(16570, orderChanRx)
	go statehandler.HandlePeerUpdates(peerUpdateCh, stateRx)
	go statehandler.SendElevatorStates(stateTx, &elevator)
	go watchdog.WatchDog(&elevator, peerUpdateCh, elevatorsMap, orderChanTx)
	go fsm.Fsm(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors, orderChanRx, orderChanTx, stateRx, stateTx, elevatorsMap)

	select {}

}
