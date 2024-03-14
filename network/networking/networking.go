package networking

import (
	"Heis/assigner"
	"Heis/config"
	"Heis/elevatorutilities"
	"time"
)

func Networking(elevator *config.Elevator, elevatorsMap map[string]config.Elevator, localElevatorChannels config.LocalElevChannels, network config.Networkchannels) {
	for {
		select {
		case stateReceived := <-network.StateRx:

			elevatorsMap[stateReceived.Id] = *stateReceived
			elevatorutilities.UpdateHallLights(elevator, elevatorsMap)

		case hallOrder := <-localElevatorChannels.AssignHallOrders:
			elevatorsMapCopy := elevatorsMap
			elevatorsMapCopy[elevator.Id].Requests[hallOrder.Floor][hallOrder.Button] = true
			assigner.AssignHallOrders(network.OrderChanTx, elevatorsMapCopy, network.AckChanRx)

		case newAssignedHallOrders := <-network.OrderChanRx:
			network.AckChanTx <- elevator.Id
			localElevatorChannels.HallOrders <- newAssignedHallOrders

		}

	}

}

func SendElevatorStates(stateTx chan *config.Elevator, elevator *config.Elevator) {
	for {
		stateTx <- elevator
		time.Sleep(200 * time.Millisecond)
	}
}
