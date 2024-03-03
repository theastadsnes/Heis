package config

import (
	"Heis/singleElev/elevio"
)

const (
	NumFloors  = 4
	NumButtons = 4
)


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
	Dirn      elevio.MotorDirection
	Requests  [][]int
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

func InitElevState(id string) Elevator {
	requests := make([][]int, 4)
	for floor := range requests {
		requests[floor] = make([]int, 4)
	}
	return Elevator{Id: id,
		Floor:     0,
		Dirn:      elevio.MD_Stop,
		Requests:  requests,
		Behaviour: EB_Idle,
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDurationS   float64
		}{
			ClearRequestVariant: 0,
			DoorOpenDurationS:   3.0,
		},
	}
}
