package requests

import (
	//"Heis/elevio"
	"Heis/elevio"
	"Heis/fsm"
	"Heis"
)

func UpdateFloor(drv_floors <-chan int) {
	
	
}

func Requests_chooseDirection(e fsm.Elevator) int {
	for f := 0; f < fsm.NumFloors; f++ {
		for b := 0; b < fsm.NumButtons; b++ {
			if e.Requests[f][b] == 1 {
				var Floor int = f
			}
		}
	}
	Current:= <-main.Drv_floors
	if Floor > Current{
		elevio.SetMotorDirection(elevio.MD_Up)
	}

}
