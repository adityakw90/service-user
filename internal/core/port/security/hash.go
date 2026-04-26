package security

// Hasher is a port for password/pin hashing.
type Hasher interface {
	Hash(plain string) (string, error)
	Compare(hashed, plain string) bool
}
