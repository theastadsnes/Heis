/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main


import (
	//"Heis/singleElev/elevio"
	//"Heis/singleElev/fsm"
	"Heis/network/bcast"
	"Heis/network/localip"
	"Heis/network/peers"
	//"Heis/config"
	"fmt"
	"time"
	"os"
	"flag"
)

type HelloMsg struct {
	Message string
	Iter    int
}
func main() {
	// Define the number of floors in the building
	// numFloors := 4

	// // Initialize elevator I/O
	// elevio.Init("localhost:15657", numFloors)

	// // Create channels for elevator I/O events
	// drv_buttons := make(chan elevio.ButtonEvent)
	// drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)

	// // Create a timer for the door
	// doorTimer := time.NewTimer(time.Duration(3) * time.Second)

	// // Start polling for elevator I/O events
	// go elevio.PollButtons(drv_buttons)
	// go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)

	// // Start the finite state machine for elevator control
	// fsm.Fsm(drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors)


	// const port = 50000
	// elevatorStateChan := make(chan config.LocalElevatorState)
	
	// go bcast.Transmitter(port, elevatorStateChan)
	
    // go sending(elevatorStateChan) // Endret til å kjøre asynkront som en gorutine

    // // Ventemekanisme (for eksempel en uendelig løkke eller select{})
    // select {}

	// fmt.Print(localip.LocalIP())

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from" + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		}
	}

	
}
