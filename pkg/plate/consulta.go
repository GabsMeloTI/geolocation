package plate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func init() {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		// Fallback para o IP anterior caso a env não esteja definida
		redisAddr = "3.238.87.0:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
}

// ─── Estruturas internas (resposta da nova API de placa) ─────────────────────

type consultarPlacaAPIResponse struct {
	Status          string `json:"status"`
	Mensagem        string `json:"mensagem"`
	DataSolicitacao string `json:"data_solicitacao"`
	Dados           struct {
		InformacoesVeiculo struct {
			DadosVeiculo struct {
				Placa         string `json:"placa"`
				Chassi        string `json:"chassi"`
				AnoFabricacao string `json:"ano_fabricacao"`
				AnoModelo     string `json:"ano_modelo"`
				Marca         string `json:"marca"`
				Modelo        string `json:"modelo"`
				Cor           string `json:"cor"`
				Segmento      string `json:"segmento"`
				Combustivel   string `json:"combustivel"`
				Procedencia   string `json:"procedencia"`
				Municipio     string `json:"municipio"`
				UFMunicipio   string `json:"uf_municipio"`
			} `json:"dados_veiculo"`
			DadosTecnicos struct {
				TipoVeiculo       string `json:"tipo_veiculo"`
				SubSegmento       string `json:"sub_segmento"`
				NumeroMotor       string `json:"numero_motor"`
				NumeroCaixaCambio string `json:"numero_caixa_cambio"`
				Potencia          string `json:"potencia"`
				Cilindradas       string `json:"cilindradas"`
			} `json:"dados_tecnicos"`
			DadosCarga struct {
				NumeroEixos          string `json:"numero_eixos"`
				CapacidadeMaxTraccao string `json:"capacidade_maxima_tracao"`
				CapacidadePassageiro string `json:"capacidade_passageiro"`
			} `json:"dados_carga"`
		} `json:"informacoes_veiculo"`
	} `json:"dados"`
	Request struct {
		Placa string `json:"placa"`
	} `json:"request"`
}

// ─── Estruturas internas (resposta da nova API de multas/renainf) ─────────────

type consultarInfracoesAPIResponse struct {
	Status          string `json:"status"`
	Mensagem        string `json:"mensagem"`
	DataSolicitacao string `json:"data_solicitacao"`
	Dados           struct {
		RegistroDebitosPorInfracoes struct {
			InfracoesRenainf struct {
				PossuiInfracoes string `json:"possui_infracoes"`
				Infracoes       []struct {
					DadosInfracao struct {
						Infracao           string `json:"infracao"`
						NumeroAutoInfracao string `json:"numero_auto_infracao"`
						ValorAplicado      string `json:"valor_aplicado"`
						OrgaoAutuador      string `json:"orgao_autuador"`
						TipoAutoInfracao   string `json:"tipo_auto_infracao"`
						LocalInfracao      string `json:"local_infracao"`
						Municipio          string `json:"municipio"`
					} `json:"dados_infracao"`
					Aplicacao struct {
						UnidadeMedida      string `json:"unidade_medida"`
						LimitePermitido    string `json:"limite_permitido"`
						MedicaoConsiderada string `json:"medicao_considerada"`
						MedicaoReal        string `json:"medicao_real"`
					} `json:"aplicacao"`
					Eventos struct {
						DataHoraInfracao      string `json:"data_hora_infracao"`
						DataCadastramento     string `json:"data_cadastramento"`
						DataNotificacao       string `json:"data_notificacao"`
						DataEmissaoPenalidade string `json:"data_emissao_penalidade"`
					} `json:"eventos"`
				} `json:"infracoes"`
			} `json:"infracoes_renainf"`
		} `json:"registro_debitos_por_infracoes_renainf"`
	} `json:"dados"`
	Request struct {
		Placa string `json:"placa"`
	} `json:"request"`
}

