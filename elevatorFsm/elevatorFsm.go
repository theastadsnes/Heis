package elevatorFsm

import (
	//"Heis/assigner"
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorhelper"
	"Heis/statemachines"
	"fmt"
	"time"
)

func ElevatorFsm(elevator *config.Elevator, doorTimer *time.Timer, motorFaultTimer *time.Timer, hardware config.Hardwarechannels, peerTxEnable chan bool, AssignHallOrders chan elevio.ButtonEvent, localElevatorHalls chan *config.AssignmentResults) {

	for {
		select {
		
		case order := <-hardware.Drv_buttons:
			if order.Button == elevio.BT_Cab {
				statemachines.CabOrderStateMachine(elevator, order.Floor, order.Button, doorTimer, motorFaultTimer)
			} else {
				if elevator.IsOnline {
					AssignHallOrders <- order
				}
			}

		case hallOrders := <-localElevatorHalls:
			statemachines.HallOrderStateMachine(elevator, hallOrders, doorTimer, motorFaultTimer)

		case floor := <-hardware.Drv_floors:
			elevator.Floor = floor
			elevio.SetFloorIndicator(floor)
			motorFaultTimer.Reset(time.Second * 4)
		
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
						elevatorhelper.ClearRequestAtFloor(elevator)
					} else {
						elevator.Behaviour = config.EB_Moving
						motorFaultTimer.Reset(time.Second * 4)
						elevatorhelper.ClearRequestAtFloor(elevator)
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
