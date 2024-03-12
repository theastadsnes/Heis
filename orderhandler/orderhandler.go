package orderhandler

import (
	"Heis/config"
	"Heis/driver/elevio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

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

func RequestsCurrentFloor(e *config.Elevator) bool {
	for b := 0; b < config.NumButtons; b++ {
		if e.Requests[e.Floor][b] {
			return true
		}
	}
	return false
}

func ShouldStop(e *config.Elevator) bool {

	if RequestsCurrentFloor(e) {

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

func ClearLights() {
	for f := 0; f < config.NumFloors; f++ {
		elevio.SetButtonLamp(0, f, false)
		elevio.SetButtonLamp(1, f, false)
		elevio.SetButtonLamp(2, f, false)
	}
	elevio.SetDoorOpenLamp(false)

}

func ClearRequestAtFloor(e *config.Elevator) {
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

func WriteCabCallsToBackup(elevator *config.Elevator) {
	filename := "orderhandler/cabOrder.txt"
	f, err := os.Create(filename)
	if err != nil {
		return
	}

	caborders := make([]bool, config.NumFloors)

	for floors := range elevator.Requests {
		caborders[floors] = elevator.Requests[floors][2]
	}

	cabordersString := strings.Trim(fmt.Sprint(caborders), "[]")
	_, err = f.WriteString(cabordersString)
	if err != nil {
		return
	}

	defer f.Close()
}

func ReadCabCallsFromBackup(buttons chan elevio.ButtonEvent) {
	filename := "orderhandler/cabOrder.txt"
	f, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	caborders := make([]bool, 0)

	cabOrders := strings.Split(string(f), " ")
	for _, order := range cabOrders {
		result, _ := strconv.ParseBool(order)
		caborders = append(caborders, result)
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

func UpdateHallLights(elevator *config.Elevator, elevatorsMap map[string]config.Elevator) {

	var lights [config.NumFloors][config.NumButtons - 1]bool

	for _, id := range elevatorsMap {
		for floor := range id.Requests {
			for button := 0; button < 2; button++ {
				if id.Requests[floor][button] {
					lights[floor][button] = true
				}
			}
		}

	}
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := 0; button < config.NumButtons-1; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, lights[floor][button])
		}
	}
}

func OpenDoor(elevator *config.Elevator, doorTimer *time.Timer) {
	elevator.Behaviour = config.EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	ClearRequestAtFloor(elevator)
	doorTimer.Reset(time.Duration(3) * time.Second)
}

func GoToValidFloor(elevator *config.Elevator) {
	for elevio.GetFloor() == -1 {
		if elevator.Dirn == elevio.MD_Down {
			elevio.SetMotorDirection(elevio.MD_Down)
		}
		if elevator.Dirn == elevio.MD_Up {
			elevio.SetMotorDirection(elevio.MD_Up)
		}
	}
	elevator.Dirn = elevio.MD_Stop
	elevio.SetMotorDirection(elevator.Dirn)
}

func StartMotor(elevator *config.Elevator, direction elevio.MotorDirection, motorFaultTimer *time.Timer) {
	elevator.Dirn = direction
	elevio.SetMotorDirection(elevator.Dirn)
	elevator.Behaviour = config.EB_Moving
	motorFaultTimer.Reset(time.Second * 4)
}
