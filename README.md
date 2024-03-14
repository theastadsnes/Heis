# Elevator Project

## Overview

This project demonstrates an elevator control system designed to control n elevators in parallell across m floors. We utilized Go for its  programming logic and used a peer to peer approach and UDP broadcasting to solve the problem. In the instructions for the code hand-in, we were asked not to include any executables within our submitted files. Consequently, users of this project will need to download the required executable separately from the project's resources. This step is essential for the full functionality of the system.

## Prerequisites

Before you start, ensure you have the following:
- Go installed on your machine (version 1.x or higher).
- Access to the project's resources for downloading the required executable.

## Setup

Follow these steps to set up your environment:

1. Download the necessary executable from the project's resources. For Ubuntu systems, you will need to use the file hall_request_assigner.
2. Place the downloaded executable in the project's root directory. We use this executable in the function getExecutableName() in the file assigner.go.
3. Change the permission of the executable to make it runnable. On an Ubuntu system, you can use the following command:
   ```bash
   chmod +x /path/to/executable
4. Open your terminal or command prompt and run the elevatorserver like this:
   ```bash
   elevatorserver
   
6. Change the port number in the elevio.Init function in the main.go file to match the port number used by the elevatorserver. 

## Compilation and Execution

To compile and run this project, execute the following steps:

1. Open another terminal or command prompt.
2. Navigate to the project's root directory.
3. Run the project using:
   ```bash
   go run main.go

