package errors

var (
	ErrEventTypeRequired  = NewCustomError(90001, "event type is required", nil)
	ErrEntityTypeRequired = NewCustomError(90002, "event entity type is required", nil)
	ErrEntityIDRequired   = NewCustomError(90003, "event entity id is required", nil)
	ErrEntityInvalidType  = NewCustomError(90004, "event entity type is invalid", nil)
)
