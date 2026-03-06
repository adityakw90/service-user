package param

// InvalidateOptions holds the options for Invalidate operations.
type InvalidateOptions struct {
	UIDs []string
	IDs  []int64
}

// InvalidateOpt is a functional option for Invalidate operations.
type InvalidateOpt func(*InvalidateOptions)

// WithUIDs specifies the UIDs to invalidate.
func WithUIDs(uids ...string) InvalidateOpt {
	return func(o *InvalidateOptions) {
		o.UIDs = uids
	}
}

// WithIDs specifies the IDs to invalidate.
func WithIDs(ids ...int64) InvalidateOpt {
	return func(o *InvalidateOptions) {
		o.IDs = ids
	}
}
