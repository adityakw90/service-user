package testutil

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	devicegrpc "github.com/adityakw90/service-user-proto/gen/go/device"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	user_filegrpc "github.com/adityakw90/service-user-proto/gen/go/user_file"
)

// TestGRPCClient holds gRPC client connections.
type TestGRPCClient struct {
	conn           *grpc.ClientConn
	UserClient     usergrpc.UserServiceClient
	AuthClient     authgrpc.AuthServiceClient
	DeviceClient   devicegrpc.DeviceServiceClient
	UserFileClient user_filegrpc.UserFileServiceClient
}

// NewTestGRPCClient creates gRPC clients connected to the test server.
func NewTestGRPCClient(serverAddr string) (*TestGRPCClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &TestGRPCClient{
		conn:           conn,
		UserClient:     usergrpc.NewUserServiceClient(conn),
		AuthClient:     authgrpc.NewAuthServiceClient(conn),
		DeviceClient:   devicegrpc.NewDeviceServiceClient(conn),
		UserFileClient: user_filegrpc.NewUserFileServiceClient(conn),
	}, nil
}

// Close closes the client connection.
func (c *TestGRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
