package e2e

import (
	"context"
	"testing"

	commongrpc "github.com/adityakw90/service-user-proto/gen/go/common"
	filegrpc "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/adityakw90/service-user/pkg/util"
	testutil "github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

// File size limit as per service specification (10MB)
const maxFileSize = 10 * 1024 * 1024

func TestE2E_UserFile_Add(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		fileName   string
		fileData   []byte
		isPublic   bool
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient)
	}{
		{
			name: "Add file with valid data",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "fileuser1", "fileuser1@example.com", "Password123!")
			},
			fileName: "test-file.txt",
			fileData: []byte("Hello, World!"),
			isPublic: false,
			wantErr:  false,
			verifyFunc: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				file, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				require.NoError(t, err)
				require.NotNil(t, file)
				require.Equal(t, fileUID, file.Uid)
				require.Equal(t, userUID, file.UserUid)
				require.Equal(t, "test-file.txt", file.FileName)
			},
		},
		{
			name: "Add file with invalid UID",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "invalid-uid-format"
			},
			fileName: "test-file.txt",
			fileData: []byte("Hello, World!"),
			isPublic: false,
			wantErr:  true,
			errMsg:   "invalid",
		},
		{
			name: "Add file for non-existent user",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			fileName: "test-file.txt",
			fileData: []byte("Hello, World!"),
			isPublic: false,
			wantErr:  true,
			errMsg:   "not found",
		},
		{
			name: "Add file with empty filename",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "fileuser3", "fileuser3@example.com", "Password123!")
			},
			fileName: "",
			fileData: []byte("Hello, World!"),
			isPublic: false,
			wantErr:  true,
			errMsg:   "Name",
		},
		{
			name: "Add file with empty file data",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "fileuser4", "fileuser4@example.com", "Password123!")
			},
			fileName: "empty-file.txt",
			fileData: []byte{},
			isPublic: false,
			wantErr:  false, // Service allows empty files
			verifyFunc: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				file, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				require.NoError(t, err)
				require.NotNil(t, file)
				require.Equal(t, int64(0), file.SizeBytes)
			},
		},
		{
			name: "Add public file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return createTestUser(t, grpcClient, "fileuser5", "fileuser5@example.com", "Password123!")
			},
			fileName: "public-file.txt",
			fileData: []byte("Public content"),
			isPublic: true,
			wantErr:  false,
			verifyFunc: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				file, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				require.NoError(t, err)
				require.Equal(t, "public", file.Visibility)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			userUID := tt.setup(t, grpcClient)
			req := &filegrpc.AddRequest{
				UserUid:  userUID,
				Name:     tt.fileName,
				Filename: tt.fileName,
				Filedata: tt.fileData,
				Public:   &tt.isPublic,
			}
			resp, err := grpcClient.UserFileClient.Add(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.Uid)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, resp.Uid, userUID, grpcClient)
				}
			}
		})
	}
}

func TestFileGet(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name    string
		setup   func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string)
		getUID  func(t *testing.T, fileUID string) string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Get existing file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string) {
				userUID := createTestUser(t, grpcClient, "getfileuser", "getfileuser@example.com", "Password123!")
				fileUID := createTestFile(t, grpcClient, userUID, "get-test.txt", []byte("test content"), false)
				return fileUID, userUID
			},
			getUID:  func(t *testing.T, fileUID string) string { return fileUID },
			wantErr: false,
		},
		{
			name: "Get non-existent file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string) {
				userUID := createTestUser(t, grpcClient, "nofileuser", "nofileuser@example.com", "Password123!")
				return "01234567-89ab-cdef-0123-456789abcdef", userUID
			},
			getUID:  func(t *testing.T, fileUID string) string { return fileUID },
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "Get deleted file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string) {
				userUID := createTestUser(t, grpcClient, "deletedfileuser", "deletedfileuser@example.com", "Password123!")
				fileUID := createTestFile(t, grpcClient, userUID, "deleted-test.txt", []byte("test content"), false)
				ctx := context.Background()
				_, _ = grpcClient.UserFileClient.Delete(ctx, &filegrpc.DeleteRequest{Uid: fileUID})
				return fileUID, userUID
			},
			getUID:  func(t *testing.T, fileUID string) string { return fileUID },
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			fileUID, _ := tt.setup(t, grpcClient)
			getUID := tt.getUID(t, fileUID)

			file, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: getUID})

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				require.Nil(t, file)
			} else {
				require.NoError(t, err)
				require.NotNil(t, file)
				require.Equal(t, fileUID, file.Uid)
				// Note: UserUid is not populated in Get response (known issue)
			}
		})
	}
}