// ─── Credenciais Basic Auth ───────────────────────────────────────────────────

func getBasicAuthCredentials() (string, string) {
	username := os.Getenv("CONSULTAR_PLACA_USERNAME")
	password := os.Getenv("CONSULTAR_PLACA_PASSWORD")
	if username == "" {
		username = "gabrielmelodsantos@gmail.com"
	}
	if password == "" {
		password = "f905cd6a53d3e76485e8cb6d67c3e0f7"
	}
	return username, password
}

// ─── Funções auxiliares ───────────────────────────────────────────────────────

// Função auxiliar para converter interface{} para string
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// Função auxiliar para converter ponteiro de int para string
func getIntPointerValue(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

// ─── ConsultarPlaca ───────────────────────────────────────────────────────────

func ConsultarPlaca(placa string) (*FullAPIResponse, error) {
	placa = strings.ToUpper(strings.TrimSpace(placa))
	placa = strings.ReplaceAll(placa, "-", "") // Remove hífens da placa
	cacheKey := "placa:" + placa
	fmt.Println(placa)

	// 🔹 Verifica cache Redis apenas para dados da placa (sem multas)
	var fullResp FullAPIResponse
	var dadosPlacaCached bool

	if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		var cachedResp FullAPIResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			fullResp = cachedResp
			dadosPlacaCached = true
		}
	}

	username, password := getBasicAuthCredentials()

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 1. Consulta dados do veículo (apenas se não estiver em cache)
	if !dadosPlacaCached {
		veiculoURL := fmt.Sprintf("https://api.consultarplaca.com.br/v2/consultarPlaca?placa=%s", placa)

		req, err := http.NewRequest("GET", veiculoURL, nil)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar requisição do veículo: %w", err)
		}
		req.SetBasicAuth(username, password)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erro ao enviar requisição do veículo: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta do veículo: %w", err)
		}

		var apiResp consultarPlacaAPIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("erro ao decodificar JSON do veículo: %w", err)
		}

		if apiResp.Status != "ok" {
			return nil, fmt.Errorf("API de placa retornou erro: %s", apiResp.Mensagem)
		}

		dv := apiResp.Dados.InformacoesVeiculo.DadosVeiculo
		dt := apiResp.Dados.InformacoesVeiculo.DadosTecnicos
		dc := apiResp.Dados.InformacoesVeiculo.DadosCarga

		fullResp = FullAPIResponse{
			Error:   false,
			Message: apiResp.Mensagem,
			Data: Response{
				Placa:                dv.Placa,
				Chassi:               dv.Chassi,
				Modelo:               dv.Modelo,
				Marca:                dv.Marca,
				Ano:                  dv.AnoFabricacao,
				AnoModelo:            dv.AnoModelo,
				AnoModelo1:           dv.AnoModelo,
				Cor:                  dv.Cor,
				Municipio:            dv.Municipio,
				Uf:                   dv.UFMunicipio,
				UfPlaca:              dv.UFMunicipio,
				Combustivel:          dv.Combustivel,
				Potencia:             dt.Potencia,
				Cilindradas:          dt.Cilindradas,
				Eixos:                dc.NumeroEixos,
				QuantidadePassageiro: dc.CapacidadePassageiro,
				TipoVeiculo: struct {
					TipoVeiculo string `json:"tipo_veiculo"`
				}{
					TipoVeiculo: dt.TipoVeiculo,
				},
				MarcaModelo: struct {
					Modelo   string `json:"modelo"`
					Marca    string `json:"marca"`
					Segmento string `json:"segmento"`
					Versao   string `json:"versao"`
				}{
					Modelo:   dv.Modelo,
					Marca:    dv.Marca,
					Segmento: dv.Segmento,
				},
				Extra: map[string]interface{}{
					"ano_fabricacao":    dv.AnoFabricacao,
					"chassi":            dv.Chassi,
					"numero_motor":      dt.NumeroMotor,
					"caixa_cambio":      dt.NumeroCaixaCambio,
					"cap_maxima_tracao": dc.CapacidadeMaxTraccao,
					"procedencia":       dv.Procedencia,
				},
				Multas: struct {
					Dados []Multa `json:"dados"`
				}{
					Dados: []Multa{},
				},
			},
		}

		// Cache apenas os dados da placa (sem multas)
		respBytes, _ := json.Marshal(fullResp)
		if err := rdb.Set(ctx, cacheKey, respBytes, 30*time.Minute).Err(); err != nil {
		}
	}

	// 2. Consulta multas (SEMPRE consulta, nunca usa cache)
	multasURL := fmt.Sprintf("https://api.consultarplaca.com.br/v2/consultarRegistrosInfracoesRenainf?placa=%s", placa)

	reqMultas, err := http.NewRequest("GET", multasURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	reqMultas.SetBasicAuth(username, password)

	respMultas, err := client.Do(reqMultas)
	if err != nil {
		fmt.Printf("erro ao buscar multas da placa: %s", placa)
		// Erro ao consultar multas - continua sem multas
	} else {
		defer respMultas.Body.Close()

		multasRespBody, _ := io.ReadAll(respMultas.Body)

		fmt.Printf("[DEBUG] Resposta API consultarplaca (multas) - Placa %s: %s\n------------------------------\n", placa, string(multasRespBody))

		var infracoesResp consultarInfracoesAPIResponse
		if err := json.Unmarshal(multasRespBody, &infracoesResp); err != nil {
			fmt.Printf("[ERROR] Falha ao decodificar multas JSON (placa %s): %v\n", placa, err)
			fullResp.Data.Multas.Dados = []Multa{}
		} else if infracoesResp.Status != "ok" {
			fmt.Printf("[WARN] Status não ok nas multas (placa %s): %s\n", placa, infracoesResp.Mensagem)
			fullResp.Data.Multas.Dados = []Multa{}
		} else {
			infracoes := infracoesResp.Dados.RegistroDebitosPorInfracoes.InfracoesRenainf.Infracoes
			multas := make([]Multa, 0, len(infracoes))
			for _, inf := range infracoes {
				multas = append(multas, Multa{
					NumeroAutoInfracao:           inf.DadosInfracao.NumeroAutoInfracao,
					Infracao:                     inf.DadosInfracao.Infracao,
					DetalheOrgaoAutuador:         inf.DadosInfracao.OrgaoAutuador,
					DetalheValorInfracao:         inf.DadosInfracao.ValorAplicado,
					DetalheTipoAutoInfracao:      inf.DadosInfracao.TipoAutoInfracao,
					DetalheLocalInfracao:         inf.DadosInfracao.LocalInfracao,
					DetalheUnidadeMedida:         inf.Aplicacao.UnidadeMedida,
					DetalheLimitePermitido:       inf.Aplicacao.LimitePermitido,
					DetalheMedicaoConsiderada:    inf.Aplicacao.MedicaoConsiderada,
					DetalheMedicaoReal:           inf.Aplicacao.MedicaoReal,
					DetalheHoraInfracao:          inf.Eventos.DataHoraInfracao,
					DetalheDataCadastramento:     inf.Eventos.DataCadastramento,
					DetalheDataNotificacao:       inf.Eventos.DataNotificacao,
					DetalheDataEmissaoPenalidade: inf.Eventos.DataEmissaoPenalidade,
				})
			}
			fullResp.Data.Multas.Dados = multas
		}
	}

	return &fullResp, nil
}

