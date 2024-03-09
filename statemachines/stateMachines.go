package statemachines

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/singleElev/elevio"
	"Heis/singleElev/requests"
	"time"
	//"fmt"
)

func CabOrderFSM(elevator *config.Elevator, orderFloor int, orderButton elevio.ButtonType, doorTimer *time.Timer) {
	if !elevio.GetStop() {
		elevio.SetButtonLamp(orderButton, orderFloor, true)
		switch {
		case elevator.Behaviour == config.EB_DoorOpen:
			if orderFloor == elevator.Floor {
				elevio.SetDoorOpenLamp(true)
				requests.Clear_request_at_floor(elevator)
				doorTimer.Reset(time.Duration(3) * time.Second)
			} else {
				elevator.Requests[orderFloor][orderButton] = 1
			}
		case elevator.Behaviour == config.EB_Moving:
			elevator.Requests[orderFloor][orderButton] = 1
		case elevator.Behaviour == config.EB_Idle:
			if orderFloor == elevator.Floor {
				elevio.SetDoorOpenLamp(true)
				requests.Clear_request_at_floor(elevator)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
			} else {
				elevator.Requests[orderFloor][orderButton] = 1
				if requests.Requests_above(elevator) {
					elevator.Dirn = elevio.MD_Up
					elevio.SetMotorDirection(elevator.Dirn)
					elevator.Behaviour = config.EB_Moving
				} else if requests.Requests_below(elevator) {
					elevator.Dirn = elevio.MD_Down
					elevio.SetMotorDirection(elevator.Dirn)
					elevator.Behaviour = config.EB_Moving
				}
			}
		}
	}

}

func HallOrderFSM(elevator *config.Elevator, newAssignedOrders *costfunc.AssignmentResults, doorTimer *time.Timer) {
	//variabler for etasje til de nye ordrene og typen
	var orderFloor int
	//var orderButton elevio.ButtonType

	for _, assignments := range (*newAssignedOrders).Assignments {
		if assignments.ID == elevator.Id {
			for floor := 0; floor < config.NumFloors; floor++ {
				if assignments.UpRequests[floor] {
					elevator.Requests[floor][0] = 1
					orderFloor = floor
					//orderButton = elevio.BT_HallUp
					//elevio.SetButtonLamp(orderButton, orderFloor, true)
				} else if !assignments.UpRequests[floor] {
					elevator.Requests[floor][0] = 0
					//elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
				}
				if assignments.DownRequests[floor] {
					elevator.Requests[floor][1] = 1
					orderFloor = floor
					//orderButton = elevio.BT_HallDown
					//elevio.SetButtonLamp(orderButton, orderFloor, true)
				} else if !assignments.DownRequests[floor] {
					elevator.Requests[floor][1] = 0
					//elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
				}
			}
		}
	}

	switch {
	case elevator.Behaviour == config.EB_DoorOpen:
		if orderFloor == elevator.Floor {
			elevio.SetDoorOpenLamp(true)
			requests.Clear_request_at_floor(elevator)
			doorTimer.Reset(time.Duration(3) * time.Second)
		}
	case elevator.Behaviour == config.EB_Moving:

	case elevator.Behaviour == config.EB_Idle:
		if orderFloor == elevator.Floor {
			elevio.SetDoorOpenLamp(true)
			requests.Clear_request_at_floor(elevator)
			elevator.Behaviour = config.EB_DoorOpen
			doorTimer.Reset(time.Duration(3) * time.Second)
		} else {
			if requests.Requests_above(elevator) {
				elevator.Dirn = elevio.MD_Up
				elevio.SetMotorDirection(elevator.Dirn)
				elevator.Behaviour = config.EB_Moving
			} else if requests.Requests_below(elevator) {
				elevator.Dirn = elevio.MD_Down
				elevio.SetMotorDirection(elevator.Dirn)
				elevator.Behaviour = config.EB_Moving
			}
		}

	}
}

func AssignHallOrders(orderChanTx chan *costfunc.AssignmentResults, ElevatorsMap map[string]config.Elevator) {
	// elevator.Requests[orderFloor][orderButton] = 1
	// ElevatorsMap[elevator.Id].Requests[orderFloor][orderButton] = 1

	transStates := costfunc.TransformElevatorStates(ElevatorsMap)
	hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
	newOrders := costfunc.GetRequestStruct(hallRequests, transStates)
	orderChanTx <- &newOrders

}

func UpdateLights(elevator *config.Elevator, elevatorsMap map[string]config.Elevator) {

	var lights[config.NumFloors][config.NumButtons-2]bool

	for _, id := range elevatorsMap {
		for floor := range id.Requests {
			for button := 0; button < 2; button++ {
				if id.Requests[floor][button] == 1 {
					lights[floor][button] = true
				}  
			}
		}
		
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := 0; button < config.NumButtons-2; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, lights[floor][button])
		}
	}
}
