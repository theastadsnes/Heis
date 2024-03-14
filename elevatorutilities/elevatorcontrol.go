package elevatorutilities

import (
	"Heis/config"
	"Heis/driver/elevio"
	"time"
)

func ClearLights() {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < config.NumFloors; f++ {
		for buttons := 0; buttons < config.NumButtons; buttons++ {
			elevio.SetButtonLamp(elevio.ButtonType(buttons), f, false)
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
