package model

import (
	"testing"
)

/* test completed by jojo */

func TestCore_Domain_Meta(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func() Meta
		checkFn func(Meta)
	}{
		{
			name: "Meta is initialized correctly",
			setupFn: func() Meta {
				return Meta{
					Page:  1,
					Limit: 10,
					Total: 100,
					Pages: 10,
				}
			},
			checkFn: func(meta Meta) {
				if meta.Page != 1 {
					t.Errorf("Page = %d, want 1", meta.Page)
				}
				if meta.Limit != 10 {
					t.Errorf("Limit = %d, want 10", meta.Limit)
				}
				if meta.Total != 100 {
					t.Errorf("Total = %d, want 100", meta.Total)
				}
				if meta.Pages != 10 {
					t.Errorf("Pages = %d, want 10", meta.Pages)
				}
			},
		},
		{
			name: "Meta is initialized with zero values",
			setupFn: func() Meta {
				return Meta{}
			},
			checkFn: func(meta Meta) {
				if meta.Page != 0 {
					t.Errorf("Page = %d, want 0", meta.Page)
				}
				if meta.Limit != 0 {
					t.Errorf("Limit = %d, want 0", meta.Limit)
				}
				if meta.Total != 0 {
					t.Errorf("Total = %d, want 0", meta.Total)
				}
				if meta.Pages != 0 {
					t.Errorf("Pages = %d, want 0", meta.Pages)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := tt.setupFn()
			tt.checkFn(meta)
		})
	}
}
