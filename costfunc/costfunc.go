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

type HallRequestAssignment struct {
	ID           string
	UpRequests   []bool
	DownRequests []bool
}

type AssignmentResults struct {
	Assignments []HallRequestAssignment
}

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
		cabRequests := make([]bool, len(elev.Requests[0]))
		for floor := 0; floor < len(elev.Requests[0]); floor++ {
			cabRequests[floor] = elev.Requests[floor][elevio.BT_Cab] > 0
		}

		states[id] = HRAElevState{
			Behavior:    behaviourToString(elev.Behaviour),
			Floor:       elev.Floor,
			Direction:   dirnToString(elev.Dirn),
			CabRequests: cabRequests,
		}
	}

	return states
}

func PrepareHallRequests(elevators map[string]config.Elevator) [][2]bool {
	numFloors := 4
	hallRequests := make([][2]bool, numFloors)

	for _, elev := range elevators {
		for floor := 0; floor < numFloors; floor++ {
			if elev.Requests[floor][0] > 0 {
				hallRequests[floor][0] = true
			}
			if elev.Requests[floor][1] > 0 {
				hallRequests[floor][1] = true
			}
		}
	}

	return hallRequests
}

func GetRequestStruct(hallRequests [][2]bool, states map[string]HRAElevState) AssignmentResults {
	output, err := Costfunc(hallRequests, states)
	if err != nil {
		fmt.Println("Error calling Costfunc:", err)
	}
	var requeststruct AssignmentResults

	for id, floors := range output {
		var upRequests, downRequests []bool
		for _, floor := range floors {
			upRequests = append(upRequests, floor[0])
			downRequests = append(downRequests, floor[1])
		}
		requeststruct.Assignments = append(requeststruct.Assignments, HallRequestAssignment{
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
