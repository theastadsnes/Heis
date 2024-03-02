package config

import (
	"Heis/singleElev/elevio"
)

const (
	NumFloors  = 4
	NumButtons = 4
)

var Our_elevator Elevator
var Pair DirnBehaviourPair

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	Id        string
	Floor     int
	NextDest  int
	Dirn      elevio.MotorDirection
	Requests  [4][4]int
	Behaviour ElevatorBehaviour

	Config struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDurationS   float64
	}
}

type RequestState int

const (
	None      RequestState = 0
	Order     RequestState = 1
	Comfirmed RequestState = 2
	Complete  RequestState = 3
)

type LocalElevatorState struct {
	ID       string
	Floor    int
	Dir      ElevatorBehaviour
	Requests [][]RequestState
	Behave   ElevatorBehaviour
}

func InitElevState(id string) LocalElevatorState {
	requests := make([][]RequestState, 4)
	for floor := range requests {
		requests[floor] = make([]RequestState, 3)
	}
	return LocalElevatorState{Requests: requests, ID: id, Floor: 0, Behave: EB_Idle}
}
