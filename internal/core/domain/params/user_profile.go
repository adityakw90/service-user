package params

type UserProfileUpdateParam struct {
	FirstName  *string
	LastName   *string
	Bio        *string
	Avatar     []byte
	Attributes map[string]any
}
