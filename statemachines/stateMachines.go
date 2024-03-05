package statemachines

import (
	"Heis/config"
	"Heis/costfunc"
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
				} else if !assignments.UpRequests[floor] {
					elevator.Requests[floor][0] = 0
				}
				if assignments.DownRequests[floor] {
					elevator.Requests[floor][1] = 1
					orderFloor = floor
					//orderButton = elevio.BT_HallDown
				} else if !assignments.DownRequests[floor] {
					elevator.Requests[floor][1] = 0
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

func AssignerFSM(stateRx chan *config.Elevator, orderChanTx chan *costfunc.AssignmentResults, elevator *config.Elevator, enAssigner chan bool) {
	for {
		select {
		case <-enAssigner:

		case stateReceived := <-stateRx:

			ElevatorsMap := make(map[string]config.Elevator)

			//fmt.Println("STATE RECEIVED", stateReceived)
			ElevatorsMap[stateReceived.Id] = *stateReceived
			fmt.Println(ElevatorsMap)
			transStates := costfunc.TransformElevatorStates(ElevatorsMap)
			hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
			newOrders := costfunc.GetRequestStruct(hallRequests, transStates)
			fmt.Println(newOrders)
			orderChanTx <- &newOrders

		}

	}
}
