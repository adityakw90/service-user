package request

import (
	"testing"

	commonpb "github.com/adityakw90/service-user-proto/gen/go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationRequest(t *testing.T) {
	tests := []struct {
		name     string
		pbReq    *commonpb.Pagination
		expected *PaginationRequest
	}{
		{
			name: "full pagination",
			pbReq: &commonpb.Pagination{
				Page:    2,
				Limit:   20,
				Sort:    " desc ",
				OrderBy: " username ",
			},
			expected: &PaginationRequest{
				Page:    2,
				Limit:   20,
				Sort:    "desc",
				OrderBy: "username",
			},
		},
		{
			name: "minimal pagination with defaults",
			pbReq: &commonpb.Pagination{
				Page:  1,
				Limit: 10,
			},
			expected: &PaginationRequest{
				Page:    1,
				Limit:   10,
				Sort:    "",
				OrderBy: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PaginationRequestFromPb(tt.pbReq)
			assert.Equal(t, tt.expected, got)

			params := got.ToPaginationParams()
			require.NotNil(t, params)
			assert.Equal(t, tt.expected.Page, *params.Page)
			assert.Equal(t, tt.expected.Limit, *params.Limit)
			assert.Equal(t, tt.expected.Sort, *params.Sort)
			assert.Equal(t, tt.expected.OrderBy, *params.OrderBy)
		})
	}
}
