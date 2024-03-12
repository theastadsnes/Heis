package elevatorhelper

import (
	"Heis/config"
	"Heis/driver/elevio"
)

func RequestsAbove(elevator *config.Elevator) bool {
	for f := elevator.Floor + 1; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			if elevator.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

func RequestsBelow(elevator *config.Elevator) bool {
	for f := elevator.Floor - 1; f >= 0; f-- {
		for b := 0; b < config.NumButtons; b++ {
			if elevator.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

func requestsCurrentFloor(elevator *config.Elevator) bool {
	for b := 0; b < config.NumButtons; b++ {
		if elevator.Requests[elevator.Floor][b] {
			return true
		}
	}
	return false
}

func ShouldStop(elevator *config.Elevator) bool {

	if requestsCurrentFloor(elevator) {

		switch {
		case elevator.Dirn == elevio.MD_Down:
			if elevator.Requests[elevator.Floor][elevio.BT_HallUp] && RequestsBelow(elevator) {
				if elevator.Requests[elevator.Floor][elevio.BT_HallDown] {
					return true
				} else {
					return false
				}
			}
		case elevator.Dirn == elevio.MD_Up:
			if elevator.Requests[elevator.Floor][elevio.BT_HallDown] && RequestsAbove(elevator) {
				if elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
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

func clearRequestAtFloor(elevator *config.Elevator) {
	elevator.Requests[elevator.Floor][int(elevio.BT_Cab)] = false
	elevio.SetButtonLamp(elevio.BT_Cab, elevator.Floor, false)

	switch {
	case elevator.Dirn == elevio.MD_Up:
		elevator.Requests[elevator.Floor][int(elevio.BT_HallUp)] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, elevator.Floor, false)
		if !RequestsAbove(elevator) {
			elevator.Requests[elevator.Floor][int(elevio.BT_HallDown)] = false
			elevio.SetButtonLamp(elevio.BT_HallDown, elevator.Floor, false)
		}

	case elevator.Dirn == elevio.MD_Down:

		elevator.Requests[elevator.Floor][int(elevio.BT_HallDown)] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, elevator.Floor, false)
		if !RequestsBelow(elevator) {
			elevator.Requests[elevator.Floor][int(elevio.BT_HallUp)] = false
			elevio.SetButtonLamp(elevio.BT_HallUp, elevator.Floor, false)
		}

	case elevator.Dirn == elevio.MD_Stop:
		elevator.Requests[elevator.Floor][int(elevio.BT_HallDown)] = false
		elevator.Requests[elevator.Floor][int(elevio.BT_HallUp)] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, elevator.Floor, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, elevator.Floor, false)
	}
}

func RequestsChooseDirection(elevator *config.Elevator) {

	switch elevator.Dirn {
	case elevio.MD_Up:
		if RequestsAbove(elevator) {
			elevator.Dirn = elevio.MD_Up
		} else if RequestsBelow(elevator) {
			elevator.Dirn = elevio.MD_Down
		} else {
			elevator.Dirn = elevio.MD_Stop
		}
	case elevio.MD_Down:
		if RequestsBelow(elevator) {
			elevator.Dirn = elevio.MD_Down
		} else if RequestsAbove(elevator) {
			elevator.Dirn = elevio.MD_Up
		} else {
			elevator.Dirn = elevio.MD_Stop
		}
	case elevio.MD_Stop:
		if RequestsAbove(elevator) {
			elevator.Dirn = elevio.MD_Up
		} else if RequestsBelow(elevator) {
			elevator.Dirn = elevio.MD_Down
		} else {
			elevator.Dirn = elevio.MD_Stop
		}
	}
}

func HasRequests(elevator *config.Elevator) bool {
	return RequestsAbove(elevator) || RequestsBelow(elevator)

}
