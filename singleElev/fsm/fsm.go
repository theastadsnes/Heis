/**
 * @file fsm.go
 * @brief Contains the finite state machine (FSM) logic for elevator control.
 */

package fsm

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/singleElev/elevio"
	"Heis/singleElev/requests"
	"Heis/statemachines"
	"Heis/watchdog"
	"fmt"
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

func Fsm(elevator *config.Elevator, buttons chan elevio.ButtonEvent, floors chan int, obstr chan bool, stop chan bool, doorTimer *time.Timer, numFloors int, orderChanRx chan *costfunc.AssignmentResults, orderChanTx chan *costfunc.AssignmentResults, stateRx chan *config.Elevator, stateTx chan *config.Elevator, elevatorsMap map[string]config.Elevator, motorFaultTimer *time.Timer, peerTxEnable chan bool) {
	requests.Clear_lights()
	//elevatorsMap := make(map[string]config.Elevator)

	for {
		select {
		case stateReceived := <-stateRx:

			elevatorsMap[stateReceived.Id] = *stateReceived
			statemachines.UpdateLights(elevator, elevatorsMap)

		case order := <-buttons:
			//peerTxEnable <- true
			if order.Button == 2 {
				statemachines.CabOrderFSM(elevator, order.Floor, order.Button, doorTimer)
			} else {
				elevator.Requests[order.Floor][order.Button] = 1
				elevatorsMap[elevator.Id].Requests[order.Floor][order.Button] = 1
				statemachines.AssignHallOrders(orderChanTx, elevatorsMap)
			}

		case newAssignedHallOrders := <-orderChanRx:
			fmt.Println("ASSIGNING HALL ORDER")
			fmt.Println(newAssignedHallOrders)
			
			statemachines.HallOrderFSM(elevator, newAssignedHallOrders, doorTimer, motorFaultTimer)

		case floor := <-floors:
			peerTxEnable <- true
			elevator.Floor = floor

			elevio.SetFloorIndicator(floor)
			motorFaultTimer.Reset(time.Second * 4)
			fmt.Println(elevator.Dirn)

			if elevator.Dirn == elevio.MD_Stop {
				motorFaultTimer.Stop()
			}

			if requests.Should_stop(elevator) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				requests.Clear_request_at_floor(elevator, doorTimer)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
				motorFaultTimer.Stop()
			}

		case <-doorTimer.C:

			if !elevio.GetObstruction() {
				elevio.SetDoorOpenLamp(false)
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:
					requests.Requests_chooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Dirn)
					if elevator.Dirn == elevio.MD_Stop {
						elevator.Behaviour = config.EB_Idle
					} else {
						elevator.Behaviour = config.EB_Moving
						motorFaultTimer.Reset(time.Second * 4)
					}

				}

			} else {
				motorFaultTimer.Reset(time.Second * 4)
			}

		case obstruction := <-obstr:
			if obstruction {

				if elevator.Behaviour == config.EB_DoorOpen {
					motorFaultTimer.Reset(time.Second * 7)
					if !doorTimer.Stop() {
						select {
						case <-doorTimer.C:
						default:
						}
					}
					elevio.SetDoorOpenLamp(true)
				}

			} else if elevator.Behaviour == config.EB_DoorOpen {
				motorFaultTimer.Stop()
				doorTimer.Reset(time.Duration(3) * time.Second)
			}

		case <-motorFaultTimer.C:
			fmt.Println("MOTORFAULT", elevator.Floor)
			peerTxEnable <- false

			for elevio.GetFloor() == -1 {
				if elevator.Dirn == elevio.MD_Down {
					elevio.SetMotorDirection(elevio.MD_Down)
				}
				if elevator.Dirn == elevio.MD_Up {
					elevio.SetMotorDirection(elevio.MD_Up)
				}

			}
			elevator.Dirn = elevio.MD_Stop
			elevio.SetMotorDirection(elevator.Dirn)
			elevio.SetDoorOpenLamp(true)
			elevator.Behaviour = config.EB_DoorOpen
			doorTimer.Reset(time.Duration(3) * time.Second)

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
				requests.Clear_request_at_floor(elevator, doorTimer)
				elevio.SetStopLamp(false)

				if elevator.Behaviour != config.EB_Moving {
					doorTimer.Reset(time.Duration(3) * time.Second)
				}

			}

		}

		watchdog.WriteToBackup(elevator)
	}
}
