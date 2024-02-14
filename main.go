package main

import (
	"Heis/elevio"
	"Heis/fsm"
	"time"

	"Heis/requests"
	"fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	doorTimer := time.NewTimer(time.Duration(3) * time.Second)

	requests.Clear_lights()
	for {
		select {
		case order := <-drv_buttons:
			elevio.SetButtonLamp(order.Button, order.Floor, true)
			switch {
			case fsm.Our_elevator.Behaviour == fsm.EB_DoorOpen:
				if order.Floor == fsm.Our_elevator.Floor {
					elevio.SetDoorOpenLamp(true)
					requests.Clear_request_at_floor(&fsm.Our_elevator)
					doorTimer.Reset(time.Duration(3) * time.Second)
				} else {
					fsm.Our_elevator.Requests[order.Floor][order.Button] = 1

				}
			case fsm.Our_elevator.Behaviour == fsm.EB_Moving:
				fsm.Our_elevator.Requests[order.Floor][order.Button] = 1

			case fsm.Our_elevator.Behaviour == fsm.EB_Idle:
				if order.Floor == fsm.Our_elevator.Floor {

					elevio.SetDoorOpenLamp(true)
					requests.Clear_request_at_floor(&fsm.Our_elevator)
					fsm.Our_elevator.Behaviour = fsm.EB_DoorOpen
					doorTimer.Reset(time.Duration(3) * time.Second)

				} else {
					fsm.Our_elevator.Requests[order.Floor][order.Button] = 1
					if requests.Requests_above(fsm.Our_elevator) {
						fsm.Our_elevator.Dirn = elevio.MD_Up
						elevio.SetMotorDirection(fsm.Our_elevator.Dirn)
						fsm.Our_elevator.Behaviour = fsm.EB_Moving
					} else if requests.Requests_below(fsm.Our_elevator) {
						fsm.Our_elevator.Dirn = elevio.MD_Down
						elevio.SetMotorDirection(fsm.Our_elevator.Dirn)
						fsm.Our_elevator.Behaviour = fsm.EB_Moving
					}
				}
			}

		case floor := <-drv_floors:
			fsm.Our_elevator.Floor = floor
			fmt.Printf("%+v\n", floor)
			fmt.Printf("retning før stop:")
			fmt.Print(fsm.Our_elevator.Dirn)

			if requests.Should_stop(fsm.Our_elevator) {

				fmt.Printf("retning:")
				fmt.Print(fsm.Our_elevator.Dirn)
				elevio.SetMotorDirection(elevio.MD_Stop)

				elevio.SetDoorOpenLamp(true)
				requests.Clear_request_at_floor(&fsm.Our_elevator)
				fsm.Our_elevator.Behaviour = fsm.EB_DoorOpen
				elevio.SetDoorOpenLamp(true)

				doorTimer.Reset(time.Duration(3) * time.Second)

			}
		case <-doorTimer.C:
			elevio.SetDoorOpenLamp(false)
			switch {
			case fsm.Our_elevator.Behaviour == fsm.EB_DoorOpen:
				requests.Requests_chooseDirection(&fsm.Our_elevator)
				elevio.SetMotorDirection(fsm.Our_elevator.Dirn)

				if fsm.Our_elevator.Dirn == elevio.MD_Stop {
					fsm.Our_elevator.Behaviour = fsm.EB_Idle
				} else {
					fsm.Our_elevator.Behaviour = fsm.EB_Moving

				}
			}

		case obstruction := <-drv_obstr:
			if fsm.Our_elevator.Behaviour == fsm.EB_DoorOpen {
				if obstruction {
					if !doorTimer.Stop() {
						<-doorTimer.C
					}

				} else {
					doorTimer.Reset(time.Duration(3) * time.Second)
				}

			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
