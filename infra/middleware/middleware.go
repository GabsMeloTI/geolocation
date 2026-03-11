package middleware

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"strings"
	"time"
)

func CheckAuthorization(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		bearerToken := c.Request().Header.Get("Authorization")
		tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

		maker, err := token.NewPasetoMaker(os.Getenv("SIGNATURE_STRING_SIMP"))
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}

		tokenPayload, err := maker.VerifyToken(tokenStr)
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}
		c.Set("token_id", tokenPayload.ID)
		c.Set("token_user_id", tokenPayload.UserID)
		c.Set("token_user_nickname", tokenPayload.UserNickname)
		c.Set("token_access_key", tokenPayload.AccessKey)
		c.Set("token_access_ID", tokenPayload.AccessID)
		c.Set("token_tenant_id", tokenPayload.TenantID)
		c.Set("token_expiry_at", tokenPayload.ExpiredAt)
		c.Set("token_user_org_id", tokenPayload.UserOrgId)
		c.Set("token_user_email", tokenPayload.UserEmail)
		c.Set("token_document", tokenPayload.Document)
		c.Set("token_user_name", tokenPayload.UserName)

		return handlerFunc(c)
	}
}

func CheckPublicAuthorization(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		bearerToken := c.Request().Header.Get("Authorization")
		tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

		maker, err := token.NewPasetoMaker(os.Getenv("SIGNATURE_STRING"))
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}

		tokenPayload, err := maker.VerifyPublicToken(tokenStr)
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}
		c.Set("token_id_hist", tokenPayload.ID)
		c.Set("token_ip", tokenPayload.IP)
		c.Set("token_number_requests", tokenPayload.NumberRequests)
		c.Set("token_valid", tokenPayload.Valid)
		c.Set("token_expired_at", tokenPayload.ExpiredAt)

		return handlerFunc(c)
	}
}

func CheckUserAuthorization(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		bearerToken := c.Request().Header.Get("Authorization")
		tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

		maker, err := token.NewPasetoMaker(os.Getenv("SIGNATURE_STRING"))
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}

		tokenPayload, err := maker.VerifyTokenUser(tokenStr)

		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}

		c.Set("token_id", tokenPayload.ID)
		c.Set("token_user_name", tokenPayload.Name)
		c.Set("token_user_email", tokenPayload.Email)
		c.Set("token_expire_at", tokenPayload.ExpireAt)
		c.Set("token_document", tokenPayload.Document)
		c.Set("token_google_id", tokenPayload.GoogleID)
		c.Set("token_profile_id", tokenPayload.ProfileID)

		return handlerFunc(c)
	}
}

func CheckUserWsAuthorization(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenStr := c.QueryParam("token")

		maker, err := token.NewPasetoMaker(os.Getenv("SIGNATURE_STRING"))

		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}

		tokenPayload, err := maker.VerifyTokenUser(tokenStr)

		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}

		c.Set("token_id", tokenPayload.ID)
		c.Set("token_user_name", tokenPayload.Name)
		c.Set("token_user_email", tokenPayload.Email)
		c.Set("token_expire_at", tokenPayload.ExpireAt)
		c.Set("token_document", tokenPayload.Document)
		c.Set("token_google_id", tokenPayload.GoogleID)
		c.Set("token_profile_id", tokenPayload.ProfileID)

		return handlerFunc(c)
	}
}

type RequestLoggerRepo interface {
	CreateUserRequestHist(ctx context.Context, arg db.CreateUserRequestHistParams) (db.UserRequestHist, error)
}

func RequestLogger(repo RequestLoggerRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			
			// Execute down the chain
			err := next(c)
			
			duration := time.Since(start).Milliseconds()

			bearerToken := c.Request().Header.Get("Authorization")
			tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

			var userID sql.NullInt64

			// token_id is set by CheckUserAuthorization (which is User ID)
			if uid := c.Get("token_id"); uid != nil {
				if id, ok := uid.(int64); ok {
					userID = sql.NullInt64{Int64: id, Valid: true}
				}
			}

			// Fallback for public tokens
			if !userID.Valid {
				if uidHist := c.Get("token_id_hist"); uidHist != nil {
					if id, ok := uidHist.(int64); ok {
						userID = sql.NullInt64{Int64: id, Valid: true}
					}
				}
			}

			if tokenStr != "" {
				// Fire and forget audit log to not block caller further
				go func(ctx context.Context, token string, path string, method string, status int, timeMs int64, uid sql.NullInt64) {
					_, _ = repo.CreateUserRequestHist(ctx, db.CreateUserRequestHistParams{
						UserID:          uid,
						Token:           token,
						Endpoint:        path,
						Method:          method,
						StatusCode:      int32(status),
						ExecutionTimeMs: int32(timeMs),
					})
				}(context.Background(), tokenStr, c.Request().URL.Path, c.Request().Method, c.Response().Status, duration, userID)
			}

			return err
		}
	}
}
