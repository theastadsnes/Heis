package elevatorFsm

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorhelper"
	"Heis/statemachines"
	"fmt"
	"time"
)

func ElevatorFsm(elevator *config.Elevator, doorTimer *time.Timer, motorFaultTimer *time.Timer, numFloors int, elevatorsMap map[string]config.Elevator, hardware config.Hardwarechannels, network config.Networkchannels, peerTxEnable chan bool) {
	elevatorhelper.ClearLights()

	for {
		select {
		case stateReceived := <-network.StateRx:

			elevatorsMap[stateReceived.Id] = *stateReceived
			elevatorhelper.UpdateHallLights(elevator, elevatorsMap)

		case order := <-hardware.Drv_buttons:

			if order.Button == elevio.BT_Cab {
				statemachines.CabOrderFSM(elevator, order.Floor, order.Button, doorTimer, motorFaultTimer)
			} else {

				elevatorsMapCopy := elevatorsMap
				elevatorsMapCopy[elevator.Id].Requests[order.Floor][order.Button] = true

				if elevator.IsOnline {
					assigner.AssignHallOrders(network.OrderChanTx, elevatorsMapCopy, network.AckChanRx)
				}
			}

		case newAssignedHallOrders := <-network.OrderChanRx:
			network.AckChanTx <- elevator.Id
			statemachines.HallOrderFSM(elevator, newAssignedHallOrders, doorTimer, motorFaultTimer)

		case floor := <-hardware.Drv_floors:
			elevator.Floor = floor
			elevio.SetFloorIndicator(floor)
			motorFaultTimer.Reset(time.Second * 4)
			fmt.Println(elevator.Dirn)

			if elevator.Dirn == elevio.MD_Stop {
				motorFaultTimer.Stop()
			}

			if elevatorhelper.ShouldStop(elevator) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevatorhelper.OpenDoor(elevator, doorTimer)
				motorFaultTimer.Stop()
			}

		case <-doorTimer.C:

			if !elevio.GetObstruction() {
				elevio.SetDoorOpenLamp(false)
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:
					elevatorhelper.RequestsChooseDirection(elevator)
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

			elevatorhelper.GoToValidFloor(elevator)

			time.AfterFunc(time.Second*2, func() {
				if !elevio.GetObstruction() {
					peerTxEnable <- true
					elevatorhelper.OpenDoor(elevator, doorTimer)
				}
			})

		}

		elevatorhelper.WriteCabCallsToBackup(elevator)
	}
}
