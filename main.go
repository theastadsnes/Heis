/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main

import (
	"Heis/costfunc"
	"fmt"
)

/**
 * @brief The entry point for the elevator control program.
 */
func main() {

	/**
	// Define the number of floors in the building
	numFloors := 4

	// Initialize elevator I/O
	elevio.Init("localhost:15657", numFloors)

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

	// Start the finite state machine for elevator control
	fsm.Fsm(drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors)

	*/

	hallRequests := [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}}
	states := map[string]costfunc.HRAElevState{
		"one": {
			Behavior:    "moving",
			Floor:       2,
			Direction:   "up",
			CabRequests: []bool{false, false, false, true},
		},
		"two": {
			Behavior:    "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: []bool{false, false, false, false},
		},
	}

	output, err := costfunc.Costfunc(hallRequests, states)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Output:")
	for k, v := range output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

}
