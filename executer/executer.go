/**
 * @file fsm.go
 * @brief Contains the finite state machine (FSM) logic for elevator control.
 */

package executer

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/orderhandler"
	"Heis/statemachines"
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

func Fsm(elevator *config.Elevator, doorTimer *time.Timer, motorFaultTimer *time.Timer, numFloors int, elevatorsMap map[string]config.Elevator, hardware config.Hardwarechannels, network config.Networkchannels, peerTxEnable chan bool) {
	orderhandler.ClearLights()
	//elevatorsMap := make(map[string]config.Elevator)

	for {
		select {
		case stateReceived := <-network.StateRx:

			elevatorsMap[stateReceived.Id] = *stateReceived
			orderhandler.UpdateLights(elevator, elevatorsMap)

		case order := <-hardware.Drv_buttons:
			//peerTxEnable <- true
			if order.Button == 2 {
				statemachines.CabOrderFSM(elevator, order.Floor, order.Button, doorTimer)
			} else {
				//elevator.Requests[order.Floor][order.Button] = 1
				elevatorsMapCopy := elevatorsMap
				elevatorsMapCopy[elevator.Id].Requests[order.Floor][order.Button] = true

				if elevator.IsOnline {
					assigner.AssignHallOrders(network.OrderChanTx, elevatorsMapCopy, network.AckChanRx)
				}

			}

		case newAssignedHallOrders := <-network.OrderChanRx:
			// fmt.Println("ASSIGNING HALL ORDER")
			// fmt.Println(newAssignedHallOrders)
			network.AckChanTx <- elevator.Id

			statemachines.HallOrderFSM(elevator, newAssignedHallOrders, doorTimer, motorFaultTimer)

		case floor := <-hardware.Drv_floors:
			//peerTxEnable <- true
			elevator.Floor = floor

			elevio.SetFloorIndicator(floor)
			motorFaultTimer.Reset(time.Second * 4)
			fmt.Println(elevator.Dirn)

			if elevator.Dirn == elevio.MD_Stop {
				motorFaultTimer.Stop()
			}

			if orderhandler.ShouldStop(elevator) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				orderhandler.ClearRequestAtFloor(elevator, doorTimer)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
				motorFaultTimer.Stop()
			}

		case <-doorTimer.C:

			if !elevio.GetObstruction() {
				elevio.SetDoorOpenLamp(false)
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:
					orderhandler.RequestsChooseDirection(elevator)
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

		case obstruction := <-hardware.Drv_obstr:
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

			if !elevio.GetObstruction() {
				elevator.Dirn = elevio.MD_Stop
				elevio.SetMotorDirection(elevator.Dirn)
				peerTxEnable <- true
				elevio.SetDoorOpenLamp(true)
				elevator.Behaviour = config.EB_DoorOpen
				doorTimer.Reset(time.Duration(3) * time.Second)
			}

		}

		orderhandler.WriteToBackup(elevator)
	}
}
