package costfunc

import (
	"Heis/config"
	"Heis/singleElev/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

// Costfunc executes the hall_request_assigner with given input and returns the assignment output
func Costfunc(hallRequests [][2]bool, states map[string]HRAElevState) (map[string][][2]bool, error) {
	hraExecutable := getExecutableName()

	input := HRAInput{
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
		return "hall_request_assigner"
	case "windows":
		return "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
}

func TransformElevatorStates(elevators map[string]config.Elevator) map[string]HRAElevState {
	states := make(map[string]HRAElevState)

	for id, elev := range elevators {
		cabRequests := make([]bool, len(elev.Requests[0])) // Assuming fixed number of floors
		for floor := 0; floor < len(elev.Requests[0]); floor++ {
			// Assuming elev.Requests[floor][BT_Cab] indicates a cab request for that floor
			cabRequests[floor] = elev.Requests[floor][elevio.BT_Cab] > 0
		}

		states[id] = HRAElevState{
			Behavior:    behaviourToString(elev.Behaviour), // Assuming Behaviour has a String method
			Floor:       elev.Floor,
			Direction:   dirnToString(elev.Dirn), // Assuming Dirn has a String method or convert manually
			CabRequests: cabRequests,
		}
	}

	return states
}

func PrepareHallRequests(elevators map[string]config.Elevator) [][2]bool {
	// Assuming a fixed number of floors for simplicity; adjust as necessary
	numFloors := 4
	hallRequests := make([][2]bool, numFloors)

	for _, elev := range elevators {
		for floor := 0; floor < numFloors; floor++ {
			if elev.Requests[floor][0] > 0 { // Check for hall up request
				hallRequests[floor][0] = true
			}
			if elev.Requests[floor][1] > 0 { // Check for hall down request
				hallRequests[floor][1] = true
			}
		}
	}

	return hallRequests
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
