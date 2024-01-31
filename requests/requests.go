package requests

import (
	//"Heis/elevio"
	"Heis/elevio"
	"Heis/fsm"
)

func Requests_chooseDirection(e *fsm.Elevator) {
	var Floor int
	for f := 0; f < fsm.NumFloors; f++ {
		for b := 0; b < fsm.NumButtons; b++ {
			if e.Requests[f][b] == 1 {
				Floor = f
				e.NextDest = f
			}
		}
	}
	if Floor > e.Floor {
		elevio.SetMotorDirection(elevio.MD_Up)
	}

}
func Fsm_onRequestButtonPress(btn_floor int, btn_type elevio.ButtonType) {

	fsm.Our_elevator.Requests[btn_floor][btn_type] = 1
	Requests_chooseDirection(&fsm.Our_elevator)
	/*
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
	*/

	//LYS er i c kode her, skjønte ikke hvorfor den skal slå på lys nå?
}
