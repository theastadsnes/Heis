/**
 * @file fsm.go
 * @brief Contains the finite state machine (FSM) logic for elevator control.
 */

package Executer

import (
	"Heis/Assigner"
	"Heis/Driver/elevio"
	"Heis/Orderhandler"
	"Heis/config"
	"Heis/statemachines"
	"fmt"
	"time"
)

func Fsm(elevator *config.Elevator, buttons chan elevio.ButtonEvent, floors chan int, obstr chan bool, stop chan bool, doorTimer *time.Timer, numFloors int, orderChanRx chan *Assigner.AssignmentResults, orderChanTx chan *Assigner.AssignmentResults, stateRx chan *config.Elevator, stateTx chan *config.Elevator, elevatorsMap map[string]config.Elevator, motorFaultTimer *time.Timer, peerTxEnable chan bool, ackChanRx chan string, ackChanTx chan string) {
	Orderhandler.ClearLights()

	for {
		select {
		case stateReceived := <-stateRx:

			elevatorsMap[stateReceived.Id] = *stateReceived
			statemachines.UpdateLights(elevator, elevatorsMap)

		case order := <-buttons:

			if order.Button == elevio.BT_Cab {
				statemachines.CabOrderFSM(elevator, order.Floor, order.Button, doorTimer)
			} else {
				elevatorsMapCopy := elevatorsMap
				elevatorsMapCopy[elevator.Id].Requests[order.Floor][order.Button] = true

				if elevator.IsOnline {
					Assigner.AssignHallOrders(orderChanTx, elevatorsMapCopy, ackChanRx)
				}

			}

		case newAssignedHallOrders := <-orderChanRx:

			ackChanTx <- elevator.Id
			statemachines.HallOrderFSM(elevator, newAssignedHallOrders, doorTimer, motorFaultTimer)

		case floor := <-floors:

			elevator.Floor = floor
			elevio.SetFloorIndicator(floor)
			motorFaultTimer.Reset(time.Second * 4)

			if elevator.Dirn == elevio.MD_Stop {
				motorFaultTimer.Stop()
			}

			if Orderhandler.ShouldStop(elevator) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevator.Dirn = elevio.MD_Stop
				elevio.SetDoorOpenLamp(true)
				Orderhandler.ClearRequestAtFloor(elevator, doorTimer)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
				motorFaultTimer.Stop()
			}

		case <-doorTimer.C:

			if !elevio.GetObstruction() {
				elevio.SetDoorOpenLamp(false)

				switch {
				case elevator.Behaviour == config.EB_DoorOpen:
					Orderhandler.RequestsChooseDirection(elevator)
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
				peerTxEnable <- true
			}

		case <-motorFaultTimer.C:
			fmt.Println("MOTORFAULT")
			peerTxEnable <- false //let the other peers know that this peer is unavailable

			for elevio.GetFloor() == -1 {
				if elevator.Dirn == elevio.MD_Down {
					elevio.SetMotorDirection(elevio.MD_Down)
				}
				if elevator.Dirn == elevio.MD_Up {
					elevio.SetMotorDirection(elevio.MD_Up)
				}
			}

			if !elevio.GetObstruction() {
				elevator.Dirn = elevio.MD_Stop
				elevio.SetMotorDirection(elevator.Dirn)
				peerTxEnable <- true
				elevio.SetDoorOpenLamp(true)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
			}

		}

		Orderhandler.WriteToBackup(elevator)
	}
}
