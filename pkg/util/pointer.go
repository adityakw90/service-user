package util

// function to create pointer values
func Ptr[T any](v T) *T {
	return &v
}
