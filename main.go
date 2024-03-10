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
	"Heis/watchdog"
	"fmt"

	"Heis/network/statehandler"
	"Heis/singleElev/elevio"
	"Heis/singleElev/fsm"
	"time"
)

func main() {

	// Initialize
	id := config.InitId()
	numFloors := 4

	fmt.Println(id)

	elevio.Init("localhost:15657", numFloors)
	//var elevator config.Elevator
	elevator := config.InitElevState(id)
	// Initialize elevator I/O

	elevatorsMap := make(map[string]config.Elevator)

	// Create channels for elevator I/O events
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	// Create a timer for the door
	doorTimer := time.NewTimer(time.Duration(3) * time.Second)
	motorFaultTimer := time.NewTimer(time.Second * 4)

	peerUpdateCh := make(chan peers.PeerUpdate, 100) //Kanal som sender/mottar uppdateringer i Peer structen, om det er noen som kobles fra nettet eller noen nye tilkoblinger
	peerTxEnable := make(chan bool, 100)             //Kanal som kan brukes for å vise at man ikke er tilgjengelig, selvom man kanskje er på nettet
	stateTx := make(chan *config.Elevator, 100)      //Gjøre denne om til å sende Elevator state
	stateRx := make(chan *config.Elevator, 100)
	orderChanTx := make(chan *costfunc.AssignmentResults, 100)
	orderChanRx := make(chan *costfunc.AssignmentResults, 100)
	ackChanTx := make(chan string, 100)
	ackChanRx := make(chan string, 100)

	// go func() {
	// 	peers := <-peerUpdateCh
	// 	fmt.Printf("Peeeers: %+v", peers)
	// }()

	// Start polling for elevator I/O events
	go elevio.PollButtons(drv_buttons)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go peers.Transmitter(23853, id, peerTxEnable)
	go peers.Receiver(23853, peerUpdateCh)
	go bcast.Transmitter(16563, stateTx)
	go bcast.Receiver(16563, stateRx)
	go bcast.Transmitter(16570, orderChanTx)
	go bcast.Receiver(16570, orderChanRx)
	go bcast.Transmitter(16590, ackChanTx)
	go bcast.Receiver(16590, ackChanRx)

	//go statehandler.HandlePeerUpdates(peerUpdateCh, stateRx)
	go statehandler.SendElevatorStates(stateTx, &elevator)
	go watchdog.WatchDogLostPeers(&elevator, peerUpdateCh, elevatorsMap, orderChanTx, ackChanRx)
	//go watchdog.WatchdogNewPeers(peerUpdateCh, elevatorsMap, orderChanTx)
	go fsm.Fsm(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors, orderChanRx, orderChanTx, stateRx, stateTx, elevatorsMap, motorFaultTimer, peerTxEnable, ackChanRx, ackChanTx)

	watchdog.ReadFromBackup(drv_buttons)

	select {}

}
