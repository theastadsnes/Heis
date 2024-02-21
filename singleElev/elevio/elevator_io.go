/**
 * @file elevio.go
 * @brief Provides functions for interacting with elevator I/O.
 */

package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

/**
 * @brief Initializes the elevator I/O connection.
 * @param addr The address of the elevator server.
 * @param numFloors The total number of floors in the building.
 */
func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

/**
 * @brief Sets the direction of the elevator motor.
 * @param dir The direction of the motor (up, down, or stop).
 */
func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

/**
 * @brief Sets the lamp status for a button.
 * @param button The type of button (hall up, hall down, or cab).
 * @param floor The floor where the button is located.
 * @param value The status of the lamp (on or off).
 */
func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

/**
 * @brief Sets the floor indicator.
 * @param floor The floor to be indicated.
 */
func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

/**
 * @brief Sets the status of the door open lamp.
 * @param value The status of the door open lamp (on or off).
 */
func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

/**
 * @brief Sets the status of the stop button lamp.
 * @param value The status of the stop button lamp (on or off).
 */
func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

/**
 * @brief Polls for button events and sends them to a channel.
 * @param receiver The channel to send button events to.
 */
func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

/**
 * @brief Polls for floor events and sends them to a channel.
 * @param receiver The channel to send floor events to.
 * @return The current floor.
 */
func PollFloorSensor(receiver chan<- int) int {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

/**
 * @brief Polls for stop button events and sends them to a channel.
 * @param receiver The channel to send stop button events to.
 */
func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

/**
 * @brief Polls for obstruction switch events and sends them to a channel.
 * @param receiver The channel to send obstruction switch events to.
 */
func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

/**
 * @brief Reads the status of a button.
 * @param button The type of button (hall up, hall down, or cab).
 * @param floor The floor where the button is located.
 * @return The status of the button.
 */
func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

/**
 * @brief Reads the current floor.
 * @return The current floor or -1 if no floor is detected.
 */
func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

/**
 * @brief Reads the status of the stop button.
 * @return The status of the stop button.
 */
func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

/**
 * @brief Reads the status of the obstruction switch.
 * @return The status of the obstruction switch.
 */
func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

/**
 * @brief Reads data from the elevator server.
 * @param in The input data to send.
 * @return The response from the elevator server.
 */
func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

/**
 * @brief Writes data to the elevator server.
 * @param in The data to send.
 */
func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

/**
 * @brief Converts a boolean value to a byte (0 or 1).
 * @param a The boolean value to convert.
 * @return The byte representation of the boolean value.
 */
func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

/**
 * @brief Converts a byte (0 or 1) to a boolean value.
 * @param a The byte to convert.
 * @return The boolean value.
 */
func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

/**
 * @brief Pauses execution for a specified duration.
 * @param duration The duration to sleep.
 */
func Set_timer(duration time.Duration) {
	time.Sleep(duration * time.Second)
}
