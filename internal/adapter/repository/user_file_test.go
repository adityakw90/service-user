package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFileRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(mock pgxmock.PgxPoolIface, id int64)
		wantErr   bool
	}{
		{
			name: "Get existing file by ID",
			id:   1,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(id, "file-uid-001", int64(1), "avatar", "profile.jpg", "/path/to/file", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE id = \$1`).
					WithArgs(id).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Get non-existent file by ID",
			id:   999999999,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"})
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE id = \$1`).
					WithArgs(id).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.id)
			}

			result, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.id, result.ID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserFileRepository_GetByUID(t *testing.T) {
	tests := []struct {
		name      string
		uid       string
		setupMock func(mock pgxmock.PgxPoolIface, uid string)
		wantErr   bool
	}{
		{
			name: "Get existing file by UID",
			uid:  "file-uid-123",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), uid, int64(1), "avatar", "profile.jpg", "/path/to/file", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE uid = \$1`).
					WithArgs(uid).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Get non-existent file by UID",
			uid:  "non-existent-uid",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"})
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE uid = \$1`).
					WithArgs(uid).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.uid)
			}

			result, err := repo.GetByUID(context.Background(), tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.uid, result.UID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserFileRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		file    *model.UserFile
		wantErr bool
	}{
		{
			name: "Create valid file",
			file: &model.UserFile{
				UID:        "file-uid-001",
				UserID:     1,
				FileType:   "avatar",
				FileName:   "profile.jpg",
				FilePath:   "/path/to/file",
				MimeType:   "image/jpeg",
				SizeBytes:  1024,
				Visibility: "public",
				CreatedAt:  time.Now().UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), nil)

			rows := pgxmock.NewRows([]string{"id"}).AddRow(int64(1))
			mockPool.ExpectQuery(`INSERT INTO user_file`).
				WithArgs(
					tt.file.UID, tt.file.UserID, tt.file.FileType, tt.file.FileName,
					tt.file.FilePath, tt.file.MimeType, tt.file.SizeBytes, tt.file.Visibility, pgxmock.AnyArg(),
				).
				WillReturnRows(rows)

			created, err := repo.Create(context.Background(), tt.file)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, tt.file.UID, created.UID)
			assert.Equal(t, tt.file.FileName, created.FileName)
			assert.NotZero(t, created.ID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserFileRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		file      *model.UserFile
		setupMock func(mock pgxmock.PgxPoolIface, file *model.UserFile)
		wantErr   bool
	}{
		{
			name: "Update file",
			file: &model.UserFile{
				ID:         1,
				FileType:   "document",
				FileName:   "newfile.pdf",
				FilePath:   "/new/path",
				MimeType:   "application/pdf",
				SizeBytes:  2048,
				Visibility: "private",
			},
			setupMock: func(mock pgxmock.PgxPoolIface, file *model.UserFile) {
				mock.ExpectExec(`UPDATE user_file SET file_type = \$1, file_name = \$2, file_path = \$3, mime_type = \$4, size_bytes = \$5, visibility = \$6 WHERE id = \$7`).
					WithArgs(
						file.FileType, file.FileName, file.FilePath,
						file.MimeType, file.SizeBytes, file.Visibility, file.ID,
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.file)
			}

			err = repo.Update(context.Background(), tt.file)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserFileRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		file      *model.UserFile
		setupMock func(mock pgxmock.PgxPoolIface, file *model.UserFile)
		wantErr   bool
	}{
		{
			name: "Delete existing file",
			file: &model.UserFile{
				ID: 1,
			},
			setupMock: func(mock pgxmock.PgxPoolIface, file *model.UserFile) {
				mock.ExpectExec(`DELETE FROM user_file WHERE id = \$1`).
					WithArgs(file.ID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.file)
			}

			err = repo.Delete(context.Background(), tt.file)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserFileRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.UserFileListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List all files with pagination",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now()).
					AddRow(int64(2), "uid2", int64(1), "document", "file2.pdf", "/path2", "application/pdf", int64(2048), "private", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "List files with filter by file type",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     &param.UserFileListFilterParam{FileType: util.Ptr("avatar")},
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file WHERE file_type = \$1`).
					WithArgs("avatar").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE file_type = \$1`).
					WithArgs("avatar", pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - uid",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "uid"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - user_id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "user_id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - file_type",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "file_type"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - file_name",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "file_name"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - created_at",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "created_at"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id; DROP TABLE user_file; --"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "nonexistent"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Nil OrderBy - should use default",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: nil},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_file`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}).
					AddRow(int64(1), "uid1", int64(1), "avatar", "file1.jpg", "/path1", "image/jpeg", int64(1024), "public", time.Now())
				mock.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.pagination, tt.filter)
			}

			result, err := repo.List(context.Background(), tt.pagination, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(result.Items))

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserFileRepository_GetByID_Logging(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	logger := &mockLogger{}
	repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), logger)

	// Test scenario 1: ErrFileNotFound -> Should not log
	mockPool.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}))

	_, err = repo.GetByID(context.Background(), 1)
	assert.ErrorIs(t, err, errors.ErrFileNotFound)
	assert.Empty(t, logger.LoggedErrors)

	// Test scenario 2: Unexpected DB error -> Should log
	mockPool.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("database query failure"))

	_, err = repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Len(t, logger.LoggedErrors, 1)
	assert.Equal(t, "failed to get user file by id", logger.LoggedErrors[0].Msg)
	assert.NotContains(t, logger.LoggedErrors[0].Fields, "id")
	assert.NotNil(t, logger.LoggedErrors[0].Fields["error"])

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserFileRepository_GetByUID_Logging(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	logger := &mockLogger{}
	repo := NewUserFileRepository(mockPool, infra.NewNoopTracer(), logger)

	// Test scenario 1: ErrFileNotFound -> Should not log
	mockPool.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE uid = \$1`).
		WithArgs("uid-1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "uid", "user_id", "file_type", "file_name", "file_path", "mime_type", "size_bytes", "visibility", "created_at"}))

	_, err = repo.GetByUID(context.Background(), "uid-1")
	assert.ErrorIs(t, err, errors.ErrFileNotFound)
	assert.Empty(t, logger.LoggedErrors)

	// Test scenario 2: Unexpected DB error -> Should log
	mockPool.ExpectQuery(`SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at FROM user_file WHERE uid = \$1`).
		WithArgs("uid-1").
		WillReturnError(fmt.Errorf("database query failure"))

	_, err = repo.GetByUID(context.Background(), "uid-1")
	assert.Error(t, err)
	assert.Len(t, logger.LoggedErrors, 1)
	assert.Equal(t, "failed to get user file by uid", logger.LoggedErrors[0].Msg)
	assert.NotContains(t, logger.LoggedErrors[0].Fields, "uid")
	assert.NotNil(t, logger.LoggedErrors[0].Fields["error"])

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
