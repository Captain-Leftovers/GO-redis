package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Shutting down server...")
		l.Close()
		os.Exit(0)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
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
		if errors.Is(err, io.EOF){
			break
		}
		if err != nil {
			print(err.Error())
			break
		}

		data := string(buff[:n])

		fmt.Println("Received data: ", data)
		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing to connection:", err.Error())
			break
		}

	}
}
