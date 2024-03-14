package networking

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/driver/elevio"
	"Heis/elevatorhelper"
	"fmt"
	"time"
)


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

		}

	}

}

func SendElevatorStates(stateTx chan *config.Elevator, elevator *config.Elevator) {
	for {
		stateTx <- elevator
		time.Sleep(200 * time.Millisecond)
	}
}
