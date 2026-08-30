package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/rushikeshg25/cool-wire/wire"
	"github.com/rushikeshg25/coolDb/internal/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CoreServer struct {
	Host     string
	Port     int
	database *database.Engine
}

type CoreServerGRPC struct {
	wire.UnimplementedWireServiceServer
	executor queryExecutor
}

type queryExecutor interface {
	Execute(query string) (database.Result, error)
}

func NewCoreServer(host string, port int, engine *database.Engine) *CoreServer {
	return &CoreServer{
		Host:     host,
		Port:     port,
		database: engine,
	}
}

func BindAndListen(ctx context.Context, s *CoreServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	wire.RegisterWireServiceServer(grpcServer, &CoreServerGRPC{executor: s.database})
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			grpcServer.GracefulStop()
		case <-stopped:
		}
	}()
	defer close(stopped)

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	return nil
}

func (s *CoreServerGRPC) SendQuery(ctx context.Context, query *wire.Query) (*wire.Response, error) {
	if query == nil || strings.TrimSpace(query.GetQuery()) == "" {
		return nil, status.Error(codes.InvalidArgument, "query cannot be empty")
	}
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}

	result, err := s.executor.Execute(query.GetQuery())
	if err != nil {
		return nil, databaseStatus(err)
	}
	return &wire.Response{Response: result.Format()}, nil
}

func databaseStatus(err error) error {
	var databaseError *database.Error
	if !errors.As(err, &databaseError) {
		return status.Error(codes.Internal, "internal database error")
	}

	code := codes.Internal
	switch databaseError.Code {
	case database.CodeSyntax, database.CodeType:
		code = codes.InvalidArgument
	case database.CodeAlreadyExists:
		code = codes.AlreadyExists
	case database.CodeNotFound:
		code = codes.NotFound
	case database.CodeConstraint:
		code = codes.FailedPrecondition
	case database.CodeStorage:
		code = codes.Internal
	}
	return status.Error(code, databaseError.Message)
}
