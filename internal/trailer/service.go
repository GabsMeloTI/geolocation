package trailer

import (
	"context"
	"database/sql"
	"errors"
)

type InterfaceService interface {
	CreateTrailerService(ctx context.Context, data CreateTrailerRequest) (TrailerResponse, error)
	UpdateTrailerService(ctx context.Context, data UpdateTrailerRequest) (TrailerResponse, error)
	DeleteTrailerService(ctx context.Context, id int64) error
	GetTrailerService(ctx context.Context, id int64) ([]TrailerResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewTrailersService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateTrailerService(ctx context.Context, data CreateTrailerRequest) (TrailerResponse, error) {
	arg := data.ParseCreateToTrailer()

	result, err := p.InterfaceService.CreateTrailer(ctx, arg)
	if err != nil {
		return TrailerResponse{}, err
	}

	createTrailerService := TrailerResponse{}
	createTrailerService.ParseFromTrailerObject(result)

	return createTrailerService, nil
}

func (p *Service) UpdateTrailerService(ctx context.Context, data UpdateTrailerRequest) (TrailerResponse, error) {
	_, err := p.InterfaceService.GetTrailerById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return TrailerResponse{}, errors.New("Trailer not found")
	}
	if err != nil {
		return TrailerResponse{}, err
	}

	arg := data.ParseUpdateToTrailer()

	result, err := p.InterfaceService.UpdateTrailer(ctx, arg)
	if err != nil {
		return TrailerResponse{}, err
	}

	updateTrailerService := TrailerResponse{}
	updateTrailerService.ParseFromTrailerObject(result)

	return updateTrailerService, nil
}

func (p *Service) DeleteTrailerService(ctx context.Context, id int64) error {
	_, err := p.InterfaceService.GetTrailerById(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("Trailer not found")
	}
	if err != nil {
		return err
	}

	err = p.InterfaceService.DeleteTrailer(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *Service) GetTrailerService(ctx context.Context, id int64) ([]TrailerResponse, error) {
	result, err := p.InterfaceService.GetTrailerByUserId(ctx, id)
	if err != nil {
		return []TrailerResponse{}, err
	}

	var getAllTrailers []TrailerResponse
	for _, trailer := range result {
		getTrailerResponse := TrailerResponse{}
		getTrailerResponse.ParseFromTrailerObject(trailer)
		getAllTrailers = append(getAllTrailers, getTrailerResponse)
	}

	return getAllTrailers, nil
}
