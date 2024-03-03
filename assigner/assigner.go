package assigner

import (
	"Heis/config"
	"Heis/singleElev/elevio"
)

func Assigner(stateRx chan *config.Elevator, buttons chan elevio.ButtonEvent, cabOrder chan elevio.ButtonEvent){
	ElevatorsMap := make(map[string]config.Elevator)
			
	for{
		select{
		case stateReceived := <- stateRx:
			ElevatorsMap[stateReceived.Id] = stateReceived
		case order:= <- buttons:
			if order.ButtonType == 2{
				cabOrder <- order
			}else if order.Button == elevio.ButtonType.BT_HallDown || elevio.ButtonType.BT_HallUp
			{
				
			}
		}
	
		}
}