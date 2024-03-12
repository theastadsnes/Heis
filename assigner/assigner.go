package assigner

import (
	"Heis/config"
	"Heis/driver/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

func AssignHallOrders(orderChanTx chan *config.AssignmentResults, ElevatorsMap map[string]config.Elevator, ackChanRx chan string) {

	transStates := TransformElevatorStates(ElevatorsMap)
	fmt.Println("-----Transformed states-----", transStates)
	hallRequests := PrepareHallRequests(ElevatorsMap)
	newOrders := GetRequestStruct(hallRequests, transStates)

	go WaitForAllACKs(orderChanTx, ElevatorsMap, ackChanRx, newOrders)

}

func Costfunc(hallRequests [][2]bool, states map[string]config.HRAElevState) (map[string][][2]bool, error) {
	hraExecutable := getExecutableName()

	input := config.HRAInput{
		HallRequests: hallRequests,
		States:       states,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal error: %v", err)
	}

	ret, err := exec.Command("../Heis/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec.Command error: %v, Output: %s", err, string(ret))
	}

	var output map[string][][2]bool
	if err = json.Unmarshal(ret, &output); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error: %v", err)
	}

	return output, nil
}

func getExecutableName() string {
	switch runtime.GOOS {
	case "linux":
		return "executables/hall_request_assigner"
	case "windows":
		return "executables/hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
}

func TransformElevatorStates(elevators map[string]config.Elevator) map[string]config.HRAElevState {
	states := make(map[string]config.HRAElevState)

	for id, elev := range elevators {
		cabRequests := make([]bool, len(elev.Requests))
		for floor := 0; floor < len(elev.Requests[0]); floor++ {

			cabRequests[floor] = elev.Requests[floor][elevio.BT_Cab]

		}

		states[id] = config.HRAElevState{
			Behavior:    behaviourToString(elev.Behaviour),
			Floor:       elev.Floor,
			Direction:   dirnToString(elev.Dirn),
			CabRequests: cabRequests,
		}
	}

	return states
}

func PrepareHallRequests(elevators map[string]config.Elevator) [][2]bool {
	hallRequests := make([][2]bool, config.NumFloors)

	for _, elev := range elevators {
		for floor := 0; floor < config.NumFloors; floor++ {
			if elev.Requests[floor][0] {
				hallRequests[floor][0] = true
			}
			if elev.Requests[floor][1] {
				hallRequests[floor][1] = true
			}
		}
	}

	return hallRequests
}

func GetRequestStruct(hallRequests [][2]bool, states map[string]config.HRAElevState) config.AssignmentResults {
	output, err := Costfunc(hallRequests, states)
	if err != nil {
		fmt.Println("Error calling Costfunc:", err)
	}
	var requeststruct config.AssignmentResults

	for id, floors := range output {
		var upRequests, downRequests []bool
		for _, floor := range floors {
			upRequests = append(upRequests, floor[0])
			downRequests = append(downRequests, floor[1])
		}
		requeststruct.Assignments = append(requeststruct.Assignments, config.HallRequestAssignment{
			ID:           id,
			UpRequests:   upRequests,
			DownRequests: downRequests,
		})
	}

	return requeststruct
}

func behaviourToString(behaviour config.ElevatorBehaviour) string {
	switch behaviour {
	case config.EB_Idle:
		return "idle"
	case config.EB_Moving:
		return "moving"
	case config.EB_DoorOpen:
		return "doorOpen"
	default:
		return "unknown"
	}
}

func dirnToString(dirn elevio.MotorDirection) string {
	switch dirn {
	case elevio.MD_Up:
		return "up"
	case elevio.MD_Down:
		return "down"
	case elevio.MD_Stop:
		return "stop"
	default:
		return "unknown"
	}
}

func WaitForAllACKs(orderChanTx chan *config.AssignmentResults, ElevatorsMap map[string]config.Elevator, ackChanRx chan string, newOrders config.AssignmentResults) {
	drainAckChannel(ackChanRx)
	acksReceived := make(map[string]bool)
	for id := range ElevatorsMap {
		acksReceived[id] = false
	}

	for {
		select {
		case orderChanTx <- &newOrders:

		case ackID := <-ackChanRx:

			if _, ok := acksReceived[ackID]; ok {
				acksReceived[ackID] = true

				allAcked := true
				for _, acked := range acksReceived {
					if !acked {
						allAcked = false
						break
					}
				}
				if allAcked {
					return // Stop broadcasting if all ACKs received
				}
			}
		case <-time.After(500 * time.Millisecond):
			fmt.Println("Timeout: Not all acknowledgments received")
			return
		}
	}
}

func drainAckChannel(ackChanRx chan string) {
	for {
		select {
		case <-ackChanRx:
			// An ACK was read from the channel, continue draining.
		default:
			// No more ACKs to read, the channel is now drained.
			return
		}
	}
}