// ─── ConsultarMultas ──────────────────────────────────────────────────────────

type MultaA struct {
	NumeroAutoInfracao           string `json:"numeroautoinfracao"`
	DataInfracao                 string `json:"datadainfracao"`
	Exigibilidade                string `json:"exigibilidade"`
	Infracao                     string `json:"infracao"`
	Orgao                        string `json:"orgao"`
	ConsultaDetalheExisteErro    string `json:"consultadetalhe_existe_erro"`
	ConsultaDetalheMensagem      string `json:"consultadetalhe_mensagem"`
	DetalheCadastramentoInfracao string `json:"detalhe_cadastramento_infracao"`
	DetalheCodInfracao           string `json:"detalhe_cod_infracao"`
	DetalheCodMunEmplacamento    string `json:"detalhe_cod_mun_emplacamento"`
	DetalheCodMunInfracao        string `json:"detalhe_cod_mun_infracao"`
	DetalheDtEmissaoPenalidade   string `json:"detalhe_dt_emissao_penalidade"`
	DetalheDtInfracao            string `json:"detalhe_dt_infracao"`
	DetalheDtNotificacaoInfracao string `json:"detalhe_dt_notificacao_infracao"`
	DetalheHrInfracao            string `json:"detalhe_hr_infracao"`
	DetalheLimitePermitido       string `json:"detalhe_limite_permitido"`
	DetalheLocalInfracao         string `json:"detalhe_local_infracao"`
	DetalheAmrcamodelo           string `json:"detalhe_amrcamodelo"`
	DetalheMedicaoConsiderada    string `json:"detalhe_medicao_considerada"`
	DetalheMedicaoReal           string `json:"detalhe_medicao_real"`
	DetalheNumAutoInfracao       string `json:"detalhe_num_auto_infracao"`
	DetalheOrgaoAutuador         string `json:"detalhe_orgao_autuador"`
	DetalhePlaca                 string `json:"detalhe_placa"`
	DetalheTipoAutoInfracao      string `json:"detalhe_tipo_auto_infracao"`
	DetalheUfJurisdicaoVeiculo   string `json:"detalhe_uf_jurisdicao_veiculo"`
	DetalheUfOrgaoAutuador       string `json:"detalhe_uf_orgao_autuador"`
	DetalheUfPlaca               string `json:"detalhe_uf_placa"`
	DetalheUnidadeMedida         string `json:"detalhe_unidade_medida"`
	DetalheValorInfracao         string `json:"detalhe_valor_infracao"`
	DadosSuspensaoAceiteUfJuris  string `json:"dadosdasuspensao_aceite_uf_jurisdicao"`
	DadosSuspensaoDataRegistro   string `json:"dadosdasuspensao_data_registro"`
	DadosSuspensaoOrigem         string `json:"dadosdasuspensao_origem"`
	DadosSuspensaoTipo           string `json:"dadosdasuspensao_tipo"`
	DadosInfratorCnhCondutor     string `json:"dadosinfrator_cnh_condutor"`
	DadosInfratorCnhInfrator     string `json:"dadosinfrator_cnh_infrator"`
	DadosPagamentoDtPagamento    string `json:"dadosdopagamento_dt_pagamento"`
	DadosPagamentoDtRegistroPgto string `json:"dadosdopagamento_dt_do_registro_do_pgmto"`
	DadosPagamentoUfPagamento    string `json:"dadosdopagamento_uf_pagamento"`
	DadosPagamentoValorPago      string `json:"dadosdopagamento_valor_pago"`
	DadosPagamentoDadosPgmto     string `json:"dadosdopagamento_dados_pgmto"`
}

