/**
 * @file main.go
 * @brief Entry point for the elevator control program.
 */

package main


import (
	//"Heis/singleElev/elevio"
	//"Heis/singleElev/fsm"
	//"Heis/network/bcast"
	//"Heis/network/localip"
	"Heis/network/conn"
	"Heis/config"
	"fmt"
	"time"
	"net"
)

 func sending(channel chan config.LocalElevatorState){
	fmt.Printf("Hei, kommer inn i func")
	for {
		state := config.InitElevState("Elevator1")
		fmt.Print("Sender state: %+v\n", state)

		channel <- state

		time.Sleep(1 * time.Second) // Juster frekvensen etter behov
	}
}

// func startUDPServer() {
//     addr := net.UDPAddr{
//         Port: 15657,
//         IP:   net.ParseIP("127.0.0.1"),
//     }
//     conn, err := net.ListenUDP("udp", &addr)
//     if err != nil {
//         fmt.Println("Failed to start server:", err)
//         return
//     }
//     defer conn.Close()

//     buffer := make([]byte, 1024)
//     for {
//         n, _, err := conn.ReadFromUDP(buffer)
//         if err != nil {
//             fmt.Println("Failed to read:", err)
//             continue
//         }
//         fmt.Printf("Server received: %s\n", string(buffer[:n]))
//     }
// }
func main() {
	// Define the number of floors in the building
	// numFloors := 4

	// // Initialize elevator I/O
	// elevio.Init("localhost:15657", numFloors)

	// // Create channels for elevator I/O events
	// drv_buttons := make(chan elevio.ButtonEvent)
	// drv_floors := make(chan int)
	// drv_obstr := make(chan bool)
	// drv_stop := make(chan bool)

	// // Create a timer for the door
	// doorTimer := time.NewTimer(time.Duration(3) * time.Second)

	// // Start polling for elevator I/O events
	// go elevio.PollButtons(drv_buttons)
	// go elevio.PollFloorSensor(drv_floors)
	// go elevio.PollObstructionSwitch(drv_obstr)
	// go elevio.PollStopButton(drv_stop)

	// // Start the finite state machine for elevator control
	// fsm.Fsm(drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer, numFloors)


	// const port = 50000
	// elevatorStateChan := make(chan config.LocalElevatorState)
	
	// go bcast.Transmitter(port, elevatorStateChan)
	
    // go sending(elevatorStateChan) // Endret til å kjøre asynkront som en gorutine

    // // Ventemekanisme (for eksempel en uendelig løkke eller select{})
    // select {}

	// fmt.Print(localip.LocalIP())

	conn := conn.DialBroadcastUDP(30000)
    defer conn.Close()

    for {
        // Send en enkel tekststreng som en melding
        message := "Hei fra sender"
        _, err := conn.WriteTo([]byte(message), &net.UDPAddr{
            IP:   net.ParseIP("10.22.74.118"),
            Port: 15657,
        })
        if err != nil {
            fmt.Println("Feil ved sending av melding:", err)
            return
        }
        fmt.Println("Melding sendt:", message)
        time.Sleep(2 * time.Second)
    }

	
}
