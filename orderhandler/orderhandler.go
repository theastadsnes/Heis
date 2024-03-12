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

func RequestsCurrentFloor(elevator *config.Elevator) bool {
	for b := 0; b < config.NumButtons; b++ {
		if elevator.Requests[elevator.Floor][b] {
			return true
		}
	}
	return false
}

func ShouldStop(elevator *config.Elevator) bool {

	if RequestsCurrentFloor(elevator) {

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

func ClearLights() {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < config.NumFloors; f++ {
		for buttons := 0; buttons < config.NumButtons; buttons++ {
			elevio.SetButtonLamp(buttons, f, false)
		}
	}
}

func ClearRequestAtFloor(elevator *config.Elevator) {
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

func ClearAllRequests(numFloors int, elevator *config.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := elevio.ButtonType(0); button < config.NumButtons; button++ {
			elevator.Requests[floor][button] = false
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
	elevio.SetDoorOpenLamp(true)
	ClearRequestAtFloor(elevator)
	elevator.Behaviour = config.EB_DoorOpen
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
