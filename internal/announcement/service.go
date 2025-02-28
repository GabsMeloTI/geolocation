package announcement

import (
	"context"
	"database/sql"
	"errors"
)

type InterfaceService interface {
	CreateAnnouncementService(ctx context.Context, data CreateAnnouncementRequest) (AnnouncementResponse, error)
	UpdateAnnouncementService(ctx context.Context, data UpdateAnnouncementRequest) (AnnouncementResponse, error)
	DeleteAnnouncementService(ctx context.Context, data DeleteAnnouncementRequest) error
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewAnnouncementsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateAnnouncementService(ctx context.Context, data CreateAnnouncementRequest) (AnnouncementResponse, error) {
	arg := data.ParseCreateToAnnouncement()

	result, err := p.InterfaceService.CreateAnnouncement(ctx, arg)
	if err != nil {
		return AnnouncementResponse{}, err
	}

	createAnnouncementService := AnnouncementResponse{}
	createAnnouncementService.ParseFromAnnouncementObject(result)

	return createAnnouncementService, nil
}

func (p *Service) UpdateAnnouncementService(ctx context.Context, data UpdateAnnouncementRequest) (AnnouncementResponse, error) {
	_, err := p.InterfaceService.GetAnnouncementById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return AnnouncementResponse{}, errors.New("Announcement not found")
	}
	if err != nil {
		return AnnouncementResponse{}, err
	}

	arg := data.ParseUpdateToAnnouncement()

	result, err := p.InterfaceService.UpdateAnnouncement(ctx, arg)
	if err != nil {
		return AnnouncementResponse{}, err
	}

	updateAnnouncementService := AnnouncementResponse{}
	updateAnnouncementService.ParseFromAnnouncementObject(result)

	return updateAnnouncementService, nil
}

func (p *Service) DeleteAnnouncementService(ctx context.Context, data DeleteAnnouncementRequest) error {
	_, err := p.InterfaceService.GetAnnouncementById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("Announcement not found")
	}
	if err != nil {
		return err
	}

	arg := data.ParseDeleteToAnnouncement()
	err = p.InterfaceService.DeleteAnnouncement(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}
