package plans

import (
	"context"
	"fmt"
	db "geolocation/db/sqlc"
	"time"
)

type InterfaceService interface {
	CreateUserPlanService(ctx context.Context, data CreateUserPlanRequest) (UserPlanResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewUserPlanService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateUserPlanService(ctx context.Context, data CreateUserPlanRequest) (UserPlanResponse, error) {
	resultExist, err := p.InterfaceService.GetUserPlanByIdUser(ctx, db.GetUserPlanByIdUserParams{
		IDUser: data.IDUser,
		IDPlan: data.IDPlan,
	})
	if err == nil {
		if resultExist.ExpirationDate.After(time.Now()) {
			return UserPlanResponse{}, fmt.Errorf("usuário já possui um plano ativo")
		}

		err = p.InterfaceService.UpdateUserPlan(ctx, db.UpdateUserPlanParams{
			IDUser: data.IDUser,
			IDPlan: data.IDPlan,
		})
		if err != nil {
			return UserPlanResponse{}, err
		}
	}

	resultGetPlans, err := p.InterfaceService.GetPlansById(ctx, data.IDPlan)
	if err != nil {
		return UserPlanResponse{}, err
	}

	var expirationDate time.Time
	switch resultGetPlans.Name {
	case "trial":
		expirationDate = time.Now().AddDate(0, 0, 30)
	case "mensal":
		expirationDate = time.Now().AddDate(0, 1, 0)
	case "trimestral":
		expirationDate = time.Now().AddDate(0, 3, 0)
	default:
		return UserPlanResponse{}, fmt.Errorf("plano inválido")
	}

	var finalPrice float64
	if data.Annual {
		finalPrice = resultGetPlans.Price - (resultGetPlans.Price * 0.20)
	} else {
		finalPrice = resultGetPlans.Price
	}

	result, err := p.InterfaceService.CreateUserPlans(ctx, db.CreateUserPlansParams{
		IDUser:         data.IDUser,
		IDPlan:         data.IDPlan,
		ExpirationDate: expirationDate,
		Annual:         data.Annual,
	})
	if err != nil {
		return UserPlanResponse{}, err
	}

	createUserPlan := UserPlanResponse{}
	createUserPlan.Price = finalPrice
	createUserPlan.ParseFromPlansObject(result)

	return createUserPlan, nil
}
