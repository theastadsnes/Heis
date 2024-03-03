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
		fmt.Print("------------------------------------------------NICE")
		select {
		case stateReceived := <-stateRx:
			ElevatorsMap[stateReceived.Id] = *stateReceived
			fmt.Print("*******************************", ElevatorsMap)

			
		case orders := <-buttons:
			fmt.Print("---------HALLA-------")
			if orders.Button == 2 {
				cabOrder <- &orders
			} else if orders.Button == 0 || orders.Button == 1 {
				transStates := costfunc.TransformElevatorStates(ElevatorsMap)
				hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
				newOrders, err := costfunc.Costfunc(hallRequests, transStates)

				if err != nil {
					fmt.Print("Panic")
				}
				fmt.Print("NEW:", newOrders)
				//sende newOrders over en kanal, eks orderChan an typen egendefinert struct
			}
		}

	}
}