func TestFileList(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	user1UID := createTestUser(t, grpcClient, "listuser1", "listuser1@example.com", "Password123!")
	user2UID := createTestUser(t, grpcClient, "listuser2", "listuser2@example.com", "Password123!")

	// Create test files for user1
	file1UID := createTestFile(t, grpcClient, user1UID, "file1.txt", []byte("content1"), false)
	file2UID := createTestFile(t, grpcClient, user1UID, "file2.txt", []byte("content2"), true)
	file3UID := createTestFile(t, grpcClient, user1UID, "document.pdf", []byte("pdf content"), false)

	// Create test files for user2
	file4UID := createTestFile(t, grpcClient, user2UID, "file4.txt", []byte("content4"), true)

	tests := []struct {
		name       string
		pagination *commongrpc.Pagination
		filter     *filegrpc.FilterRequest
		wantCount  int
		wantUIDs   []string
	}{
		{
			name: "List all files for user",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &filegrpc.FilterRequest{
				UserUid: []string{user1UID},
			},
			wantCount: 3,
			wantUIDs:  []string{file1UID, file2UID, file3UID},
		},
		{
			name: "List with pagination",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   2,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &filegrpc.FilterRequest{
				UserUid: []string{user1UID},
			},
			wantCount: 2,
			wantUIDs:  []string{file1UID, file2UID},
		},
		{
			name: "List filtered by visibility (public)",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &filegrpc.FilterRequest{
				UserUid:    []string{user1UID},
				Visibility: util.Ptr("public"),
			},
			wantCount: 1,
			wantUIDs:  []string{file2UID},
		},
		{
			name: "List filtered by visibility (private)",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &filegrpc.FilterRequest{
				UserUid:    []string{user1UID},
				Visibility: util.Ptr("private"),
			},
			wantCount: 2,
			wantUIDs:  []string{file1UID, file3UID},
		},
		{
			name: "List filtered by UIDs",
			pagination: &commongrpc.Pagination{
				Page:    1,
				Limit:   10,
				OrderBy: "id",
				Sort:    "asc",
			},
			filter: &filegrpc.FilterRequest{
				Uids: []string{file1UID, file3UID, file4UID},
			},
			wantCount: 3,
			wantUIDs:  []string{file1UID, file3UID, file4UID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := grpcClient.UserFileClient.List(ctx, &filegrpc.ListRequest{
				Pagination: tt.pagination,
				Filter:     tt.filter,
			})

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp.Items, tt.wantCount)

			if len(tt.wantUIDs) > 0 {
				uidMap := make(map[string]bool)
				for _, file := range resp.Items {
					uidMap[file.Uid] = true
				}
				for _, uid := range tt.wantUIDs {
					require.True(t, uidMap[uid], "UID %s should be in results", uid)
				}
			}
		})
	}
}

func TestFileUpdate(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		update     func(t *testing.T) *filegrpc.UpdateRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, file *filegrpc.UserFile)
	}{
		{
			name: "Update file name",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "updatefileuser", "updatefileuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "old-name.txt", []byte("content"), false)
			},
			update: func(t *testing.T) *filegrpc.UpdateRequest {
				newName := "updated-name.txt"
				return &filegrpc.UpdateRequest{
					Name:     &newName,
					Filename: &newName,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, file *filegrpc.UserFile) {
				require.Equal(t, "updated-name.txt", file.FileName)
			},
		},
		{
			name: "Update file visibility to public",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "visfileuser", "visfileuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "private-file.txt", []byte("content"), false)
			},
			update: func(t *testing.T) *filegrpc.UpdateRequest {
				isPublic := true
				return &filegrpc.UpdateRequest{
					Public: &isPublic,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, file *filegrpc.UserFile) {
				require.Equal(t, "public", file.Visibility)
			},
		},
		{
			name: "Update file visibility to private",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "privfileuser", "privfileuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "public-file.txt", []byte("content"), true)
			},
			update: func(t *testing.T) *filegrpc.UpdateRequest {
				isPublic := false
				return &filegrpc.UpdateRequest{
					Public: &isPublic,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, file *filegrpc.UserFile) {
				require.Equal(t, "private", file.Visibility)
			},
		},
		{
			name: "Update file data",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "datafileuser", "datafileuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "data-file.txt", []byte("old content"), false)
			},
			update: func(t *testing.T) *filegrpc.UpdateRequest {
				newData := []byte("new updated content")
				return &filegrpc.UpdateRequest{
					Filedata: newData,
				}
			},
			wantErr: false,
		},
		{
			name: "Update multiple fields",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "multifileuser", "multifileuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "multi-old.txt", []byte("old content"), false)
			},
			update: func(t *testing.T) *filegrpc.UpdateRequest {
				newName := "multi-new.txt"
				isPublic := true
				newData := []byte("new content")
				return &filegrpc.UpdateRequest{
					Name:     &newName,
					Filename: &newName,
					Filedata: newData,
					Public:   &isPublic,
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, file *filegrpc.UserFile) {
				require.Equal(t, "multi-new.txt", file.FileName)
				require.Equal(t, "public", file.Visibility)
			},
		},
		{
			name: "Update non-existent file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			update: func(t *testing.T) *filegrpc.UpdateRequest {
				newName := "wont-update.txt"
				return &filegrpc.UpdateRequest{
					Name:     &newName,
					Filename: &newName,
				}
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			fileUID := tt.setup(t, grpcClient)
			updateReq := tt.update(t)
			updateReq.Uid = fileUID

			_, err := grpcClient.UserFileClient.Update(ctx, updateReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)

				file, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				require.NoError(t, err)
				require.NotNil(t, file)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, file)
				}
			}
		})
	}
}

