/**
 * @file requests.go
 * @brief Contains functions related to elevator requests handling.
 */

package Orderhandler

import (
	"Heis/Driver/elevio"
	"Heis/config"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// var Floors int = 4  // Number of floors in the building
// var Buttons int = 4 // Number of elevator buttons (e.g., Up, Down, Cab)

/**
 * @brief Checks if there are any requests above the current floor.
 * @param e The current state of the elevator.
 * @return Returns true if there are requests above the current floor, otherwise false.
 */
func RequestsAbove(e *config.Elevator) bool {
	for f := e.Floor + 1; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

/**
 * @brief Checks if there are any requests below the current floor.
 * @param e The current state of the elevator.
 * @return Returns true if there are requests below the current floor, otherwise false.
 */
func RequestsBelow(e *config.Elevator) bool {
	for f := e.Floor - 1; f >= 0; f-- {
		for b := 0; b < config.NumButtons; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

/**
 * @brief Checks if there are any requests on the current floor.
 * @param e The current state of the elevator.
 * @return Returns true if there are requests on the current floor, otherwise false.
 */
func RequestsCurrentFloor(e *config.Elevator) bool {
	for b := 0; b < config.NumButtons; b++ {
		if e.Requests[e.Floor][b] {
			return true
		}
	}
	return false
}

/**
 * @brief Determines if the elevator should stop at the current floor based on requests.
 * @param e The current state of the elevator.
 * @return Returns true if the elevator should stop at the current floor, otherwise false.
 */
func ShouldStop(e *config.Elevator) bool {

	if RequestsCurrentFloor(e) {
		fmt.Println("retning", e.Dirn)

		switch {
		case e.Dirn == elevio.MD_Down:
			if e.Requests[e.Floor][elevio.BT_HallUp] && RequestsBelow(e) {
				if e.Requests[e.Floor][elevio.BT_HallDown] {
					return true
				} else {
					return false
				}
			}
		case e.Dirn == elevio.MD_Up:
			if e.Requests[e.Floor][elevio.BT_HallDown] && RequestsAbove(e) {
				if e.Requests[e.Floor][elevio.BT_HallUp] {
					return true
				} else {
					return false
				}
			}

		}
		return true
	}
	return false
}

/**
 * @brief Clears all button lights in the elevator.
 */
func ClearLights() {
	for f := 0; f < config.NumFloors; f++ {
		elevio.SetButtonLamp(0, f, false)
		elevio.SetButtonLamp(1, f, false)
		elevio.SetButtonLamp(2, f, false)
	}
	elevio.SetDoorOpenLamp(false)

}

/**
 * @brief Clears requests and button lights at the current floor.
 * @param e A pointer to the current state of the elevator.
 */
func ClearRequestAtFloor(e *config.Elevator, doorTimer *time.Timer) {
	e.Requests[e.Floor][int(elevio.BT_Cab)] = false
	elevio.SetButtonLamp(elevio.BT_Cab, e.Floor, false)

	switch {
	case e.Dirn == elevio.MD_Up:
		e.Requests[e.Floor][int(elevio.BT_HallUp)] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, e.Floor, false)
		if !RequestsAbove(e) {
			e.Requests[e.Floor][int(elevio.BT_HallDown)] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, e.Floor, false)
		}

	case e.Dirn == elevio.MD_Down:

		e.Requests[e.Floor][int(elevio.BT_HallDown)] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, e.Floor, false)
		if !RequestsBelow(e) {
			e.Requests[e.Floor][int(elevio.BT_HallUp)] = false
			elevio.SetButtonLamp(elevio.BT_HallUp, e.Floor, false)
		}

	case e.Dirn == elevio.MD_Stop:
		e.Requests[e.Floor][int(elevio.BT_HallDown)] = false
		e.Requests[e.Floor][int(elevio.BT_HallUp)] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, e.Floor, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, e.Floor, false)
	}

}

/**
 * @brief Chooses the elevator direction based on current requests.
 * @param e A pointer to the current state of the elevator.
 */
func RequestsChooseDirection(e *config.Elevator) {

	switch e.Dirn {
	case elevio.MD_Up:
		if RequestsAbove(e) {
			e.Dirn = elevio.MD_Up
		} else if RequestsBelow(e) {
			e.Dirn = elevio.MD_Down
		} else {
			e.Dirn = elevio.MD_Stop
		}
	case elevio.MD_Down:
		if RequestsBelow(e) {
			e.Dirn = elevio.MD_Down
		} else if RequestsAbove(e) {
			e.Dirn = elevio.MD_Up
		} else {
			e.Dirn = elevio.MD_Stop
		}
	case elevio.MD_Stop:
		if RequestsAbove(e) {
			e.Dirn = elevio.MD_Up
		} else if RequestsBelow(e) {
			e.Dirn = elevio.MD_Down
		} else {
			e.Dirn = elevio.MD_Stop
		}
	}
}

/**
 * @brief Clears requests and button lights at all floors.
 * @param numFloors Number of floors
 */
func ClearAllRequests(numFloors int, e *config.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := elevio.ButtonType(0); button < config.NumButtons; button++ {
			e.Requests[floor][button] = false
			elevio.SetButtonLamp(button, floor, false)
		}
	}
}

func HasRequests(elevator *config.Elevator) bool {
	return RequestsAbove(elevator) || RequestsBelow(elevator)

}

func WriteToBackup(elevator *config.Elevator) {
	filename := "Orderhandler/cabOrder.txt"
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	caborders := make([]bool, config.NumFloors)

	for floors, _ := range elevator.Requests {
		caborders[floors] = elevator.Requests[floors][2]
	}

	cabordersString := strings.Trim(fmt.Sprint(caborders), "[]")
	_, err = f.WriteString(cabordersString)

	defer f.Close()
}

func ReadFromBackup(buttons chan elevio.ButtonEvent) {
	filename := "Orderhandler/cabOrder.txt"
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
