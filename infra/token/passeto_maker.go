package token

import (
	"encoding/json"
	"fmt"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
	"time"
)

type PasetoMaker struct {
	paseto      *paseto.V2
	symetricKey []byte
}

func (maker *PasetoMaker) CreateTokenUser(id int64, name string, email string, profileId int64, document string, googleId string, expireAt time.Time) (string, error) {
	payload, err := NewPayloadUser(id, name, email, profileId, document, googleId, expireAt)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symetricKey, payload, nil)
}

func NewPasetoMaker(symetricKey string) (Maker, error) {
	if len(symetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characacteres", chacha20poly1305.KeySize)
	}

	return &PasetoMaker{
		paseto:      paseto.NewV2(),
		symetricKey: []byte(symetricKey),
	}, nil

}

func (maker *PasetoMaker) VerifyToken(token string) (*PayloadSimp, error) {
	payload := &PayloadSimp{}

	err := maker.paseto.Decrypt(token, maker.symetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (maker *PasetoMaker) VerifyTokenUser(token string) (*PayloadUser, error) {
	payload := &PayloadUser{}

	err := maker.paseto.Decrypt(token, maker.symetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (maker *PasetoMaker) VerifyPublicToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.validPublic()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (maker *PasetoMaker) CreateToken(tokenHistID int64, ip string, numberRequests int64, valid bool, expiredAt time.Time) (string, error) {
	payload, err := NewPayload(tokenHistID, ip, numberRequests, valid, expiredAt)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symetricKey, payload, nil)
}

func (maker *PasetoMaker) CreateTokenUserID(userID int64) (string, error) {
	payload, err := NewPayloadUserID(userID)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return maker.paseto.Encrypt(maker.symetricKey, payloadJSON, nil)
}
