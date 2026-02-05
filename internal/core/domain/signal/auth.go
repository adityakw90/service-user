package signal

import "github.com/adityakw90/service-user/internal/core/domain/model"

type AuthSignal struct {
	UID      string
	Username string
	Email    string
	Status   model.UserStatus
	Active   bool
	Deleted  bool
}
