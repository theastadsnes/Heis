package requests

import (
	//"Heis/elevio"
	"Heis/elevio"
	"Heis/fsm"
	"fmt"
)

var Floors int = 4
var Buttons int = 4

func Requests_above(e fsm.Elevator) bool {
	for f := e.Floor + 1; f < Floors; f++ {
		for b := 0; b < Buttons; b++ {
			if e.Requests[f][b] == 1 {
				return true
			}
		}
	}
	return false
}

func Requests_below(e fsm.Elevator) bool { //Kanskje feil, nå teller vi nedover fra etasjen vi er i, men kanskje riktig å telle fra 0 TIL etasjen vi er i
	for f := e.Floor - 1; f >= 0; f-- {
		for b := 0; b < Buttons; b++ {
			if e.Requests[f][b] == 1 {
				return true
			}
		}
	}
	return false
}

func Requests_current_floor(e fsm.Elevator) bool {

	for b := 0; b < Buttons; b++ {
		if e.Requests[e.Floor][b] == 1 {
			return true
		}

	}
	return false
}

func Should_stop(e fsm.Elevator) bool {
	if Requests_current_floor(e) {
		switch {
		case e.Dirn == elevio.MD_Down:
			if fsm.Our_elevator.Requests[e.Floor][elevio.BT_HallUp] == 1 && Requests_below(e) {
				return false
			}
		case e.Dirn == elevio.MD_Up:
			if fsm.Our_elevator.Requests[e.Floor][elevio.BT_HallDown] == 1 && Requests_above(e) {
				return false
			}

		}
		return true
	}
	return false
}

func Clear_lights() {
	for f := 0; f < Floors; f++ {
		elevio.SetButtonLamp(0, f, false)
		elevio.SetButtonLamp(1, f, false)
		elevio.SetButtonLamp(2, f, false)

	}
}

func Clear_request_at_floor(e *fsm.Elevator) {
	fsm.Our_elevator.Requests[e.Floor][int(elevio.BT_Cab)] = 0
	elevio.SetButtonLamp(elevio.BT_Cab, e.Floor, false)

	switch {
	case fsm.Our_elevator.Dirn == elevio.MD_Up:
		fsm.Our_elevator.Requests[e.Floor][int(elevio.BT_HallUp)] = 0
		elevio.SetButtonLamp(elevio.BT_HallUp, e.Floor, false)
		if !Requests_above(fsm.Our_elevator) {
			fsm.Our_elevator.Requests[e.Floor][int(elevio.BT_HallDown)] = 0
			elevio.SetButtonLamp(elevio.BT_HallDown, e.Floor, false)
		}
	case fsm.Our_elevator.Dirn == elevio.MD_Down:
		fsm.Our_elevator.Requests[e.Floor][int(elevio.BT_HallDown)] = 0
		elevio.SetButtonLamp(elevio.BT_HallDown, e.Floor, false)
		if !Requests_below(fsm.Our_elevator) {
			fsm.Our_elevator.Requests[e.Floor][int(elevio.BT_HallUp)] = 0
			elevio.SetButtonLamp(elevio.BT_HallUp, e.Floor, false)
		}

	}
}

func Requests_chooseDirection(e *fsm.Elevator) {
	fmt.Printf("retning, inni choose:")
	fmt.Print(e.Dirn)
	switch e.Dirn {
	case elevio.MD_Up:
		if Requests_above(*e) {
			e.Dirn = elevio.MD_Up
		} else if Requests_below(*e) {
			e.Dirn = elevio.MD_Down
		} else {
			e.Dirn = elevio.MD_Stop
		}
	case elevio.MD_Down:
		if Requests_below(*e) {
			e.Dirn = elevio.MD_Down
		} else if Requests_above(*e) {
			e.Dirn = elevio.MD_Up
		} else {
			e.Dirn = elevio.MD_Stop
		}
	case elevio.MD_Stop:
		if Requests_above(*e) {
			e.Dirn = elevio.MD_Up
		} else if Requests_below(*e) {
			e.Dirn = elevio.MD_Down
		} else {
			e.Dirn = elevio.MD_Stop
		}
	}

}
