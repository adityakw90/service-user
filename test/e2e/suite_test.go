package e2e

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/test/util"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/stretchr/testify/require"
)

// setupE2ETest creates and starts test infrastructure (services, gRPC server, client).
// Returns a cleanup function that should be called in defer.
func setupE2ETest(t *testing.T) (*util.TestServices, *util.TestGRPCClient, func()) {
	t.Helper()

	ctx := context.Background()

	// Setup test services
	testServices, err := util.SetupTestServices(ctx)
	require.NoError(t, err)

	// Start gRPC server
	grpcServer, err := util.NewTestGRPCServer(testServices)
	require.NoError(t, err)

	// Create gRPC clients
	grpcClient, err := util.NewTestGRPCClient(grpcServer.Addr())
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		grpcClient.Close()
		grpcServer.Close()
		util.TeardownTestServices(testServices)
	}

	// Clean up test data before each test
	err = util.CleanupTestData(ctx, testServices)
	require.NoError(t, err)

	return testServices, grpcClient, cleanup
}

// createTestUser creates a test user via gRPC and returns the UID.
func createTestUser(t *testing.T, grpcClient *util.TestGRPCClient, username, email, password string) string {
	t.Helper()
	ctx := context.Background()
	req := &usergrpc.AddRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	resp, err := grpcClient.UserClient.Add(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	return resp.Uid
}

// deactivateUser sets a user's status to inactive via gRPC.
func deactivateUser(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
	t.Helper()
	ctx := context.Background()
	inactive := int32(model.UserStatusInactive)
	_, err := grpcClient.UserClient.Update(ctx, &usergrpc.UpdateRequest{
		Uid:    uid,
		Status: &inactive,
	})
	require.NoError(t, err)
}

// deleteUser soft-deletes a user via gRPC.
func deleteUser(t *testing.T, grpcClient *util.TestGRPCClient, uid string) {
	t.Helper()
	ctx := context.Background()
	_, err := grpcClient.UserClient.Delete(ctx, &usergrpc.DeleteRequest{
		Uid: uid,
	})
	require.NoError(t, err)
}
