package assigner

import (
	"Heis/config"
	"Heis/costfunc"
	"Heis/singleElev/elevio"
	"fmt"
)

func Assigner(stateRx chan *config.Elevator, buttons chan elevio.ButtonEvent, cabOrder chan *elevio.ButtonEvent) {
	ElevatorsMap := make(map[string]config.Elevator)

	for {
		//fmt.Println("------------------------------------------------NICE")
		select {
		case stateReceived := <-stateRx:
			ElevatorsMap[stateReceived.Id] = *stateReceived
			//fmt.Println("******************************STATEMAP", ElevatorsMap)

		case orders := <-buttons:
			fmt.Println("---------KNAPP TRYKKET-------")
			if orders.Button == 2 {
				cabOrder <- &orders
			} else if orders.Button == 0 || orders.Button == 1 {
				transStates := costfunc.TransformElevatorStates(ElevatorsMap)
				hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
				fmt.Print(costfunc.Costfunc(hallRequests, transStates))
				newOrders := costfunc.AssignOrders(orders, ElevatorsMap)

				// if err != nil {
				// 	fmt.Println("Panic")
				// }

				//sende newOrders over en kanal, eks orderChan an typen egendefinert struct
			}
		}

	}
}
