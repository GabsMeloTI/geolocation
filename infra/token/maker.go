package token

import (
	"time"
)

type Maker interface {
	VerifyToken(token string) (*PayloadSimp, error)
	CreateToken(
		tokenHistID int64,
		ip string,
		numberRequests int64,
		valid bool,
		expiredAt time.Duration,
	) (string, error)
}
