package fsm

import "Heis/elevio"

const (
	NumFloors  = 4 // Example values
	NumButtons = 4
)

var Our_elevator Elevator
var Pair DirnBehaviourPair

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

/*
type Dirn int

const (
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up   Dirn = 1
)
*/

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
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

	} else if Our_elevator.Behaviour == EB_Moving {
		Our_elevator.Requests[btn_floor][btn_type] = 1

	} else if Our_elevator.Behaviour == EB_Idle{
		Our_elevator.Requests[btn_floor][btn_type] = 1
		Pair = Requests_chooseDirection(Our_elevator)
		Our_elevator.Dirn = Pair.Dirn
		Our_elevator.Behaviour = Pair.Behaviour

		if Our_elevator.Behaviour == EB_DoorOpen {
			elevio.SetDoorOpenLamp(true)
			Timer_start(Our_elevator.Config.DoorOpenDurationS)
			Our_elevator = Requests_clearAtCurrentFloor(Our_elevator)

		} else if Our_elevator.Behaviour == EB_Moving{
			elevio.SetMotorDirection(Our_elevator.Dirn)

		} else if Our_elevator.Behaviour == EB_Idle{

		}


	}

	//LYS er i c kode her, skjønte ikke hvorfor den skal slå på lys nå?
}
