package repository

import (
	stderrors "errors"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestValidateOrderBy(t *testing.T) {
	// Create a test allowed order by map
	testAllowedOrderBy := map[string]param.UserOrderBy{
		"id":         param.OrderByUserID,
		"uid":        param.OrderByUserUID,
		"username":   param.OrderByUserUsername,
		"email":      param.OrderByUserEmail,
		"status":     param.OrderByUserStatus,
		"created_at": param.OrderByUserCreatedAt,
		"updated_at": param.OrderByUserUpdatedAt,
	}

	tests := []struct {
		name       string
		pagination *param.PaginationParam
		defaultOrd string
		allowed    map[string]param.UserOrderBy
		want       string
	}{
		{
			name: "Valid OrderBy - username",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "username"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "username",
		},
		{
			name: "Valid OrderBy - email",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "email"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "email",
		},
		{
			name: "Valid OrderBy - id",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "id",
		},
		{
			name: "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id; DROP TABLE users; --"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at", // fallback to default
		},
		{
			name: "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "nonexistent"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at", // fallback to default
		},
		{
			name:       "Nil pagination",
			pagination: nil,
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at",
		},
		{
			name: "Nil OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: nil,
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at",
		},
		{
			name: "Empty OrderBy string",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := ""; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at", // empty string is not in map
		},
		{
			name: "Valid OrderBy with different default",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "status"; return &s }(),
			},
			defaultOrd: "id",
			allowed:    testAllowedOrderBy,
			want:       "status",
		},
		{
			name: "Invalid OrderBy with custom default",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "invalid"; return &s }(),
			},
			defaultOrd: "updated_at",
			allowed:    testAllowedOrderBy,
			want:       "updated_at",
		},
		{
			name: "Empty allowed map",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    map[string]param.UserOrderBy{},
			want:       "created_at", // fallback to default since map is empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateOrderBy(tt.pagination, tt.defaultOrd, tt.allowed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateOrderByWithDevice(t *testing.T) {
	// Test with DeviceOrderBy to ensure generic works with different types
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		defaultOrd string
		want       string
	}{
		{
			name: "Valid Device OrderBy - device_name",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "device_name"; return &s }(),
			},
			defaultOrd: "created_at",
			want:       "device_name",
		},
		{
			name: "Valid Device OrderBy - device_fingerprint",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "device_fingerprint"; return &s }(),
			},
			defaultOrd: "created_at",
			want:       "device_fingerprint",
		},
		{
			name: "Invalid Device OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "invalid_column"; return &s }(),
			},
			defaultOrd: "created_at",
			want:       "created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateOrderBy(tt.pagination, tt.defaultOrd, allowedOrderByDevice)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateOrderByWithUserFile(t *testing.T) {
	// Test with UserFileOrderBy to ensure generic works with different types
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		defaultOrd string
		want       string
	}{
		{
			name: "Valid UserFile OrderBy - file_type",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "file_type"; return &s }(),
			},
			defaultOrd: "created_at",
			want:       "file_type",
		},
		{
			name: "Valid UserFile OrderBy - file_name",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "file_name"; return &s }(),
			},
			defaultOrd: "created_at",
			want:       "file_name",
		},
		{
			name: "Invalid UserFile OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "invalid_column"; return &s }(),
			},
			defaultOrd: "created_at",
			want:       "created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateOrderBy(tt.pagination, tt.defaultOrd, allowedOrderByUserFile)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandlePgError(t *testing.T) {
	tests := []struct {
		name    string
		input   error
		wantErr error
	}{
		{
			name: "email unique constraint violation",
			input: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "idx_user_email_active",
			},
			wantErr: errors.ErrDuplicateEmail,
		},
		{
			name: "username unique constraint violation",
			input: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "idx_user_username_active",
			},
			wantErr: errors.ErrDuplicateUsername,
		},
		{
			name: "unknown unique constraint",
			input: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "unknown_constraint",
			},
			wantErr: errors.ErrResourceConflict,
		},
		{
			name:    "non-pg error",
			input:   stderrors.New("some other error"),
			wantErr: nil,
		},
		{
			name:    "nil error",
			input:   nil,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := HandlePgError(tt.input)
			if tt.wantErr == nil {
				if gotErr != nil {
					t.Errorf("HandlePgError() = %v, want nil", gotErr)
				}
				return
			}
			if gotErr == nil {
				t.Errorf("HandlePgError() = nil, want %v", tt.wantErr)
				return
			}
			var gotCustomErr *errors.CustomError
			var wantCustomErr *errors.CustomError
			if !stderrors.As(gotErr, &gotCustomErr) || !stderrors.As(tt.wantErr, &wantCustomErr) {
				t.Errorf("HandlePgError() error type mismatch")
				return
			}
			if gotCustomErr.Code != wantCustomErr.Code {
				t.Errorf("HandlePgError() code = %d, want %d", gotCustomErr.Code, wantCustomErr.Code)
			}
		})
	}
}
