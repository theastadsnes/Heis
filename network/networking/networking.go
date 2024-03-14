package networking

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorhelper"
	"fmt"
)

//

//
// "Heis/statemachines"
// "fmt"
//

func Network(elevator *config.Elevator, elevatorsMap map[string]config.Elevator, hardware config.Hardwarechannels, network config.Networkchannels, AssignHallOrders chan elevio.ButtonEvent, localElevatorHalls chan *config.AssignmentResults) {
	for {
		select {
		case stateReceived := <-network.StateRx:

			elevatorsMap[stateReceived.Id] = *stateReceived
			elevatorhelper.UpdateHallLights(elevator, elevatorsMap)
		case hallOrder := <-AssignHallOrders:
			elevatorsMapCopy := elevatorsMap
			elevatorsMapCopy[elevator.Id].Requests[hallOrder.Floor][hallOrder.Button] = true
			assigner.AssignHallOrders(network.OrderChanTx, elevatorsMapCopy, network.AckChanRx)
		case newAssignedHallOrders := <-network.OrderChanRx:
			fmt.Println("Mottatt nye ordra fra costfunksjon")
			network.AckChanTx <- elevator.Id
			localElevatorHalls <- newAssignedHallOrders
			// statemachines.HallOrderFSM(elevator, newAssignedHallOrders, doorTimer, motorFaultTimer)

		}

	}

}
