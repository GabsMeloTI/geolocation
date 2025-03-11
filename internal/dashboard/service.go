package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
)

type InterfaceService interface {
	GetDashboardService(ctx context.Context, id int64, idProfile int64) (Response, error)
}

type Service struct {
	repo InterfaceRepository
}

func NewDashboardService(repo InterfaceRepository) *Service {
	return &Service{repo: repo}
}

func (p *Service) GetDashboardService(ctx context.Context, id int64, idProfile int64) (Response, error) {
	resultGetPlans, err := p.repo.GetProfileById(ctx, idProfile)
	if err != nil {
		return Response{}, err
	}

	result, err := p.repo.GetDashboardDriver(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result = db.GetDashboardDriverRow(DashboardDriver{})
		} else {
			return Response{}, err
		}
	}

	resultHist, errHist := p.repo.GetDashboardHist(ctx, id)
	if errHist != nil {
		if errors.Is(errHist, sql.ErrNoRows) {
			resultHist = nil
		} else {
			return Response{}, errHist
		}
	}

	resultFuture, errFuture := p.repo.GetDashboardFuture(ctx, id)
	if errFuture != nil {
		if errors.Is(errFuture, sql.ErrNoRows) {
			resultFuture = nil
		} else {
			return Response{}, errFuture
		}
	}

	resultCalendar, errCalendar := p.repo.GetDashboardCalendar(ctx, id)
	if errCalendar != nil {
		if errors.Is(errCalendar, sql.ErrNoRows) {
			resultCalendar = nil
		} else {
			return Response{}, errCalendar
		}
	}

	resultFaturamento, errFaturamento := p.repo.GetDashboardFaturamento(ctx, id)
	if errFaturamento != nil {
		if errors.Is(errFaturamento, sql.ErrNoRows) {
			resultFaturamento = nil
		} else {
			return Response{}, errFaturamento
		}
	}

	resultDriverEnterprise, errDriverEnterprise := p.repo.GetDashboardDriverEnterprise(ctx, id)
	if errDriverEnterprise != nil {
		if errors.Is(errDriverEnterprise, sql.ErrNoRows) {
			resultDriverEnterprise = nil
		} else {
			return Response{}, errDriverEnterprise
		}
	}

	resultTractUnitEnterprise, errTractUnitEnterprise := p.repo.GetDashboardTractUnitEnterprise(ctx, id)
	if errTractUnitEnterprise != nil {
		if errors.Is(errTractUnitEnterprise, sql.ErrNoRows) {
			resultTractUnitEnterprise = nil
		} else {
			return Response{}, errTractUnitEnterprise
		}
	}

	resultTrailerEnterprise, errTrailerEnterprise := p.repo.GetDashboardTrailerEnterprise(ctx, id)
	if errTrailerEnterprise != nil {
		if errors.Is(errTrailerEnterprise, sql.ErrNoRows) {
			resultTrailerEnterprise = nil
		} else {
			return Response{}, errTrailerEnterprise
		}
	}

	resultOffersFor, errOffersFor := p.repo.GetOffersForDashboard(ctx, idProfile)
	if errOffersFor != nil {
		if errors.Is(errOffersFor, sql.ErrNoRows) {
			resultOffersFor.TotalOffers = 0
		} else {
			return Response{}, errOffersFor
		}
	}

	var response Response
	switch resultGetPlans.Name {
	case "Driver":
		response = Response{
			UserID:                result.UserID,
			DriverID:              result.DriverID,
			TotalFreightCompleted: result.TotalFretesFinalizados,
			TotalReceivable:       result.TotalAReceber,
			CustomersServed:       result.ClientesAtendidos,
			Proposals:             resultOffersFor.TotalOffers,
			FreightHistory:        convertFreightHistory(resultHist),
			FutureFreights:        convertFutureFreights(resultFuture),
			Calendar:              convertCalendar(resultCalendar),
			MonthlyBilling:        convertMonthlyBilling(resultFaturamento),
		}
	case "Carrier":
		response = Response{
			UserID:                result.UserID,
			DriverID:              result.DriverID,
			TotalFreightCompleted: result.TotalFretesFinalizados,
			TotalReceivable:       result.TotalAReceber,
			CustomersServed:       result.ClientesAtendidos,
			FreightHistory:        convertFreightHistory(resultHist),
			FutureFreights:        convertFutureFreights(resultFuture),
			Calendar:              convertCalendar(resultCalendar),
			MonthlyBilling:        convertMonthlyBilling(resultFaturamento),
			DriverEnterprise:      convertDriverEnterprise(resultDriverEnterprise),
			TrailerEnterprise:     convertTrailerEnterprise(resultTrailerEnterprise),
			TractUnitEnterprise:   convertTractUnitEnterprise(resultTractUnitEnterprise),
		}
	case "Shipper":
		response = Response{
			UserID:                result.UserID,
			DriverID:              result.DriverID,
			TotalFreightCompleted: result.TotalFretesFinalizados,
			TotalReceivable:       result.TotalAReceber,
			CustomersServed:       result.ClientesAtendidos,
			FreightHistory:        convertFreightHistory(resultHist),
			FutureFreights:        convertFutureFreights(resultFuture),
			Calendar:              convertCalendar(resultCalendar),
			MonthlyBilling:        convertMonthlyBilling(resultFaturamento),
		}
	default:
		return Response{}, fmt.Errorf("plano inválido")
	}

	return response, nil
}
