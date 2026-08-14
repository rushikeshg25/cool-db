package client

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/rushikeshg25/cool-wire/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	defaultHost              = "localhost"
	defaultPort              = 3040
	defaultConnectionTimeout = 5 * time.Second
	defaultQueryTimeout      = 30 * time.Second
)

type Config struct {
	Host              string
	Port              int
	ConnectionTimeout time.Duration
	QueryTimeout      time.Duration
}

type Result struct {
	Text string
}

type QueryError struct {
	Code    string
	Message string
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type Client struct {
	connection   *grpc.ClientConn
	wireClient   wire.WireServiceClient
	queryTimeout time.Duration
}

func Connect(ctx context.Context, config Config) (*Client, error) {
	config = withDefaults(config)
	target := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	return connect(ctx, target, config, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func connect(ctx context.Context, target string, config Config, options ...grpc.DialOption) (*Client, error) {
	config = withDefaults(config)
	connection, err := grpc.NewClient(target, options...)
	if err != nil {
		return nil, fmt.Errorf("create connection to %s: %w", target, err)
	}
	if err := waitUntilReady(ctx, connection, config.ConnectionTimeout); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("connect to CoolDB at %s: %w", target, err)
	}
	return &Client{
		connection:   connection,
		wireClient:   wire.NewWireServiceClient(connection),
		queryTimeout: config.QueryTimeout,
	}, nil
}

func (c *Client) Execute(ctx context.Context, query string) (Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Result{}, fmt.Errorf("query cannot be empty")
	}
	queryContext, cancel := context.WithTimeout(ctx, c.queryTimeout)
	defer cancel()
	response, err := c.wireClient.SendQuery(queryContext, &wire.Query{Query: query})
	if err != nil {
		grpcStatus := status.Convert(err)
		return Result{}, &QueryError{Code: grpcStatus.Code().String(), Message: grpcStatus.Message()}
	}
	return Result{Text: response.GetResponse()}, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}

func withDefaults(config Config) Config {
	if config.Host == "" {
		config.Host = defaultHost
	}
	if config.Port == 0 {
		config.Port = defaultPort
	}
	if config.ConnectionTimeout <= 0 {
		config.ConnectionTimeout = defaultConnectionTimeout
	}
	if config.QueryTimeout <= 0 {
		config.QueryTimeout = defaultQueryTimeout
	}
	return config
}

func waitUntilReady(ctx context.Context, connection *grpc.ClientConn, timeout time.Duration) error {
	waitContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	connection.Connect()
	for {
		state := connection.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if !connection.WaitForStateChange(waitContext, state) {
			return waitContext.Err()
		}
	}
}