type MultasResponse struct {
	Data struct {
		Alerta                   string   `json:"alerta"`
		Placa                    string   `json:"placa"`
		QuantidadeOcorrencias    string   `json:"quantidade_ocorrencias"`
		QuantidadeOcorrenciasTot string   `json:"quantidade_ocorrencias_total"`
		Registros                []MultaA `json:"registros"`
	} `json:"data"`
}

func ConsultarMultas(placa string) (MultasResponse, error) {
	placa = strings.ToUpper(strings.TrimSpace(placa))
	placa = strings.ReplaceAll(placa, "-", "") // Remove hífens da placa

	username, password := getBasicAuthCredentials()

	multasURL := fmt.Sprintf("https://api.consultarplaca.com.br/v2/consultarRegistrosInfracoesRenainf?placa=%s", placa)

	req, err := http.NewRequest("GET", multasURL, nil)
	if err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	req.SetBasicAuth(username, password)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao enviar requisição de multas: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao ler resposta de multas: %w", err)
	}

	var infracoesResp consultarInfracoesAPIResponse
	if err := json.Unmarshal(body, &infracoesResp); err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao decodificar JSON de multas: %w", err)
	}

	if infracoesResp.Status != "ok" {
		// Retorna resposta com array vazio de multas
		multaResp := MultasResponse{}
		multaResp.Data.Placa = placa
		multaResp.Data.Registros = []MultaA{}
		return multaResp, nil
	}

	infracoes := infracoesResp.Dados.RegistroDebitosPorInfracoes.InfracoesRenainf.Infracoes
	registros := make([]MultaA, 0, len(infracoes))
	for _, inf := range infracoes {
		registros = append(registros, MultaA{
			NumeroAutoInfracao:           inf.DadosInfracao.NumeroAutoInfracao,
			Infracao:                     inf.DadosInfracao.Infracao,
			DetalheOrgaoAutuador:         inf.DadosInfracao.OrgaoAutuador,
			DetalheValorInfracao:         inf.DadosInfracao.ValorAplicado,
			DetalheTipoAutoInfracao:      inf.DadosInfracao.TipoAutoInfracao,
			DetalheLocalInfracao:         inf.DadosInfracao.LocalInfracao,
			DetalheUnidadeMedida:         inf.Aplicacao.UnidadeMedida,
			DetalheLimitePermitido:       inf.Aplicacao.LimitePermitido,
			DetalheMedicaoConsiderada:    inf.Aplicacao.MedicaoConsiderada,
			DetalheMedicaoReal:           inf.Aplicacao.MedicaoReal,
			DetalheHrInfracao:            inf.Eventos.DataHoraInfracao,
			DetalheCadastramentoInfracao: inf.Eventos.DataCadastramento,
			DetalheDtNotificacaoInfracao: inf.Eventos.DataNotificacao,
			DetalheDtEmissaoPenalidade:   inf.Eventos.DataEmissaoPenalidade,
		})
	}

	var multaResp MultasResponse
	multaResp.Data.Placa = placa
	multaResp.Data.QuantidadeOcorrencias = fmt.Sprintf("%d", len(registros))
	multaResp.Data.QuantidadeOcorrenciasTot = fmt.Sprintf("%d", len(registros))
	multaResp.Data.Registros = registros

	return multaResp, nil
}

// ─── ConsultarMultiplasPlacas ─────────────────────────────────────────────────

// ConsultarMultiplasPlacas consulta múltiplas placas e retorna um mapa com os resultados
func ConsultarMultiplasPlacas(placas []string) (map[string]*FullAPIResponse, error) {

	results := make(map[string]*FullAPIResponse)

	// Processa cada placa em paralelo usando goroutines
	type result struct {
		placa string
		resp  *FullAPIResponse
		err   error
	}

	resultsChan := make(chan result, len(placas))

	// Cria goroutines para cada placa
	for _, placa := range placas {
		go func(p string) {
			resp, err := ConsultarPlaca(p)
			resultsChan <- result{placa: p, resp: resp, err: err}
		}(placa)
	}

	// Coleta os resultados
	for i := 0; i < len(placas); i++ {
		res := <-resultsChan
		if res.err != nil {
			// Cria uma resposta de erro para esta placa
			results[res.placa] = &FullAPIResponse{
				Error:   true,
				Message: res.err.Error(),
				Data:    Response{},
			}
		} else {
			results[res.placa] = res.resp
		}
	}

	return results, nil
}