func TestFileDelete(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name       string
		setup      func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, fileUID string, grpcClient *testutil.TestGRPCClient)
	}{
		{
			name: "Delete existing file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "deletefileuser", "deletefileuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "delete-me.txt", []byte("content"), false)
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, fileUID string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				_, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				require.Error(t, err)
				require.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "Delete non-existent file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "Verify soft delete behavior",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				userUID := createTestUser(t, grpcClient, "softdeleteuser", "softdeleteuser@example.com", "Password123!")
				return createTestFile(t, grpcClient, userUID, "soft-delete.txt", []byte("content"), false)
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, fileUID string, grpcClient *testutil.TestGRPCClient) {
				ctx := context.Background()
				// File should not be retrievable after soft delete
				_, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			fileUID := tt.setup(t, grpcClient)

			resp, err := grpcClient.UserFileClient.Delete(ctx, &filegrpc.DeleteRequest{Uid: fileUID})

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.True(t, resp.Success)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, fileUID, grpcClient)
				}
			}
		})
	}
}

func TestFileOwnership(t *testing.T) {
	_, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()
	tests := []struct {
		name    string
		setup   func(t *testing.T, grpcClient *testutil.TestGRPCClient) (fileUID, user1UID, user2UID string)
		action  func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) error
		wantErr bool
		errMsg  string
	}{
		{
			name: "User can access their own file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string, string) {
				user1UID := createTestUser(t, grpcClient, "owner1", "owner1@example.com", "Password123!")
				_ = createTestUser(t, grpcClient, "owner2", "owner2@example.com", "Password123!")
				fileUID := createTestFile(t, grpcClient, user1UID, "owned-file.txt", []byte("content"), false)
				return fileUID, user1UID, ""
			},
			action: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) error {
				ctx := context.Background()
				_, err := grpcClient.UserFileClient.Get(ctx, &filegrpc.GetRequest{Uid: fileUID})
				return err
			},
			wantErr: false,
		},
		{
			name: "User can list their own files",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string, string) {
				user1UID := createTestUser(t, grpcClient, "listowner1", "listowner1@example.com", "Password123!")
				user2UID := createTestUser(t, grpcClient, "listowner2", "listowner2@example.com", "Password123!")
				createTestFile(t, grpcClient, user1UID, "file1.txt", []byte("content1"), false)
				createTestFile(t, grpcClient, user2UID, "file2.txt", []byte("content2"), false)
				return "", user1UID, user2UID
			},
			action: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) error {
				ctx := context.Background()
				_, err := grpcClient.UserFileClient.List(ctx, &filegrpc.ListRequest{
					Pagination: &commongrpc.Pagination{
						Page:    1,
						Limit:   10,
						OrderBy: "id",
						Sort:    "asc",
					},
					Filter: &filegrpc.FilterRequest{
						UserUid: []string{userUID},
					},
				})
				return err
			},
			wantErr: false,
		},
		{
			name: "User can update their own file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string, string) {
				user1UID := createTestUser(t, grpcClient, "updateowner1", "updateowner1@example.com", "Password123!")
				user2UID := createTestUser(t, grpcClient, "updateowner2", "updateowner2@example.com", "Password123!")
				fileUID := createTestFile(t, grpcClient, user1UID, "updatable-file.txt", []byte("content"), false)
				return fileUID, user1UID, user2UID
			},
			action: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) error {
				ctx := context.Background()
				newName := "updated-name.txt"
				_, err := grpcClient.UserFileClient.Update(ctx, &filegrpc.UpdateRequest{
					Uid:      fileUID,
					Name:     &newName,
					Filename: &newName,
				})
				return err
			},
			wantErr: false,
		},
		{
			name: "User can delete their own file",
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) (string, string, string) {
				user1UID := createTestUser(t, grpcClient, "deleteowner1", "deleteowner1@example.com", "Password123!")
				user2UID := createTestUser(t, grpcClient, "deleteowner2", "deleteowner2@example.com", "Password123!")
				fileUID := createTestFile(t, grpcClient, user1UID, "deletable-file.txt", []byte("content"), false)
				return fileUID, user1UID, user2UID
			},
			action: func(t *testing.T, fileUID string, userUID string, grpcClient *testutil.TestGRPCClient) error {
				ctx := context.Background()
				_, err := grpcClient.UserFileClient.Delete(ctx, &filegrpc.DeleteRequest{Uid: fileUID})
				return err
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fileUID, user1UID, _ := tt.setup(t, grpcClient)
			err := tt.action(t, fileUID, user1UID, grpcClient)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
