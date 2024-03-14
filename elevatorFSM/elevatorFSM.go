package elevatorFSM

import (
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorutilities"
	"Heis/statemachines"
	"fmt"
	"time"
)

func ElevatorFsm(elevator *config.Elevator, doorTimer *time.Timer, motorFaultTimer *time.Timer, localElevatorChannels config.LocalElevChannels, peerTxEnable chan bool) {

	for {
		select {

		case order := <-localElevatorChannels.Drv_buttons:
			if order.Button == elevio.BT_Cab {
				statemachines.CabOrderStateMachine(elevator, order.Floor, order.Button, doorTimer, motorFaultTimer)
			} else {
				if elevator.IsOnline {
					localElevatorChannels.AssignHallOrders <- order
				}
			}

		case hallOrders := <-localElevatorChannels.HallOrders:
			statemachines.HallOrderStateMachine(elevator, hallOrders, doorTimer, motorFaultTimer)

		case floor := <-localElevatorChannels.Drv_floors:
			elevator.Floor = floor
			elevio.SetFloorIndicator(floor)
			motorFaultTimer.Reset(time.Second * 4)

			if elevator.Dirn == elevio.MD_Stop {
				motorFaultTimer.Stop()
			}

			if elevatorutilities.ShouldStop(elevator) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevatorutilities.OpenDoor(elevator, doorTimer)
				motorFaultTimer.Stop()
			}

		case <-doorTimer.C:

			if !elevio.GetObstruction() {
				elevio.SetDoorOpenLamp(false)
				switch {
				case elevator.Behaviour == config.EB_DoorOpen:
					elevatorutilities.RequestsChooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Dirn)
					if elevator.Dirn == elevio.MD_Stop {
						elevator.Behaviour = config.EB_Idle
						elevatorutilities.ClearRequestAtFloor(elevator)
					} else {
						elevator.Behaviour = config.EB_Moving
						motorFaultTimer.Reset(time.Second * 4)
						elevatorutilities.ClearRequestAtFloor(elevator)
					}

				}
			} else {
				motorFaultTimer.Reset(time.Second * 4)
			}

		case obstruction := <-localElevatorChannels.Drv_obstr:
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
			elevatorutilities.GoToValidFloor(elevator)

			time.AfterFunc(time.Second*2, func() {
				if !elevio.GetObstruction() {
					peerTxEnable <- true
					elevatorutilities.OpenDoor(elevator, doorTimer)
				}
			})
		}

		elevatorutilities.WriteCabCallsToBackup(elevator)
	}
}
