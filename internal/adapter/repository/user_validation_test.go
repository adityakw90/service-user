package repository

import (
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_validateOrderBy(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		want       string
		wantErr    bool
	}{
		{
			name: "Valid OrderBy - username",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "username"; return &s }(),
			},
			want:    "username",
			wantErr: false,
		},
		{
			name: "Valid OrderBy - email",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "email"; return &s }(),
			},
			want:    "email",
			wantErr: false,
		},
		{
			name: "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id; DROP TABLE users; --"; return &s }(),
			},
			wantErr: true,
		},
		{
			name: "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "nonexistent"; return &s }(),
			},
			wantErr: true,
		},
		{
			name:       "Nil pagination",
			pagination: nil,
			want:       "created_at",
			wantErr:    false,
		},
		{
			name: "Nil OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: nil,
			},
			want:    "created_at",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateOrderBy(tt.pagination, "created_at", allowedOrderByUser)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
