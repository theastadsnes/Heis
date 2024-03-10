package packetloss

import (
	//"Heis/costfunc"
	"Heis/config"
	"Heis/costfunc"
	"fmt"
	"time"
)

//send ack
// if not ack within time period
// resend neworders 5 times
// if not run cabOrderfsm

func WaitForAllACKs(orderChanTx chan *costfunc.AssignmentResults, ElevatorsMap map[string]config.Elevator, ackChanRx chan string, newOrders costfunc.AssignmentResults) {
	acksReceived := make(map[string]bool)
	for id := range ElevatorsMap {
		acksReceived[id] = false // Initially, no ACKs received
	}

	// Function to broadcast orders
	broadcastOrders := func() {
		for {
			select {
			case orderChanTx <- &newOrders:
			case ackID := <-ackChanRx:
				if _, ok := acksReceived[ackID]; ok {
					acksReceived[ackID] = true // Mark ACK as received
					// Check if ACKs received from all elevators
					allAcked := true
					for _, acked := range acksReceived {
						if !acked {
							allAcked = false
							break
						}
					}
					if allAcked {
						return // Stop broadcasting if all ACKs received
					}
				}
			case <-time.After(200 * time.Millisecond):
				fmt.Println("Timeout: Not all acknowledgments received")
				return
			}
		}
	}

	go broadcastOrders()
}
