package core

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"
)

type CoreServer struct {
	Host    string
	Port    int
	f       *os.File
	clients *clientManager
}

func NewCoreServer(host string, port int, f *os.File) *CoreServer {
	return &CoreServer{
		Host:    host,
		Port:    port,
		clients: newClientManager(),
		f:       f,
	}
}

func BindAndListen(ctx context.Context, s *CoreServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on %s:%d...\n", s.Host, s.Port)

	go func() {
		<-ctx.Done()
		fmt.Println("Server shutting down gracefully...")
		listener.Close()
		s.f.Close()
	}()

	clientID := int32(0)
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Println("Error accepting connection:", err)
			}
			continue
		}

		atomic.AddInt32(&clientID, 1)
		s.clients.registerClient(int(clientID), conn)
		go handleConnection(int(clientID), conn, s)
	}
}

func handleConnection(clientID int, conn net.Conn, s *CoreServer) {
	defer func() {
		conn.Close()
		s.clients.unregisterClient(clientID)
		fmt.Println("Client disconnected:", conn.RemoteAddr())
	}()

	fmt.Println("New client connected:", conn.RemoteAddr())
	conn.Write([]byte("Welcome to CoolDB!\n"))
	time.Sleep(5 * time.Second)
}
