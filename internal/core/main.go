package core

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/rushikeshg25/cool-wire/wire"
	"google.golang.org/grpc"
)

type CoreServer struct {
	Host    string
	Port    int
	f       *os.File
	clients *clientManager
}

type CoreServerGRPC struct {
	wire.UnimplementedWireServiceServer
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

	go func() {
		<-ctx.Done()
		fmt.Println("Server shutting down gracefully...")
		listener.Close()
		s.f.Close()
	}()

	grpcServer := grpc.NewServer()
	wire.RegisterWireServiceServer(grpcServer, &CoreServerGRPC{})
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	return nil
}

func (s *CoreServerGRPC) SendQuery(ctx context.Context, query *wire.Query) (*wire.Response, error) {
	fmt.Println("Query received:", query.Query)
	return &wire.Response{Response: "Hello world"}, nil
}
