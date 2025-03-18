package tractor_unit

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
)

type InterfaceService interface {
	CreateTractorUnitService(ctx context.Context, data CreateTractorUnitDto) (TractorUnitResponse, error)
	UpdateTractorUnitService(ctx context.Context, data UpdateTractorUnitDto) (TractorUnitResponse, error)
	DeleteTractorUnitService(ctx context.Context, id, idUser int64) error
	GetTractorUnitService(ctx context.Context, id int64) ([]TractorUnitResponse, error)
	GetTractorUnitByIdService(ctx context.Context, id int64) (TractorUnitResponse, error)
	CheckPlate(plate string) (interface{}, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewTractorUnitsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateTractorUnitService(ctx context.Context, data CreateTractorUnitDto) (TractorUnitResponse, error) {
	arg := data.ParseCreateToTractorUnit()

	result, err := p.InterfaceService.CreateTractorUnit(ctx, arg)
	if err != nil {
		return TractorUnitResponse{}, err
	}

	createTractorUnitService := TractorUnitResponse{}
	createTractorUnitService.ParseFromTractorUnitObject(result)

	return createTractorUnitService, nil
}

func (p *Service) UpdateTractorUnitService(ctx context.Context, data UpdateTractorUnitDto) (TractorUnitResponse, error) {
	_, err := p.InterfaceService.GetTractorUnitById(ctx, data.UpdateTractorUnitRequest.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return TractorUnitResponse{}, errors.New("driver not found")
	}
	if err != nil {
		return TractorUnitResponse{}, err
	}

	arg := data.ParseUpdateToTractorUnit()

	result, err := p.InterfaceService.UpdateTractorUnit(ctx, arg)
	if err != nil {
		return TractorUnitResponse{}, err
	}

	updateTractorUnitService := TractorUnitResponse{}
	updateTractorUnitService.ParseFromTractorUnitObject(result)

	return updateTractorUnitService, nil
}

func (p *Service) DeleteTractorUnitService(ctx context.Context, id, idUser int64) error {
	_, err := p.InterfaceService.GetTractorUnitById(ctx, id)
	if err != nil {
		return err
	}

	err = p.InterfaceService.DeleteTractorUnit(ctx, db.DeleteTractorUnitParams{
		ID:     id,
		UserID: idUser,
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Service) GetTractorUnitService(ctx context.Context, id int64) ([]TractorUnitResponse, error) {
	result, err := p.InterfaceService.GetTractorUnitByUserId(ctx, id)
	if err != nil {
		return []TractorUnitResponse{}, err
	}

	var getAllTractorUnit []TractorUnitResponse
	for _, tractorUnit := range result {
		getTractorUnitResponse := TractorUnitResponse{}
		getTractorUnitResponse.ParseFromTractorUnitObject(tractorUnit)
		getAllTractorUnit = append(getAllTractorUnit, getTractorUnitResponse)
	}

	return getAllTractorUnit, nil
}

func (p *Service) GetTractorUnitByIdService(ctx context.Context, id int64) (TractorUnitResponse, error) {
	result, err := p.InterfaceService.GetOneTractorUnitByUserId(ctx, id)
	if err != nil {
		return TractorUnitResponse{}, err
	}

	getTractorUnitResponse := TractorUnitResponse{}
	getTractorUnitResponse.ParseFromTractorUnitObject(result)

	return getTractorUnitResponse, nil
}

func (p *Service) CheckPlate(plate string) (interface{}, error) {
	if plate == "FUT8C76" {
		result := CheckPlateResponse{
			User: User{
				FirstName:    "GUILHERME",
				Email:        "guilherme_souza.lima@hotmail.com",
				Cellphone:    "17991109011",
				Notification: "yes",
			},
			Balance: "14,650",
			Error:   false,
			Message: "Dados válidos! Você foi tarifado em R$ 3.45!",
			Homolog: "false",
			Data: Data{
				Alerta:                     "",
				Placa:                      "FUT8C76",
				QuantidadeOcorrencias:      1,
				QuantidadeOcorrenciasTotal: 1,
				Registros: []Registro{
					{
						NumeroAutoInfracao:                  "6VA1242519",
						DataDaInfracao:                      "21/07/2024",
						Exigibilidade:                       "",
						Infracao:                            "7463",
						Orgao:                               "271070",
						ConsultaDetalheExisteErro:           "",
						ConsultaDetalheMensagem:             "",
						DetalheCadastramentoInfracao:        "",
						DetalheCodInfracao:                  "TRANSITAR EM VELOCIDADE SUPERIOR A MAXIMA PERMITIDA EM MAIS DE 20% ATE 50%",
						DetalheCodMunEmplacamento:           "",
						DetalheCodMunInfracao:               "",
						DetalheDtEmissaoPenalidade:          "",
						DetalheDtInfracao:                   "21/07/2024",
						DetalheDtNotificacaoInfracao:        "",
						DetalheHrInfracao:                   "",
						DetalheLimitePermitido:              "",
						DetalheLocalInfracao:                "AV DAS NACOES UNIDAS, PISTA CE",
						DetalheAmrcaModelo:                  "",
						DetalheMedicaoConsiderada:           "",
						DetalheMedicaoReal:                  "",
						DetalheNumAutoInfracao:              "6VA1242519",
						DetalheOrgaoAutuador:                "CET-SP",
						DetalhePlaca:                        "FUT8C76",
						DetalheTipoAutoInfracao:             "",
						DetalheUfJurisdicaoVeiculo:          "",
						DetalheUfOrgaoAutuador:              "",
						DetalheUfPlaca:                      "",
						DetalheUnidadeMedida:                "",
						DetalheValorInfracao:                "195,23",
						DadosDaSuspensaoAceiteUfJurisdicao:  "",
						DadosDaSuspensaoDataRegistro:        "",
						DadosDaSuspensaoOrigem:              "0",
						DadosDaSuspensaoTipo:                "0",
						DadosInfratorCnhCondutor:            "",
						DadosInfratorCnhInfrator:            "",
						DadosDoPagamentoDtPagamento:         "",
						DadosDoPagamentoDtDoRegistroDoPgmto: "",
						DadosDoPagamentoUfPagamento:         "",
						DadosDoPagamentoValorPago:           "",
						DadosDoPagamentoDadosPgmto:          "",
					},
				},
			},
		}
		return result, nil
	}

	result := CheckPlateResponse{
		User: User{
			FirstName:    "GUILHERME",
			Email:        "guilherme_souza.lima@hotmail.com",
			Cellphone:    "17991109011",
			Notification: "yes",
		},
		Balance: "7,750",
		Error:   false,
		Message: "Dados válidos! Você foi tarifado em R$ 3.45!",
		Homolog: "false",
		Data: Data{
			Alerta:                     "",
			Placa:                      plate,
			QuantidadeOcorrencias:      0,
			QuantidadeOcorrenciasTotal: 0,
			Registros:                  []Registro{},
		},
	}
	return result, nil
}
