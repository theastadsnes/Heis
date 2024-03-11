package statemachines

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/packetloss"
	"Heis/singleElev/elevio"
	"Heis/singleElev/requests"
	"fmt"
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
				requests.ClearRequestAtFloor(elevator, doorTimer)
				doorTimer.Reset(time.Duration(3) * time.Second)
			} else {
				elevator.Requests[orderFloor][orderButton] = true
			}
		case elevator.Behaviour == config.EB_Moving:
			elevator.Requests[orderFloor][orderButton] = true

		case elevator.Behaviour == config.EB_Idle:
			if orderFloor == elevator.Floor {
				elevio.SetDoorOpenLamp(true)
				requests.ClearRequestAtFloor(elevator, doorTimer)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
			} else {
				elevator.Requests[orderFloor][orderButton] = true
				if requests.RequestsAbove(elevator) {
					elevator.Dirn = elevio.MD_Up
					elevio.SetMotorDirection(elevator.Dirn)
					elevator.Behaviour = config.EB_Moving

				} else if requests.RequestsBelow(elevator) {
					elevator.Dirn = elevio.MD_Down
					elevio.SetMotorDirection(elevator.Dirn)
					elevator.Behaviour = config.EB_Moving
				}
			}
		}
	}

}

func updateHallOrders(elevator *config.Elevator, orderFloor *[config.NumFloors][config.NumButtons - 2]bool, newAssignedOrders *costfunc.AssignmentResults) {

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

func HallOrderFSM(elevator *config.Elevator, newAssignedOrders *costfunc.AssignmentResults, doorTimer *time.Timer, motorFaultTimer *time.Timer) {

	var orderFloor [config.NumFloors][config.NumButtons - 2]bool
	updateHallOrders(elevator, &orderFloor, newAssignedOrders)
	fmt.Println(orderFloor)

	for floor := 0; floor < config.NumFloors; floor++ {
		for button := 0; button < config.NumButtons-2; button++ {
			if orderFloor[floor][button] {
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:

					if floor == elevator.Floor {
						elevio.SetDoorOpenLamp(true)
						requests.ClearRequestAtFloor(elevator, doorTimer)
						doorTimer.Reset(time.Duration(3) * time.Second)
					}
				case (elevator.Behaviour == config.EB_Moving) && !requests.HasRequests(elevator):
					for elevio.GetFloor() == -1 {
						if elevator.Dirn == elevio.MD_Down {
							elevio.SetMotorDirection(elevio.MD_Down)
						}
						if elevator.Dirn == elevio.MD_Up {
							elevio.SetMotorDirection(elevio.MD_Up)
						}
					}
				case elevator.Behaviour == config.EB_Idle:
					if floor == elevator.Floor {
						elevio.SetDoorOpenLamp(true)
						requests.ClearRequestAtFloor(elevator, doorTimer)
						elevator.Behaviour = config.EB_DoorOpen
						doorTimer.Reset(time.Duration(3) * time.Second)
					} else {
						if requests.RequestsAbove(elevator) {
							elevator.Dirn = elevio.MD_Up
							elevio.SetMotorDirection(elevator.Dirn)
							elevator.Behaviour = config.EB_Moving
							motorFaultTimer.Reset(time.Second * 4)
						} else if requests.RequestsBelow(elevator) {
							elevator.Dirn = elevio.MD_Down
							elevio.SetMotorDirection(elevator.Dirn)
							elevator.Behaviour = config.EB_Moving
							motorFaultTimer.Reset(time.Second * 4)
						}
					}

				}
			}
		}
	}
}
func AssignHallOrders(orderChanTx chan *costfunc.AssignmentResults, ElevatorsMap map[string]config.Elevator, ackChanRx chan string) {

	transStates := costfunc.TransformElevatorStates(ElevatorsMap)
	hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
	newOrders := costfunc.GetRequestStruct(hallRequests, transStates)
	//orderChanTx <- &newOrders
	go packetloss.WaitForAllACKs(orderChanTx, ElevatorsMap, ackChanRx, newOrders)

}

func UpdateLights(elevator *config.Elevator, elevatorsMap map[string]config.Elevator) {

	var lights [config.NumFloors][config.NumButtons - 2]bool

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
		for button := 0; button < config.NumButtons-2; button++ {
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, lights[floor][button])
		}
	}
}
