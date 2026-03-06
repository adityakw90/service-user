package repository

import (
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/param"
)

func TestUserRepository_validateOrderBy(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		want       string
	}{
		{
			name: "Valid OrderBy - username",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "username"; return &s }(),
			},
			want: "username",
		},
		{
			name: "Valid OrderBy - email",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "email"; return &s }(),
			},
			want: "email",
		},
		{
			name: "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id; DROP TABLE users; --"; return &s }(),
			},
			want: "created_at", // fallback to default
		},
		{
			name: "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "nonexistent"; return &s }(),
			},
			want: "created_at", // fallback to default
		},
		{
			name:       "Nil pagination",
			pagination: nil,
			want:       "created_at",
		},
		{
			name: "Nil OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: nil,
			},
			want: "created_at",
		},
	}

	repo := &UserRepository{} // minimal setup for validation test

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.validateOrderBy(tt.pagination, "created_at")
			if got != tt.want {
				t.Errorf("validateOrderBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
