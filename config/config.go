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
	NumButtons = 3
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	Id        string
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [][]bool
	Behaviour ElevatorBehaviour
	IsOnline  bool
}

func InitElevState(id string) Elevator {
	requests := make([][]bool, NumFloors)
	for floor := range requests {
		requests[floor] = make([]bool, NumButtons)
	}

	for elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	floor := elevio.GetFloor()

	return Elevator{Id: id,
		Floor:     floor,
		Dirn:      elevio.MD_Stop,
		Requests:  requests,
		Behaviour: EB_Idle,
		IsOnline:  true,
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
