package token

import (
	"time"
)

type Maker interface {
	VerifyToken(token string) (*PayloadSimp, error)
	VerifyTokenUser(token string) (*PayloadUser, error)
	VerifyPublicToken(token string) (*Payload, error)
	CreateToken(
		tokenHistID int64,
		ip string,
		numberRequests int64,
		valid bool,
		expiredAt time.Time,
	) (string, error)
	CreateTokenUser(
		id int64,
		name string,
		email string,
		profileId int64,
		document string,
		googleId string,
		expireAt time.Time,
	) (string, error)
}
