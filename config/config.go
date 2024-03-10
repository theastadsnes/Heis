package config

import (
	"Heis/network/localip"
	"Heis/singleElev/elevio"
	"flag"
	"fmt"
	"os"
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
	IsOnline  bool
	Config    struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDurationS   float64
	}
}

type RequestState int

const (
	None      RequestState = 0
	Order     RequestState = 1
	Confirmed RequestState = 2
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

	floor := 0
	for elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)
	}

	elevio.SetMotorDirection(elevio.MD_Stop)
	floor = elevio.GetFloor()

	return Elevator{Id: id,
		Floor:     floor,
		Dirn:      elevio.MD_Stop,
		Requests:  requests,
		Behaviour: EB_Idle,
		IsOnline:  true,
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDurationS   float64
		}{
			ClearRequestVariant: 0,
			DoorOpenDurationS:   3.0,
		},
	}
}

func InitId() string {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("%s-%d", localIP, os.Getpid())
	}
	return id
}
