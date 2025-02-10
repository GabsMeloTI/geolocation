package middleware

import (
	"geolocation/infra/token"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"strings"
)

func CheckAuthorization(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		bearerToken := c.Request().Header.Get("Authorization")
		tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

		maker, err := token.NewPasetoMaker(os.Getenv("TOKEN_SIGNATURE"))
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
