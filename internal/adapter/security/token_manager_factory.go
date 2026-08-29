package security

import (
	"fmt"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// NewTokenManager creates a TokenManager based on the strategy string.
// strategy: "stateless" | "whitelist" | "blacklist"
// store is only used for whitelist and blacklist strategies (can be nil for stateless).
func NewTokenManager(strategy string, gen portsec.TokenGenerator, store portsec.TokenStore) (portsec.TokenManager, error) {
	switch strategy {
	case "stateless", "":
		return NewStatelessTokenManager(gen), nil
	case "whitelist":
		return NewWhitelistTokenManager(gen, store), nil
	case "blacklist":
		return NewBlacklistTokenManager(gen, store), nil
	default:
		return nil, fmt.Errorf("unknown token manager strategy: %s", strategy)
	}
}
