package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"syscall"
)

func main() {
	listener, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Fatal("Error setting up TCP server:", err)
	}
	defer listener.Close()

	for {
		log.Println("Listening on port 3333")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue // Continue instead of exiting on accept error
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer log.Println("Closing connection.")

	reader := bufio.NewReader(conn)

	for {
		request, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected gracefully.")
				break
			}
			
			// Handle connection reset by peer errors
			if netErr, ok := err.(*net.OpError); ok && netErr.Err.Error() == "read: connection reset by peer" {
				log.Println("Connection reset by peer.")
				break
			}

			// Handle syscall-level connection resets (platform-dependent)
			if pathErr, ok := err.(*os.PathError); ok && pathErr.Err == syscall.ECONNRESET {
				log.Println("Connection reset by peer (syscall level).")
				break
			}

			log.Printf("Failed to read data: %v", err)
			break
		}

		log.Printf("Received request: %s", request)

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: 6\r\n" +
			"Connection: close\r\n\r\n" +
			"Hello bro!\n"

		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
			break
		}
	}
}
