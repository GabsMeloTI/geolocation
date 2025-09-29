package dashboard

import (
	db "geolocation/db/sqlc"
)

type Response struct {
	UserID                                 int64                 `json:"user_id"`
	DriverID                               int64                 `json:"driver_id"`
	TotalFreightCompleted                  float64               `json:"total_freight_completed"`
	ComparisonPreviousMonthTotalFreight    float64               `json:"comparison_previous_month_total_freight"`
	TotalReceivable                        float64               `json:"total_receivable"`
	CustomersServed                        int64                 `json:"customers_served"`
	ComparisonPreviousMonthCustomersServed float64               `json:"comparison_previous_month_customers_served"`
	Proposals                              int64                 `json:"proposals"`
	ComparisonPreviousMonthProposals       float64               `json:"comparison_previous_month_proposals"`
	FreightHistory                         []FreightHistory      `json:"freight_history"`
	FutureFreights                         []FutureFreights      `json:"future_freights"`
	Calendar                               []Calendar            `json:"calendar"`
	MonthlyBilling                         []MonthlyBilling      `json:"monthly_billing"`
	DriverEnterprise                       []DriverEnterprise    `json:"driver_enterprise"`
	TrailerEnterprise                      []TrailerEnterprise   `json:"trailer_enterprise"`
	TractUnitEnterprise                    []TractUnitEnterprise `json:"tract_unit_enterprise"`
}

type DashboardDriver struct {
	UserID                            int64   `json:"user_id"`
	DriverID                          int64   `json:"driver_id"`
	TotalFretesFinalizadosMesAtual    float64 `json:"total_fretes_finalizados_mes_atual"`
	TotalAReceberMesAtual             float64 `json:"total_a_receber_mes_atual"`
	ClientesAtendidosMesAtual         int64   `json:"clientes_atendidos_mes_atual"`
	TotalFretesFinalizadosMesAnterior float64 `json:"total_fretes_finalizados_mes_anterior"`
	ClientesAtendidosMesAnterior      int64   `json:"clientes_atendidos_mes_anterior"`
}

type FreightHistory struct {
	NameEnterprise string  `json:"name_enterprise"`
	Freight        int64   `json:"freight"`
	Price          float64 `json:"price"`
	DateDelivery   string  `json:"date_delivery"`
	TimeDelivery   string  `json:"time_delivery"`
}

type FutureFreights struct {
	Origin          string `json:"origin"`
	Destination     string `json:"destination"`
	DatePickup      string `json:"date_pickup"`
	TimePickup      string `json:"time_pickup"`
	AdvertisementId int64  `json:"advertisement_id"`
}

type Calendar struct {
	Situation  string `json:"situation"`
	DatePickup string `json:"date_pickup"`
	TimePickup string `json:"time_pickup"`
}

type MonthlyBilling struct {
	Year        int64   `json:"year"`
	Month       int64   `json:"month"`
	TotalBilled float64 `json:"total_billed"`
}

type DriverEnterprise struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	LicenseNumber   string `json:"license_number"`
	LicenseCategory string `json:"license_category"`
	Disponibilidade string `json:"disponibilidade"`
	RecesFinished   int64  `json:"reces_finished"`
}

type TrailerEnterprise struct {
	ID           int64   `json:"id"`
	Model        string  `json:"model"`
	BodyType     string  `json:"body_type"`
	LoadCapacity float64 `json:"load_capacity"`
}

type TractUnitEnterprise struct {
	ID       int64  `json:"id"`
	Model    string `json:"model"`
	UnitType string `json:"unit_type"`
	Capacity string `json:"capacity"`
}

func convertFreightHistory(rows []db.GetDashboardHistRow) []FreightHistory {
	result := make([]FreightHistory, len(rows))
	for i, r := range rows {
		result[i] = FreightHistory{
			NameEnterprise: r.HistoricoUserName,
			Freight:        r.HistoricoFreteID,
			Price:          r.HistoricoPrice,
			DateDelivery:   r.HistoricoDeliveryDate.Format("2006-01-02"),
			TimeDelivery:   r.HistoricoDeliveryDate.Format("15:04:05"),
		}
	}
	return result
}

func convertFutureFreights(rows []db.GetDashboardFutureRow) []FutureFreights {
	result := make([]FutureFreights, len(rows))
	for i, r := range rows {
		result[i] = FutureFreights{
			Origin:          r.LembretesOrigin,
			Destination:     r.LembretesDestination,
			DatePickup:      r.LembretesPickupDate.Format("2006-01-02"),
			TimePickup:      r.LembretesPickupDate.Format("15:04:05"),
			AdvertisementId: r.AdvertisementID,
		}
	}
	return result
}

func convertCalendar(rows []db.GetDashboardCalendarRow) []Calendar {
	result := make([]Calendar, len(rows))
	for i, r := range rows {
		result[i] = Calendar{
			Situation:  r.Situation,
			DatePickup: r.PickupDate.Format("2006-01-02"),
			TimePickup: r.PickupDate.Format("15:04:05"),
		}
	}
	return result
}

func convertMonthlyBilling(rows []db.GetDashboardFaturamentoRow) []MonthlyBilling {
	result := make([]MonthlyBilling, len(rows))
	for i, r := range rows {
		result[i] = MonthlyBilling{
			Year:        r.Ano,
			Month:       r.Mes,
			TotalBilled: r.TotalFaturado,
		}
	}
	return result
}

func convertDriverEnterprise(rows []db.GetDashboardDriverEnterpriseRow) []DriverEnterprise {
	result := make([]DriverEnterprise, len(rows))
	for i, r := range rows {
		result[i] = DriverEnterprise{
			ID:              r.ID,
			Name:            r.Name,
			LicenseNumber:   r.LicenseNumber,
			LicenseCategory: r.LicenseCategory,
			Disponibilidade: r.Disponibilidade,
			RecesFinished:   r.RacesFinished,
		}
	}
	return result
}

func convertTrailerEnterprise(rows []db.GetDashboardTrailerEnterpriseRow) []TrailerEnterprise {
	result := make([]TrailerEnterprise, len(rows))
	for i, r := range rows {
		result[i] = TrailerEnterprise{
			ID:           r.ID,
			Model:        r.Model,
			BodyType:     r.BodyType.String,
			LoadCapacity: r.LoadCapacity.Float64,
		}
	}
	return result
}

func convertTractUnitEnterprise(rows []db.GetDashboardTractUnitEnterpriseRow) []TractUnitEnterprise {
	result := make([]TractUnitEnterprise, len(rows))
	for i, r := range rows {
		result[i] = TractUnitEnterprise{
			ID:       r.ID,
			Model:    r.Model,
			UnitType: r.UnitType.String,
			Capacity: r.Capacity,
		}
	}
	return result
}
