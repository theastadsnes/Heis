package main

import (
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorFsm"
	"Heis/network/bcast"
	"Heis/network/peers"
	"Heis/orderhandler"
	"Heis/watchdog"
	"time"
)

func main() {

	// Initializing
	id := config.InitId()
	elevio.Init("localhost:15657", config.NumFloors)
	elevator := config.InitElevState(id)
	elevatorsMap := make(map[string]config.Elevator)

	// Creating channels
	hardware := config.Hardwarechannels{
		Drv_buttons: make(chan elevio.ButtonEvent),
		Drv_floors:  make(chan int),
		Drv_obstr:   make(chan bool),
	}

	network := config.Networkchannels{
		OrderChanRx: make(chan *config.AssignmentResults, 100),
		OrderChanTx: make(chan *config.AssignmentResults, 100),
		StateRx:     make(chan *config.Elevator, 100),
		StateTx:     make(chan *config.Elevator, 100),
		AckChanRx:   make(chan string, 100),
		AckChanTx:   make(chan string, 100),
	}

	peerschannels := config.Peerchannels{
		PeerUpdateCh: make(chan peers.PeerUpdate, 100),
		PeerTxEnable: make(chan bool, 100),
	}

	// Creating timers
	doorTimer := time.NewTimer(time.Duration(3) * time.Second)
	motorFaultTimer := time.NewTimer(time.Second * 4)

	// Start polling
	go elevio.PollButtons(hardware.Drv_buttons)
	go elevio.PollFloorSensor(hardware.Drv_floors)
	go elevio.PollObstructionSwitch(hardware.Drv_obstr)

	go peers.Transmitter(23853, id, peerschannels.PeerTxEnable)
	go peers.Receiver(23853, peerschannels.PeerUpdateCh)
	go bcast.Transmitter(16563, network.StateTx)
	go bcast.Receiver(16563, network.StateRx)
	go bcast.Transmitter(16570, network.OrderChanTx)
	go bcast.Receiver(16570, network.OrderChanRx)
	go bcast.Transmitter(16590, network.AckChanTx)
	go bcast.Receiver(16590, network.AckChanRx)

	go watchdog.SendElevatorStates(network.StateTx, &elevator)
	go watchdog.Watchdog(&elevator, peerschannels.PeerUpdateCh, elevatorsMap, network.OrderChanTx, network.AckChanRx)

	go elevatorFsm.ElevatorFsm(&elevator, doorTimer, motorFaultTimer, config.NumFloors, elevatorsMap, hardware, network, peerschannels.PeerTxEnable)

	orderhandler.ReadCabCallsFromBackup(hardware.Drv_buttons)

	select {}

}
