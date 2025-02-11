package token

import (
	"time"
)

type Maker interface {
	VerifyToken(token string) (*PayloadSimp, error)
	VerifyPublicToken(token string) (*Payload, error)
	CreateToken(
		tokenHistID int64,
		ip string,
		numberRequests int64,
		valid bool,
		expiredAt time.Time,
	) (string, error)
}
