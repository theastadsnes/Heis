package elevatorhelper

import (
	"Heis/config"
	"Heis/driver/elevio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func WriteCabCallsToBackup(elevator *config.Elevator) {
	filename := "orderhandler/cabOrder.txt"
	f, err := os.Create(filename)
	if err != nil {
		return
	}

	caborders := make([]bool, config.NumFloors)

	for floors := range elevator.Requests {
		caborders[floors] = elevator.Requests[floors][2]
	}

	cabordersString := strings.Trim(fmt.Sprint(caborders), "[]")
	_, err = f.WriteString(cabordersString)
	if err != nil {
		return
	}

	defer f.Close()
}

func ReadCabCallsFromBackup(buttons chan elevio.ButtonEvent) {
	filename := "orderhandler/cabOrder.txt"
	f, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	caborders := make([]bool, 0)

	cabOrders := strings.Split(string(f), " ")
	for _, order := range cabOrders {
		result, _ := strconv.ParseBool(order)
		caborders = append(caborders, result)
	}

	time.Sleep(20 * time.Millisecond)
	for floor, order := range caborders {
		if order {
			backupOrder := elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
			buttons <- backupOrder
			time.Sleep(20 * time.Millisecond)
		}
	}
}
