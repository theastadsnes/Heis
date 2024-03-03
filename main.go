/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/singleElev/elevio"
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

	elevators := map[string]config.Elevator{
		"elevator1": {
			Floor:     1,
			Dirn:      elevio.MD_Up,
			Requests:  [4][4]int{{0, 1, 0}, {0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
			Behaviour: config.EB_Moving,
			ID:        "elevator1",
		},
		"elevator2": {
			Floor:     3,
			Dirn:      elevio.MD_Down,
			Requests:  [4][4]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, {0, 0, 0}},
			Behaviour: config.EB_Idle,
			ID:        "elevator2",
		},
		"elevator3": {
			Floor:     2,
			Dirn:      elevio.MD_Stop,
			Requests:  [4][4]int{{0, 0, 1}, {1, 0, 0}, {0, 1, 0}, {0, 0, 0}},
			Behaviour: config.EB_DoorOpen,
			ID:        "elevator3",
		},
	}

	states := costfunc.TransformElevatorStates(elevators)

	fmt.Println("Transformed Elevator States:")
	for id, state := range states {
		fmt.Printf("ID: %s, State: %+v\n", id, state)
	}

	hallRequests := [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}}

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
