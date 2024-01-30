package fsm

import "Driver-go/elevio"

const (
	NumFloors  = 4 // Example values
	NumButtons = 4
)

var Our_elevator Elevator

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Dirn int

const (
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up   Dirn = 1
)

type DirnBehaviourPair struct {
	Dirn      Dirn
	Behaviour ElevatorBehaviour
}

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	Floor     int
	Dirn      Dirn
	Requests  [NumFloors][NumButtons]int
	Behaviour ElevatorBehaviour

	Config struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDurationS   float64
	}
}

func Fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType) {

	if Our_elevator.Behaviour == EB_DoorOpen {
		//Har ikke lagt til requests_shouldClearImmediatelym har ikke helt skjønt hva denne gjør, kan se på senere
		Our_elevator.Requests[btn_floor][btn_type] = 1
	}

}
