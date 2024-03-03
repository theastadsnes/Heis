package assigner

import (
	"Heis/config"
	"Heis/singleElev/elevio"
	"Heis/costfunc"
	"fmt"
)

func Assigner(stateRx chan *config.Elevator, buttons chan elevio.ButtonEvent, cabOrder chan *elevio.ButtonEvent){
	ElevatorsMap := make(map[string]config.Elevator)
			
	for{
		select{
		case stateReceived := <- stateRx:
			ElevatorsMap[stateReceived.Id] = *stateReceived
		case order:= <- buttons:
			if order.Button == 2{
				cabOrder <- &order
			}else if order.Button == 0 || order.Button == 1{
				transStates := costfunc.TransformElevatorStates(ElevatorsMap)
				hallRequests := costfunc.PrepareHallRequests(ElevatorsMap)
				newOrders := costfunc.Costfunc(hallRequests, transStates)
				fmt.Print(newOrders)
				//sende newOrders over en kanal, eks orderChan an typen egendefinert struct
			}
		}
	
		}
}