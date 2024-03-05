package assigner

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/singleElev/elevio"
	"fmt"
	//"sync"
)

func Assigner(stateRx chan *config.Elevator, buttons chan elevio.ButtonEvent, cabOrder chan *elevio.ButtonEvent, orderChanTx chan *costfunc.AssignmentResults, elevator *config.Elevator) {

	ElevatorsMap := make(map[string]config.Elevator)

	for {
		select {

		case order := <-buttons:
			if order.Button == 2 {
				cabOrder <- &order
			}

			fmt.Print("heieiieieieieieieeiieieieiiieie")

		case stateReceived := <-stateRx:

			ElevatorsMap[stateReceived.Id] = *stateReceived

			transStates := costfunc.TransformElevatorStates(ElevatorsMap)
			hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
			//fmt.Println(costfunc.GetRequestStruct(hallRequests, transStates))
			newOrders := costfunc.GetRequestStruct(hallRequests, transStates)
			orderChanTx <- &newOrders
		}

	}

}

func AssigningHallOrders(orderChanRx chan *costfunc.AssignmentResults, elevator *config.Elevator, fsmOrders chan *elevio.ButtonEvent) {
	for {
		select {
		case orders := <-orderChanRx:
			//fmt.Println(orders)
			for _, assignments := range (*orders).Assignments {
				for floor := 0; floor < config.NumFloors; floor++ {
					if assignments.UpRequests[floor] {
						elevator.Requests[floor][0] = 1

						buttonEvent := elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.BT_HallUp,
						}
						fmt.Println("ASSIGNING HALL ORDER", buttonEvent)
						fsmOrders <- &buttonEvent
					}
					if assignments.DownRequests[floor] {
						elevator.Requests[floor][1] = 1
						buttonEvent := elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.BT_HallDown,
						}
						fsmOrders <- &buttonEvent
					}
				}
			}
		}
	}
}
