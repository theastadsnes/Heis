package statemachines

import (
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorutilities"
	"time"
)

func CabOrderStateMachine(elevator *config.Elevator, orderFloor int, orderButton elevio.ButtonType, doorTimer *time.Timer, motorFaultTimer *time.Timer) {

	elevio.SetButtonLamp(orderButton, orderFloor, true)

	switch {
	case elevator.Behaviour == config.EB_DoorOpen:
		if orderFloor == elevator.Floor {
			elevatorutilities.OpenDoor(elevator, doorTimer)

		} else {
			elevator.Requests[orderFloor][orderButton] = true
		}
	case elevator.Behaviour == config.EB_Moving:
		elevator.Requests[orderFloor][orderButton] = true

	case elevator.Behaviour == config.EB_Idle:
		if orderFloor == elevator.Floor {
			elevatorutilities.OpenDoor(elevator, doorTimer)

		} else {
			elevator.Requests[orderFloor][orderButton] = true
			if elevatorutilities.RequestsAbove(elevator) {
				elevatorutilities.StartMotor(elevator, elevio.MD_Up, motorFaultTimer)

			} else if elevatorutilities.RequestsBelow(elevator) {
				elevatorutilities.StartMotor(elevator, elevio.MD_Down, motorFaultTimer)

			}
		}
	}
}

func updateHallOrders(elevator *config.Elevator, orderFloor *[config.NumFloors][config.NumButtons - 1]bool, newAssignedOrders *config.AssignmentResults) {

	for _, assignments := range newAssignedOrders.Assignments {
		if assignments.ID == elevator.Id {
			for floor := 0; floor < config.NumFloors; floor++ {
				if assignments.UpRequests[floor] {
					elevator.Requests[floor][elevio.BT_HallUp] = true
					orderFloor[floor][elevio.BT_HallUp] = true

				} else if !assignments.UpRequests[floor] {
					elevator.Requests[floor][elevio.BT_HallUp] = false

				}

				if assignments.DownRequests[floor] {
					elevator.Requests[floor][elevio.BT_HallDown] = true
					orderFloor[floor][elevio.BT_HallDown] = true

				} else if !assignments.DownRequests[floor] {
					elevator.Requests[floor][elevio.BT_HallDown] = false

				}
			}
		}
	}
}

func HallOrderStateMachine(elevator *config.Elevator, newAssignedOrders *config.AssignmentResults, doorTimer *time.Timer, motorFaultTimer *time.Timer) {

	var orderFloor [config.NumFloors][config.NumButtons - 1]bool
	updateHallOrders(elevator, &orderFloor, newAssignedOrders)
	if !elevatorutilities.HasRequests(elevator) {
		elevatorutilities.GoToValidFloor(elevator)
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := 0; button < config.NumButtons-1; button++ {
			if orderFloor[floor][button] {
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:

					if floor == elevator.Floor {
						elevatorutilities.OpenDoor(elevator, doorTimer)
					}

				case elevator.Behaviour == config.EB_Idle:
					if floor == elevator.Floor {
						elevatorutilities.OpenDoor(elevator, doorTimer)

					} else {
						if elevatorutilities.RequestsAbove(elevator) {
							elevatorutilities.StartMotor(elevator, elevio.MD_Up, motorFaultTimer)

						} else if elevatorutilities.RequestsBelow(elevator) {
							elevatorutilities.StartMotor(elevator, elevio.MD_Down, motorFaultTimer)

						}
					}

				}
			}
		}
	}
}