package packetloss

import (
	//"Heis/costfunc"
	"Heis/config"
	"fmt"
	"time"
)

//send ack
// if not ack within time period
// resend neworders 5 times
// if not run cabOrderfsm

func WaitForAllAcks(ElevatorsMap map[string]config.Elevator, ackChan chan string, localElev *config.Elevator) {
	retryLimit := 5

	// Check each elevator for an acknowledgement
	for id := range ElevatorsMap {
		for try := 0; try < retryLimit; try++ {
			select {
			case ackID := <-ackChan: // Wait for ACK
				if ackID == id {
					break // Move on to the next elevator
				}
			case <-time.After(time.Millisecond * 200): // Timeout
				if try == retryLimit-1 {
					fmt.Printf("Elevator %s did not acknowledge after %d attempts\n", id, retryLimit)
				}
			}
		}
	}
}
