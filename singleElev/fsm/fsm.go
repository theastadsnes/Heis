/**
 * @file fsm.go
 * @brief Contains the finite state machine (FSM) logic for elevator control.
 */

package fsm

import (
	"Heis/config"
	"Heis/singleElev/elevio"
	"Heis/singleElev/requests"
	"time"
)

/**
 * @brief Implements the finite state machine (FSM) logic for elevator control.
 * @param buttons Channel for receiving button events.
 * @param floors Channel for receiving floor events.
 * @param obstr Channel for receiving obstruction events.
 * @param stop Channel for receiving stop events.
 * @param doorTimer Pointer to the door timer.
 * @param numFloors Total number of floors in the building.
 */



func Fsm(elevator *config.Elevator,buttons chan elevio.ButtonEvent, floors chan int, obstr chan bool, stop chan bool, doorTimer *time.Timer, numFloors int) {
	requests.Clear_lights()
	

	for {
		select {
		case order := <-buttons:
			if !elevio.GetStop() {
				elevio.SetButtonLamp(order.Button, order.Floor, true)
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:
					if order.Floor == elevator.Floor {
						elevio.SetDoorOpenLamp(true)
						requests.Clear_request_at_floor(elevator)
						doorTimer.Reset(time.Duration(3) * time.Second)
					} else {
						elevator.Requests[order.Floor][order.Button] = 1
					}
				case elevator.Behaviour == config.EB_Moving:
					elevator.Requests[order.Floor][order.Button] = 1
				case elevator.Behaviour == config.EB_Idle:
					if order.Floor == elevator.Floor {
						elevio.SetDoorOpenLamp(true)
						requests.Clear_request_at_floor(elevator)
						elevator.Behaviour = config.EB_DoorOpen
						doorTimer.Reset(time.Duration(3) * time.Second)
					} else {
						elevator.Requests[order.Floor][order.Button] = 1
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

		case floor := <-floors:
			elevator.Floor = floor
			if requests.Should_stop(elevator) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				requests.Clear_request_at_floor(elevator)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
			}

		case <-doorTimer.C:
			elevio.SetDoorOpenLamp(false)
			switch {
			case elevator.Behaviour == config.EB_DoorOpen:
				requests.Requests_chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Dirn)
				if elevator.Dirn == elevio.MD_Stop {
					elevator.Behaviour = config.EB_Idle
				} else {
					elevator.Behaviour = config.EB_Moving
				}
			}

		case obstruction := <-obstr:
			if elevator.Behaviour == config.EB_DoorOpen {
				if obstruction {
					if !doorTimer.Stop() {
						<-doorTimer.C
					}
				} else {
					doorTimer.Reset(time.Duration(3) * time.Second)
				}
			}

		case a := <-stop:
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
				requests.Clear_lights()
				requests.Clear_all_requests(numFloors, elevator)
				elevio.SetStopLamp(true)

				if elevator.Behaviour != config.EB_Moving {
					elevio.SetDoorOpenLamp(true)
					elevator.Behaviour = config.EB_DoorOpen
					doorTimer.Reset(time.Duration(3) * time.Second)
				}

				if elevator.Behaviour == config.EB_Moving {
					time.Sleep(3 * time.Second)
					elevio.SetDoorOpenLamp(false)
					elevator.Behaviour = config.EB_Idle

				}
			} else {
				requests.Clear_lights()
				requests.Clear_request_at_floor(elevator)
				elevio.SetStopLamp(false)

				if elevator.Behaviour != config.EB_Moving {
					doorTimer.Reset(time.Duration(3) * time.Second)
				}

			}

		}
	}
}
