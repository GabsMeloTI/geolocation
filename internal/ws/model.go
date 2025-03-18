package ws

import (
	"database/sql"
	"time"

	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
	routes "geolocation/internal/new_routes"
)

type CreateChatRoomRequest struct {
	AdvertisementID int64 `json:"advertisement_id"`
}

type CreateChatRoomResponse struct {
	ID                  int64     `json:"id"`
	AdvertisementID     int64     `json:"advertisement_id"`
	AdvertisementUserID int64     `json:"advertisement_user_id"`
	InterestedUserID    int64     `json:"interested_user_id"`
	CreatedAt           time.Time `json:"created_at"`
}

type HomeResponse struct {
	Advertisement  []RoomResponse  `json:"advertisements"`
	Interested     []RoomResponse  `json:"interested"`
	ActiveFreights []ActiveFreight `json:"active_freights"`
	TotalCount     int64           `json:"total_count"`
}

type ActiveFreight struct {
	AdvertisementId         int64   `json:"advertisement_id"`
	Latitude                float64 `json:"latitude"`
	Longitude               float64 `json:"longitude"`
	DurationText            string  `json:"duration"`
	DistanceText            string  `json:"distance"`
	DriverName              string  `json:"driver_name"`
	TractorUnitLicensePlate string  `json:"tractor_unit_license_plate"`
	TrailerLicensePlate     string  `json:"trailer_license_p_late"`
}

type RoomResponse struct {
	RoomId              int64            `json:"room_id"`
	Title               string           `json:"title"`
	Origin              string           `json:"origin"`
	Destination         string           `json:"destination"`
	AdvertisementUserId int64            `json:"advertisement_user_id"`
	AdvertisementId     int64            `json:"advertisement_id"`
	Distance            int64            `json:"distance"`
	LastMessage         *MessageResponse `json:"last_message,omitempty"`
	UnreadCount         int64            `json:"unread_count"`
}

type MessageResponse struct {
	MessageId      int64     `json:"message_id"`
	RoomId         int64     `json:"room_id"`
	UserId         int64     `json:"user_id"`
	Content        string    `json:"content"`
	Name           string    `json:"name"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
	TypeMessage    string    `json:"type_message,omitempty"`
}

func (r CreateChatRoomRequest) ParseToCreateChatRoomResponse(
	room db.ChatRoom,
) CreateChatRoomResponse {
	return CreateChatRoomResponse{
		ID:                  room.ID,
		AdvertisementID:     room.AdvertisementID,
		AdvertisementUserID: room.AdvertisementUserID,
		InterestedUserID:    room.InterestedUserID,
		CreatedAt:           room.CreatedAt,
	}
}

type UpdateOfferRequest struct {
	MessageId       int64   `json:"message_id"`
	IsAccepted      bool    `json:"is_accepted"`
	Price           float64 `json:"price"`
	AdvertisementId int64   `json:"advertisement_id"`
	TractorUnitId   int64   `json:"tractor_unit_id"`
	TrailerId       int64   `json:"trailer_id"`
	DriverId        int64   `json:"driver_id"`
}

type UpdateOfferDTO struct {
	Request UpdateOfferRequest
	Payload get_token.PayloadUserDTO
}

func (u UpdateOfferDTO) ToCreateOfferParams(interestedId int64) db.CreateOfferParams {
	return db.CreateOfferParams{
		AdvertisementID: sql.NullInt64{
			Int64: u.Request.AdvertisementId,
			Valid: true,
		},
		Price: u.Request.Price,
		InterestedID: sql.NullInt64{
			Int64: interestedId,
			Valid: true,
		},
		Status: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
	}
}

func (u UpdateOfferDTO) ToUpdateAdvertisementSituationParams() db.UpdateAdvertisementSituationParams {
	return db.UpdateAdvertisementSituationParams{
		Situation: "inactive",
		UpdatedWho: sql.NullString{
			String: "system",
			Valid:  true,
		},
		ID: u.Request.AdvertisementId,
	}
}

func (u UpdateOfferDTO) ToCreateTruckParams() db.CreateTruckParams {
	return db.CreateTruckParams{
		TractorUnitID: u.Request.TractorUnitId,
		TrailerID: sql.NullInt64{
			Int64: u.Request.TrailerId,
			Valid: true,
		},
		DriverID: u.Request.DriverId,
	}
}

func (u UpdateOfferDTO) ToCreateAppointmentParams(
	advertisementUserId, interestedUserId, offerId, truckId int64,
) db.CreateAppointmentParams {
	return db.CreateAppointmentParams{
		AdvertisementUserID: advertisementUserId,
		InterestedUserID:    interestedUserId,
		OfferID:             offerId,
		TruckID:             truckId,
		AdvertisementID:     u.Request.AdvertisementId,
		CreatedWho:          "system",
	}
}

type FreightLocationDetailsResponse struct {
	DurationText            string `json:"duration"`
	DistanceText            string `json:"distance_text"`
	DriverName              string `json:"driver_name"`
	AdvertisementUserId     int64  `json:"advertisement_user_id"`
	TractorUnitLicensePlate string `json:"tractor_unit_license_plate"`
	TrailerLicensePlate     string `json:"trailer_license_p_late"`
}

type UpdateFreightData struct {
	AdvertisementId int64   `json:"advertisement_id"`
	OriginLatitude  float64 `json:"latitude"`
	OriginLongitude float64 `json:"longitude"`
}

func (u UpdateFreightData) ToCreateActiveFreightParams(
	route routes.SimpleRouteResponse,
	freightDetails db.GetAppointmentDetailsByAdvertisementIdRow,
) db.CreateActiveFreightParams {
	return db.CreateActiveFreightParams{
		AdvertisementID:     u.AdvertisementId,
		AdvertisementUserID: freightDetails.AdvertisementUserID.Int64,
		Latitude:            u.OriginLatitude,
		Longitude:           u.OriginLongitude,
		Duration:            route.Summary.SimpleRoute.Duration.Text,
		Distance:            route.Summary.SimpleRoute.Distance.Text,
		DriverName:          freightDetails.Name,
		TractorUnitLicensePlate: sql.NullString{
			String: freightDetails.TractorUnitLicensePlate.String,
			Valid:  true,
		},
		TrailerLicensePlate: sql.NullString{
			String: freightDetails.TrailerLicensePlate.String,
			Valid:  true,
		},
	}
}

func (u UpdateFreightData) ToUpdateActiveFreightParams(
	route routes.SimpleRouteResponse,
	freightId int64,
) db.UpdateActiveFreightParams {
	return db.UpdateActiveFreightParams{
		Latitude:  u.OriginLatitude,
		Longitude: u.OriginLongitude,
		Duration:  route.Summary.SimpleRoute.Duration.Text,
		Distance:  route.Summary.SimpleRoute.Distance.Text,
		ID:        freightId,
	}
}
