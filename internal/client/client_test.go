package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/rushikeshg25/cool-wire/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type testWireServer struct {
	wire.UnimplementedWireServiceServer
	sendQuery func(context.Context, *wire.Query) (*wire.Response, error)
}

func (s *testWireServer) SendQuery(ctx context.Context, query *wire.Query) (*wire.Response, error) {
	return s.sendQuery(ctx, query)
}

func connectTestClient(t *testing.T, server wire.WireServiceServer, queryTimeout time.Duration) *Client {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	wire.RegisterWireServiceServer(grpcServer, server)
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	dialer := func(context.Context, string) (net.Conn, error) { return listener.Dial() }
	client, err := connect(context.Background(), "passthrough:///test", Config{
		ConnectionTimeout: time.Second,
		QueryTimeout:      queryTimeout,
	}, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
	if err != nil {
		t.Fatalf("connect() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestExecuteReturnsTransportIndependentResult(t *testing.T) {
	client := connectTestClient(t, &testWireServer{sendQuery: func(ctx context.Context, query *wire.Query) (*wire.Response, error) {
		if got, want := query.GetQuery(), "SELECT 1"; got != want {
			t.Errorf("query = %q, want %q", got, want)
		}
		return &wire.Response{Response: "one"}, nil
	}}, time.Second)

	result, err := client.Execute(context.Background(), "  SELECT 1  ")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := result.Text, "one"; got != want {
		t.Errorf("result text = %q, want %q", got, want)
	}
}

func TestExecuteTranslatesGRPCError(t *testing.T) {
	client := connectTestClient(t, &testWireServer{sendQuery: func(context.Context, *wire.Query) (*wire.Response, error) {
		return nil, status.Error(codes.NotFound, "table missing")
	}}, time.Second)

	_, err := client.Execute(context.Background(), "SELECT * FROM missing")
	var queryError *QueryError
	if !errors.As(err, &queryError) {
		t.Fatalf("error = %v, want *QueryError", err)
	}
	if queryError.Code != "NotFound" || queryError.Message != "table missing" {
		t.Errorf("query error = %#v", queryError)
	}
}

func TestExecuteHonorsTimeout(t *testing.T) {
	client := connectTestClient(t, &testWireServer{sendQuery: func(ctx context.Context, query *wire.Query) (*wire.Response, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}, 10*time.Millisecond)

	_, err := client.Execute(context.Background(), "SELECT 1")
	var queryError *QueryError
	if !errors.As(err, &queryError) || queryError.Code != "DeadlineExceeded" {
		t.Fatalf("error = %v, want DeadlineExceeded QueryError", err)
	}
}
