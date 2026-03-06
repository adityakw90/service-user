package params

import "testing"

func TestInvalidateOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []InvalidateOpt
		wantUIDs []string
		wantIDs  []int64
	}{
		{
			name:     "WithUIDs only",
			opts:     []InvalidateOpt{WithUIDs("uid1", "uid2")},
			wantUIDs: []string{"uid1", "uid2"},
			wantIDs:  nil,
		},
		{
			name:     "WithIDs only",
			opts:     []InvalidateOpt{WithIDs(1, 2, 3)},
			wantUIDs: nil,
			wantIDs:  []int64{1, 2, 3},
		},
		{
			name:     "WithUIDs and WithIDs",
			opts:     []InvalidateOpt{WithUIDs("uid1"), WithIDs(1)},
			wantUIDs: []string{"uid1"},
			wantIDs:  []int64{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &InvalidateOptions{}
			for _, opt := range tt.opts {
				opt(options)
			}

			if !slicesEqual(options.UIDs, tt.wantUIDs) {
				t.Errorf("UIDs = %v, want %v", options.UIDs, tt.wantUIDs)
			}
			if !slicesEqual(options.IDs, tt.wantIDs) {
				t.Errorf("IDs = %v, want %v", options.IDs, tt.wantIDs)
			}
		})
	}
}

func slicesEqual[S ~[]E, E comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
