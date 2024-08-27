package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"


	"github.com/codecrafters-io/redis-starter-go/resp"
)





func main() {

	go resp.StartCleanupRoutine()

	// PORT := 6379
	// TODO use netcat to send commands and develop the Redis protocol parser

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Flag to indicate whether the server is running
	running := true

	go func() {
		<-sigChan
		fmt.Println("Shutting down server...")
		l.Close()
		running = false
	}()

	for running {

		conn, err := l.Accept()

		if err != nil {
			// Check if the error is due to the listener being closed
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				fmt.Println("Listener closed, stopping server...")
				break
			}

			fmt.Println("Error accepting connection: ", err.Error())
			break
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buff := make([]byte, 1024)

		n, err := conn.Read(buff)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			print(err.Error())
			break
		}

		data := buff[:n] // buff byte slice

		fmt.Println("Received data: ", data)
		answer, err := resp.ExecuteRespData(data)
		if err != nil {
			fmt.Println("Error executing command: ", err.Error())
			break
		}

		_, err = conn.Write(answer)
		if err != nil {
			fmt.Println("Error writing to connection:", err.Error())
			break
		}
	}
}
