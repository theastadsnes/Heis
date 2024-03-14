package main

import (
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorFSM"
	"Heis/elevatorutilities"
	"Heis/network/bcast"
	"Heis/network/networking"
	"Heis/network/peers"
	"Heis/watchdog"
	"fmt"
	"time"
)

func main() {

	// Initializing
	id := config.InitId()
	fmt.Println(id)
	elevio.Init("localhost:15657", config.NumFloors)
	elevator := config.InitElevState(id)
	elevatorsMap := make(map[string]config.Elevator)

	// Creating channels
	localElevatorChannels := config.LocalElevChannels{
		Drv_buttons:      make(chan elevio.ButtonEvent),
		Drv_floors:       make(chan int),
		Drv_obstr:        make(chan bool),
		AssignHallOrders: make(chan elevio.ButtonEvent),
		HallOrders:       make(chan *config.AssignmentResults),
	}

	networkChannels := config.Networkchannels{
		OrderChanRx: make(chan *config.AssignmentResults, 100),
		OrderChanTx: make(chan *config.AssignmentResults, 100),
		StateRx:     make(chan *config.Elevator, 100),
		StateTx:     make(chan *config.Elevator, 100),
		AckChanRx:   make(chan string, 100),
		AckChanTx:   make(chan string, 100),
	}

	peersChannels := config.Peerchannels{
		PeerUpdateCh: make(chan peers.PeerUpdate, 100),
		PeerTxEnable: make(chan bool, 100),
	}

	// Creating timers
	doorTimer := time.NewTimer(time.Duration(3) * time.Second)
	motorFaultTimer := time.NewTimer(time.Second * 4)

	// Start polling
	go elevio.PollButtons(localElevatorChannels.Drv_buttons)
	go elevio.PollFloorSensor(localElevatorChannels.Drv_floors)
	go elevio.PollObstructionSwitch(localElevatorChannels.Drv_obstr)

	go peers.Transmitter(23853, id, peersChannels.PeerTxEnable)
	go peers.Receiver(23853, peersChannels.PeerUpdateCh)
	go bcast.Transmitter(16563, networkChannels.StateTx)
	go bcast.Receiver(16563, networkChannels.StateRx)
	go bcast.Transmitter(16570, networkChannels.OrderChanTx)
	go bcast.Receiver(16570, networkChannels.OrderChanRx)
	go bcast.Transmitter(16590, networkChannels.AckChanTx)
	go bcast.Receiver(16590, networkChannels.AckChanRx)

	go networking.SendElevatorStates(networkChannels.StateTx, &elevator)
	go watchdog.Watchdog(&elevator, peersChannels.PeerUpdateCh, elevatorsMap, networkChannels.OrderChanTx, networkChannels.AckChanRx)

	go elevatorFSM.ElevatorFsm(&elevator, doorTimer, motorFaultTimer, localElevatorChannels, peersChannels.PeerTxEnable)
	go networking.Networking(&elevator, elevatorsMap, localElevatorChannels, networkChannels)
	elevatorutilities.ReadCabCallsFromBackup(localElevatorChannels.Drv_buttons)

	select {}

}
