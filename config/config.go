package config

import (
	"Heis/driver/elevio"
	"Heis/network/localip"
	"Heis/network/peers"
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

type LocalElevChannels struct {
	Drv_buttons      chan elevio.ButtonEvent
	Drv_floors       chan int
	Drv_obstr        chan bool
	AssignHallOrders chan elevio.ButtonEvent
	HallOrders       chan *AssignmentResults
}

type Networkchannels struct {
	OrderChanRx chan *AssignmentResults
	OrderChanTx chan *AssignmentResults
	StateRx     chan *Elevator
	StateTx     chan *Elevator
	AckChanRx   chan string
	AckChanTx   chan string
}

type Peerchannels struct {
	PeerUpdateCh chan peers.PeerUpdate
	PeerTxEnable chan bool
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
