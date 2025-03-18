package tractor_unit

import (
	"database/sql"
	"time"

	db "geolocation/db/sqlc"
)

type CreateTractorUnitRequest struct {
	LicensePlate    string  `json:"license_plate"`
	DriverID        int64   `json:"driver_id"`
	UserID          int64   `json:"user_id"`
	Chassis         string  `json:"chassis"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	ManufactureYear int64   `json:"manufacture_year"`
	EnginePower     string  `json:"engine_power"`
	UnitType        string  `json:"unit_type"        validate:"oneof=stump truck tractor_unit"`
	CanCouple       bool    `json:"can_couple"`
	Height          float64 `json:"height"`
	State           string  `json:"state"`
	Renavan         string  `json:"renavan"`
	Capacity        string  `json:"capacity"`
	Width           float64 `json:"width"`
	Length          float64 `json:"length"`
	Color           string  `json:"color"`
	Axles           int64   `json:"axles"`
}
type CreateTractorUnitDto struct {
	CreateTractorUnitRequest CreateTractorUnitRequest
	UserID                   int64 `json:"user_id"`
}
type UpdateTractorUnitRequest struct {
	ID              int64   `json:"id"`
	LicensePlate    string  `json:"license_plate"`
	DriverID        int64   `json:"driver_id"`
	Chassis         string  `json:"chassis"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	ManufactureYear int64   `json:"manufacture_year"`
	EnginePower     string  `json:"engine_power"`
	UnitType        string  `json:"unit_type"`
	Height          float64 `json:"height"`
	UserID          int64   `json:"user_id"`
	State           string  `json:"state"`
	Renavan         string  `json:"renavan"`
	Capacity        string  `json:"capacity"`
	Width           float64 `json:"width"`
	Length          float64 `json:"length"`
	Color           string  `json:"color"`
	Axles           int64   `json:"axles"`
}
type UpdateTractorUnitDto struct {
	UpdateTractorUnitRequest UpdateTractorUnitRequest
	UserID                   int64 `json:"user_id"`
}
type TractorUnitResponse struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	LicensePlate    string     `json:"license_plate"`
	DriverID        int64      `json:"driver_id"`
	Chassis         string     `json:"chassis"`
	Brand           string     `json:"brand"`
	Model           string     `json:"model"`
	ManufactureYear int64      `json:"manufacture_year"`
	EnginePower     string     `json:"engine_power"`
	UnitType        string     `json:"unit_type"`
	Height          float64    `json:"height"`
	State           string     `json:"state"`
	Renavan         string     `json:"renavan"`
	Capacity        string     `json:"capacity"`
	CanCouple       bool       `json:"can_couple"`
	Width           float64    `json:"width"`
	Length          float64    `json:"length"`
	Color           string     `json:"color"`
	Status          bool       `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	Axles           int64      `json:"axles"`
}
func (p *CreateTractorUnitDto) ParseCreateToTractorUnit() db.CreateTractorUnitParams {
	arg := db.CreateTractorUnitParams{
		Axles:        p.CreateTractorUnitRequest.Axles,
		LicensePlate: p.CreateTractorUnitRequest.LicensePlate,
		DriverID:     p.CreateTractorUnitRequest.DriverID,
		UserID:       p.UserID,
		Chassis:      p.CreateTractorUnitRequest.Chassis,
		Brand:        p.CreateTractorUnitRequest.Brand,
		Model:        p.CreateTractorUnitRequest.Model,
		ManufactureYear: sql.NullInt64{
			Int64: p.CreateTractorUnitRequest.ManufactureYear,
			Valid: true,
		},
		EnginePower: sql.NullString{
			String: p.CreateTractorUnitRequest.EnginePower,
			Valid:  true,
		},
		UnitType: sql.NullString{
			String: p.CreateTractorUnitRequest.UnitType,
			Valid:  true,
		},
		CanCouple: sql.NullBool{
			Bool:  p.CreateTractorUnitRequest.CanCouple,
			Valid: true,
		},
		Height:   p.CreateTractorUnitRequest.Height,
		State:    p.CreateTractorUnitRequest.State,
		Renavan:  p.CreateTractorUnitRequest.Renavan,
		Capacity: p.CreateTractorUnitRequest.Capacity,
		Width:    p.CreateTractorUnitRequest.Width,
		Length:   p.CreateTractorUnitRequest.Length,
		Color:    p.CreateTractorUnitRequest.Color,
	}
	return arg
}
func (p *UpdateTractorUnitDto) ParseUpdateToTractorUnit() db.UpdateTractorUnitParams {
	arg := db.UpdateTractorUnitParams{
		LicensePlate: p.UpdateTractorUnitRequest.LicensePlate,
		DriverID:     p.UpdateTractorUnitRequest.DriverID,
		UserID:       p.UserID,
		Chassis:      p.UpdateTractorUnitRequest.Chassis,
		Brand:        p.UpdateTractorUnitRequest.Brand,
		Model:        p.UpdateTractorUnitRequest.Model,
		ManufactureYear: sql.NullInt64{
			Int64: p.UpdateTractorUnitRequest.ManufactureYear,
			Valid: true,
		},
		EnginePower: sql.NullString{
			String: p.UpdateTractorUnitRequest.EnginePower,
			Valid:  true,
		},
		UnitType: sql.NullString{
			String: p.UpdateTractorUnitRequest.UnitType,
			Valid:  true,
		},
		Height:   p.UpdateTractorUnitRequest.Height,
		State:    p.UpdateTractorUnitRequest.State,
		Renavan:  p.UpdateTractorUnitRequest.Renavan,
		Capacity: p.UpdateTractorUnitRequest.Capacity,
		Width:    p.UpdateTractorUnitRequest.Width,
		Length:   p.UpdateTractorUnitRequest.Length,
		Color:    p.UpdateTractorUnitRequest.Color,
		ID:       p.UpdateTractorUnitRequest.ID,
		Axles:    p.UpdateTractorUnitRequest.Axles,
	}
	return arg
}
func (p *TractorUnitResponse) ParseFromTractorUnitObject(result db.TractorUnit) {
	p.ID = result.ID
	p.DriverID = result.DriverID
	p.UserID = result.UserID
	p.LicensePlate = result.LicensePlate
	p.Chassis = result.Chassis
	p.Brand = result.Brand
	p.Model = result.Model
	p.ManufactureYear = result.ManufactureYear.Int64
	p.EnginePower = result.EnginePower.String
	p.UnitType = result.UnitType.String
	p.Height = result.Height
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	p.State = result.State
	p.Renavan = result.Renavan
	p.Capacity = result.Capacity
	p.Width = result.Width
	p.Length = result.Length
	p.Color = result.Color
	p.Axles = result.Axles
	p.CanCouple = result.CanCouple.Bool
}


//mock check plate
type User struct {
    FirstName   string `json:"first_name"`
    Email       string `json:"email"`
    Cellphone   string `json:"cellphone"`
    Notification string `json:"notification"`
}

type Registro struct {
    NumeroAutoInfracao              string `json:"numeroautoinfracao"`
    DataDaInfracao                  string `json:"datadainfracao"`
    Exigibilidade                   string `json:"exigibilidade"`
    Infracao                        string `json:"infracao"`
    Orgao                           string `json:"orgao"`
    ConsultaDetalheExisteErro       string `json:"consultadetalhe_existe_erro"`
    ConsultaDetalheMensagem         string `json:"consultadetalhe_mensagem"`
    DetalheCadastramentoInfracao    string `json:"detalhe_cadastramento_infracao"`
    DetalheCodInfracao              string `json:"detalhe_cod_infracao"`
    DetalheCodMunEmplacamento       string `json:"detalhe_cod_mun_emplacamento"`
    DetalheCodMunInfracao           string `json:"detalhe_cod_mun_infracao"`
    DetalheDtEmissaoPenalidade      string `json:"detalhe_dt_emissao_penalidade"`
    DetalheDtInfracao               string `json:"detalhe_dt_infracao"`
    DetalheDtNotificacaoInfracao    string `json:"detalhe_dt_notificacao_infracao"`
    DetalheHrInfracao               string `json:"detalhe_hr_infracao"`
    DetalheLimitePermitido          string `json:"detalhe_limite_permitido"`
    DetalheLocalInfracao            string `json:"detalhe_local_infracao"`
    DetalheAmrcaModelo              string `json:"detalhe_amrcamodelo"`
    DetalheMedicaoConsiderada       string `json:"detalhe_medicao_considerada"`
    DetalheMedicaoReal              string `json:"detalhe_medicao_real"`
    DetalheNumAutoInfracao          string `json:"detalhe_num_auto_infracao"`
    DetalheOrgaoAutuador            string `json:"detalhe_orgao_autuador"`
    DetalhePlaca                    string `json:"detalhe_placa"`
    DetalheTipoAutoInfracao         string `json:"detalhe_tipo_auto_infracao"`
    DetalheUfJurisdicaoVeiculo      string `json:"detalhe_uf_jurisdicao_veiculo"`
    DetalheUfOrgaoAutuador          string `json:"detalhe_uf_orgao_autuador"`
    DetalheUfPlaca                  string `json:"detalhe_uf_placa"`
    DetalheUnidadeMedida            string `json:"detalhe_unidade_medida"`
    DetalheValorInfracao            string `json:"detalhe_valor_infracao"`
    DadosDaSuspensaoAceiteUfJurisdicao string `json:"dadosdasuspensao_aceite_uf_jurisdicao"`
    DadosDaSuspensaoDataRegistro       string `json:"dadosdasuspensao_data_registro"`
    DadosDaSuspensaoOrigem             string `json:"dadosdasuspensao_origem"`
    DadosDaSuspensaoTipo               string `json:"dadosdasuspensao_tipo"`
    DadosInfratorCnhCondutor           string `json:"dadosinfrator_cnh_condutor"`
    DadosInfratorCnhInfrator           string `json:"dadosinfrator_cnh_infrator"`
    DadosDoPagamentoDtPagamento        string `json:"dadosdopagamento_dt_pagamento"`
    DadosDoPagamentoDtDoRegistroDoPgmto string `json:"dadosdopagamento_dt_do_registro_do_pgmto"`
    DadosDoPagamentoUfPagamento        string `json:"dadosdopagamento_uf_pagamento"`
    DadosDoPagamentoValorPago          string `json:"dadosdopagamento_valor_pago"`
    DadosDoPagamentoDadosPgmto         string `json:"dadosdopagamento_dados_pgmto"`
}

type Data struct {
    Alerta                      string     `json:"alerta"`
    Placa                       string     `json:"placa"`
    QuantidadeOcorrencias       int        `json:"quantidade_ocorrencias"`
    QuantidadeOcorrenciasTotal  int        `json:"quantidade_ocorrencias_total"`
    Registros                   []Registro `json:"registros"`
}

type CheckPlateResponse struct {
    User     User   `json:"user"`
    Balance  string `json:"balance"`
    Error    bool   `json:"error"`
    Message  string `json:"message"`
    Homolog  string `json:"homolog"`
    Data     Data   `json:"data"`
}