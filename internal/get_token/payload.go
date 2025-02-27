package get_token

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"time"
)

func GetPayloadToken(c echo.Context) PayloadDTO {
	strID, _ := c.Get("token_id").(uuid.UUID)
	strUserID, _ := c.Get("token_user_id").(string)
	strUserNickname, _ := c.Get("token_user_nickname").(string)
	strExpiryAt, _ := c.Get("token_expiry_at").(time.Time)
	strAccessKey, _ := c.Get("token_access_key").(int64)
	strAccessID, _ := c.Get("token_access_ID").(int64)
	strTenantID, _ := c.Get("token_tenant_id").(string)
	strUserOrgId, _ := c.Get("token_user_org_id").(int64)
	strUserEmail, _ := c.Get("token_user_email").(string)
	strDocument, _ := c.Get("token_document").(string)
	strUserName, _ := c.Get("token_user_name").(string)

	tenantID, _ := uuid.Parse(strTenantID)

	return PayloadDTO{
		ID:           strID,
		UserID:       strUserID,
		UserNickname: strUserNickname,
		ExpiryAt:     strExpiryAt,
		AccessKey:    strAccessKey,
		AccessID:     strAccessID,
		TenantID:     tenantID,
		UserOrgId:    strUserOrgId,
		UserEmail:    strUserEmail,
		Document:     strDocument,
		UserName:     strUserName,
	}
}

func GetPublicPayloadToken(c echo.Context) PublicPayloadDTO {
	strID, _ := c.Get("token_id_hist").(int64)
	strIP, _ := c.Get("token_ip").(string)
	strNumberRequests, _ := c.Get("token_number_requests").(int64)
	strValid, _ := c.Get("token_valid").(bool)
	strExpiredAt, _ := c.Get("token_expired_at").(time.Time)

	return PublicPayloadDTO{
		ID:             strID,
		IP:             strIP,
		NumberRequests: strNumberRequests,
		Valid:          strValid,
		ExpiredAt:      strExpiredAt,
	}
}

func GetUserPayloadToken(c echo.Context) PayloadUserDTO {
	strID, _ := c.Get("token_id").(int64)
	strName, _ := c.Get("token_user_name").(string)
	strEmail, _ := c.Get("token_user_email").(string)
	strProfileID, _ := c.Get("token_profile_id").(int64)
	strDocument, _ := c.Get("token_document").(string)
	strGoogleID, _ := c.Get("token_google_id").(string)
	strExpireAt, _ := c.Get("token_expire_at").(time.Time)

	return PayloadUserDTO{
		ID:        strID,
		Name:      strName,
		Email:     strEmail,
		ProfileID: strProfileID,
		Document:  strDocument,
		GoogleID:  strGoogleID,
		ExpireAt:  strExpireAt,
	}
}
