package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	"time"
)

type InterfaceService interface {
	GetDashboardService(ctx context.Context, id int64, idProfile int64, startDate, endDate *time.Time) (Response, error)
}

type Service struct {
	repo InterfaceRepository
}

func NewDashboardService(repo InterfaceRepository) *Service {
	return &Service{repo: repo}
}

func (p *Service) GetDashboardService(ctx context.Context, id int64, idProfile int64, startDate, endDate *time.Time) (Response, error) {
	resultGetPlans, err := p.repo.GetProfileById(ctx, idProfile)
	if err != nil {
		return Response{}, err
	}

	var startVal, endVal time.Time
	if startDate != nil && !startDate.IsZero() {
		startVal = *startDate
	}
	if endDate != nil && !endDate.IsZero() {
		endVal = *endDate
	}

	result, err := p.repo.GetDashboardDriver(ctx, db.GetDashboardDriverParams{
		ID:      id,
		Column2: startVal,
		Column3: endVal,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result = db.GetDashboardDriverRow(DashboardDriver{})
		} else {
			return Response{}, err
		}
	}

	resultHist, errHist := p.repo.GetDashboardHist(ctx, db.GetDashboardHistParams{
		ID:      id,
		Column2: startVal,
		Column3: endVal,
	})
	if errHist != nil {
		if errors.Is(errHist, sql.ErrNoRows) {
			resultHist = nil
		} else {
			return Response{}, errHist
		}
	}

	resultFuture, errFuture := p.repo.GetDashboardFuture(ctx, db.GetDashboardFutureParams{
		UserID:  id,
		Column2: startVal,
		Column3: endVal,
	})
	if errFuture != nil {
		if errors.Is(errFuture, sql.ErrNoRows) {
			resultFuture = nil
		} else {
			return Response{}, errFuture
		}
	}

	resultCalendar, errCalendar := p.repo.GetDashboardCalendar(ctx, db.GetDashboardCalendarParams{
		UserID:  id,
		Column2: startVal,
		Column3: endVal,
	})
	if errCalendar != nil {
		if errors.Is(errCalendar, sql.ErrNoRows) {
			resultCalendar = nil
		} else {
			return Response{}, errCalendar
		}
	}

	resultFaturamento, errFaturamento := p.repo.GetDashboardFaturamento(ctx, db.GetDashboardFaturamentoParams{
		InterestedUserID: id,
		Column2:          startVal,
		Column3:          endVal,
	})
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

	resultTractUnitEnterprise, errTractUnitEnterprise := p.repo.GetDashboardTractUnitEnterprise(ctx, db.GetDashboardTractUnitEnterpriseParams{
		UserID:  id,
		Column2: startVal,
		Column3: endVal,
	})
	if errTractUnitEnterprise != nil {
		if errors.Is(errTractUnitEnterprise, sql.ErrNoRows) {
			resultTractUnitEnterprise = nil
		} else {
			return Response{}, errTractUnitEnterprise
		}
	}

	resultTrailerEnterprise, errTrailerEnterprise := p.repo.GetDashboardTrailerEnterprise(ctx, db.GetDashboardTrailerEnterpriseParams{
		UserID:  id,
		Column2: startVal,
		Column3: endVal,
	})
	if errTrailerEnterprise != nil {
		if errors.Is(errTrailerEnterprise, sql.ErrNoRows) {
			resultTrailerEnterprise = nil
		} else {
			return Response{}, errTrailerEnterprise
		}
	}

	resultOffersFor, errOffersFor := p.repo.GetOffersForDashboard(ctx, db.GetOffersForDashboardParams{
		AdvertisementUserID: id,
		Column2:             startVal,
		Column3:             endVal,
	})
	if errOffersFor != nil {
		if errors.Is(errOffersFor, sql.ErrNoRows) {
			resultOffersFor.TotalOffersMesAtual = 0
		} else {
			return Response{}, errOffersFor
		}
	}

	comparisonFreight := calcPercentage(result.TotalFretesFinalizadosMesAtual, result.TotalFretesFinalizadosMesAnterior)
	comparisonCustomers := calcPercentage(float64(result.ClientesAtendidosMesAtual), float64(result.ClientesAtendidosMesAnterior))
	comparisonProposals := calcPercentage(float64(resultOffersFor.TotalOffersMesAtual), float64(resultOffersFor.TotalOffersMesAnterior))

	var response Response
	switch resultGetPlans.Name {
	case "Driver":
		response = Response{
			UserID:                                 result.UserID,
			DriverID:                               result.DriverID,
			TotalFreightCompleted:                  result.TotalFretesFinalizadosMesAtual,
			ComparisonPreviousMonthTotalFreight:    comparisonFreight,
			TotalReceivable:                        result.TotalAReceberMesAtual,
			CustomersServed:                        result.ClientesAtendidosMesAtual,
			ComparisonPreviousMonthCustomersServed: comparisonCustomers,
			Proposals:                              resultOffersFor.TotalOffersMesAtual,
			ComparisonPreviousMonthProposals:       comparisonProposals,
			FreightHistory:                         convertFreightHistory(resultHist),
			FutureFreights:                         convertFutureFreights(resultFuture),
			Calendar:                               convertCalendar(resultCalendar),
			MonthlyBilling:                         convertMonthlyBilling(resultFaturamento),
		}
	case "Carrier":
		response = Response{
			UserID:                                 result.UserID,
			DriverID:                               result.DriverID,
			TotalFreightCompleted:                  result.TotalFretesFinalizadosMesAtual,
			ComparisonPreviousMonthTotalFreight:    comparisonFreight,
			TotalReceivable:                        result.TotalAReceberMesAtual,
			CustomersServed:                        result.ClientesAtendidosMesAtual,
			ComparisonPreviousMonthCustomersServed: comparisonCustomers,
			Proposals:                              resultOffersFor.TotalOffersMesAtual,
			ComparisonPreviousMonthProposals:       comparisonProposals,
			FreightHistory:                         convertFreightHistory(resultHist),
			FutureFreights:                         convertFutureFreights(resultFuture),
			Calendar:                               convertCalendar(resultCalendar),
			MonthlyBilling:                         convertMonthlyBilling(resultFaturamento),
			DriverEnterprise:                       convertDriverEnterprise(resultDriverEnterprise),
			TrailerEnterprise:                      convertTrailerEnterprise(resultTrailerEnterprise),
			TractUnitEnterprise:                    convertTractUnitEnterprise(resultTractUnitEnterprise),
		}
	case "Shipper":
		response = Response{
			UserID:                                 result.UserID,
			DriverID:                               result.DriverID,
			TotalFreightCompleted:                  result.TotalFretesFinalizadosMesAtual,
			ComparisonPreviousMonthTotalFreight:    comparisonFreight,
			TotalReceivable:                        result.TotalAReceberMesAtual,
			CustomersServed:                        result.ClientesAtendidosMesAtual,
			ComparisonPreviousMonthCustomersServed: comparisonCustomers,
			Proposals:                              resultOffersFor.TotalOffersMesAtual,
			ComparisonPreviousMonthProposals:       comparisonProposals,
			FreightHistory:                         convertFreightHistory(resultHist),
			FutureFreights:                         convertFutureFreights(resultFuture),
			Calendar:                               convertCalendar(resultCalendar),
			MonthlyBilling:                         convertMonthlyBilling(resultFaturamento),
		}
	default:
		return Response{}, fmt.Errorf("plano inválido")
	}

	return response, nil
}

func calcPercentage(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / previous) * 100
}
