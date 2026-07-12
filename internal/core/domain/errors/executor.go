package errors

const (
	// Error Code
	ErrCodeExecutorClosed    = 13001
	ErrCodeExecutorFnInvalid = 13002

	// Error Message
	ErrMessageExecutorClosed    = "executor closed"
	ErrMessageExecutorFnInvalid = "executor function must not be nil"
)

var (
	ErrExecutorClosed    = NewCustomError(ErrCodeExecutorClosed, ErrMessageExecutorClosed, nil)
	ErrExecutorFnInvalid = NewCustomError(ErrCodeExecutorFnInvalid, ErrMessageExecutorFnInvalid, nil)
)
