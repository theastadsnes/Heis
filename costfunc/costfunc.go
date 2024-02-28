package costfunc

import (
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
	case "linux", "darwin":
		return "hall_request_assigner"
	case "windows":
		return "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
}


