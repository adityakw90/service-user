package e2e

import (
	"context"
	"testing"

	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	filegrpc "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	testutil "github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

// setupE2ETest creates and starts test infrastructure (services, gRPC server, client).
// Returns a cleanup function that should be called in defer.
func setupE2ETest(t *testing.T) (*testutil.TestServices, *testutil.TestGRPCClient, func()) {
	t.Helper()

	ctx := context.Background()

	// Setup test services
	testServices, err := testutil.SetupTestServices(t, ctx)
	require.NoError(t, err)

	// Start gRPC server
	grpcServer, err := testutil.NewTestGRPCServer(testServices)
	require.NoError(t, err)

	// Create gRPC clients
	grpcClient, err := testutil.NewTestGRPCClient(grpcServer.Addr())
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		grpcClient.Close()
		grpcServer.Close()
		// testutil.TeardownTestServices(testServices)
	}

	// Clean up test data before each test
	// err = testutil.CleanupTestData(ctx, testServices)
	// require.NoError(t, err)

	return testServices, grpcClient, cleanup
}

// createTestUser creates a test user via gRPC and returns the UID.
func createTestUser(t *testing.T, grpcClient *testutil.TestGRPCClient, username, email, password string) string {
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
func deactivateUser(t *testing.T, grpcClient *testutil.TestGRPCClient, uid string) {
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
func deleteUser(t *testing.T, grpcClient *testutil.TestGRPCClient, uid string) {
	t.Helper()
	ctx := context.Background()
	_, err := grpcClient.UserClient.Delete(ctx, &usergrpc.DeleteRequest{
		Uid: uid,
	})
	require.NoError(t, err)
}

// createTestFile creates a test file via gRPC.
func createTestFile(t *testing.T, grpcClient *testutil.TestGRPCClient, userUID, fileName string, fileData []byte, isPublic bool) string {
	t.Helper()
	ctx := context.Background()
	req := &filegrpc.AddRequest{
		UserUid:  userUID,
		Name:     fileName,
		Filename: fileName,
		Filedata: fileData,
		Public:   &isPublic,
	}

	resp, err := grpcClient.UserFileClient.Add(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	return resp.Uid
}

// generateTestFileData generates test file data of the specified size.
func generateTestFileData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}
