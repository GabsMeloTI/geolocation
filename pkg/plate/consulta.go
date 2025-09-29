package plate

import (
	"bytes"
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

// Estrutura para os dados do veículo
type VeiculoData struct {
	Placa               string `json:"placa"`
	Chassi              string `json:"chassi"`
	Fabricante          string `json:"fabricante"`
	Modelo              string `json:"modelo"`
	AnoFabricacao       int    `json:"ano_fabricacao"`
	AnoModelo           int    `json:"ano_modelo"`
	Combustivel         string `json:"combustivel"`
	TipoVeiculo         string `json:"tipo_veiculo"`
	Especie             string `json:"especie"`
	Cor                 string `json:"cor"`
	TipoCarroceria      string `json:"tipo_carroceria"`
	Nacionalidade       string `json:"nacionalidade"`
	NumeroMotor         string `json:"numero_motor"`
	Potencia            int    `json:"potencia"`
	Carga               *int   `json:"carga"`
	NumeroCarroceria    *int   `json:"numero_carroceria"`
	NumeroCaixaCambio   *int   `json:"numero_caixa_cambio"`
	NumeroEixoTraseiro  *int   `json:"numero_eixo_traseiro"`
	NumeroTerceiroEixo  *int   `json:"numero_terceiro_eixo"`
	QuantidadeEixo      int    `json:"quantidade_eixo"`
	Cilindradas         string `json:"cilindradas"`
	CapacidadeMaxTracao int    `json:"capacidade_max_tracao"`
	PesoBrutoTotal      int    `json:"peso_bruto_total"`
	QuantidadeLugares   int    `json:"quantidade_lugares"`
	TipoMontagem        *int   `json:"tipo_montagem"`
	UfJurisdicao        string `json:"uf_jurisdicao"`
	UfFaturado          string `json:"uf_faturado"`
	Cidade              string `json:"cidade"`
}

// Estrutura para a resposta da API alternativa (veiculos-dados-v1)
type FallbackAPIResponse struct {
	User struct {
		FirstName    string `json:"first_name"`
		Email        string `json:"email"`
		Cellphone    string `json:"cellphone"`
		Notification string `json:"notification"`
	} `json:"user"`
	Balance string          `json:"balance"`
	Error   bool            `json:"error"`
	Message string          `json:"message"`
	Homolog bool            `json:"homolog"`
	Data    json.RawMessage `json:"data"` // Campo flexível para receber array ou objeto
}

// Converte a resposta da API alternativa para o formato FullAPIResponse
func convertFallbackToFullResponse(fallbackResp FallbackAPIResponse) *FullAPIResponse {
	var dataItem VeiculoData

	// Tenta decodificar como array primeiro
	var dataArray []VeiculoData
	if err := json.Unmarshal(fallbackResp.Data, &dataArray); err == nil && len(dataArray) > 0 {
		// Se for um array, pega o primeiro item
		dataItem = dataArray[0]
	} else {
		// Se não for array, tenta decodificar como objeto único
		if err := json.Unmarshal(fallbackResp.Data, &dataItem); err != nil {
			// Se falhar, cria um objeto vazio
			dataItem = VeiculoData{}
		}
	}

	return &FullAPIResponse{
		Error:   fallbackResp.Error,
		Message: fallbackResp.Message,
		Data: Response{
			Placa:           dataItem.Placa,
			Chassi:          dataItem.Chassi,
			Modelo:          dataItem.Modelo,
			Marca:           dataItem.Fabricante,
			Ano:             fmt.Sprintf("%d", dataItem.AnoFabricacao),
			AnoModelo:       fmt.Sprintf("%d", dataItem.AnoModelo),
			Cor:             dataItem.Cor,
			Uf:              dataItem.UfJurisdicao,
			UfPlaca:         dataItem.UfFaturado,
			Municipio:       dataItem.Cidade,
			Combustivel:     dataItem.Combustivel,
			Potencia:        fmt.Sprintf("%d", dataItem.Potencia),
			CapacidadeCarga: getIntPointerValue(dataItem.Carga),
			Nacionalidade: struct {
				Nacionalidade string `json:"nacionalidade"`
			}{
				Nacionalidade: dataItem.Nacionalidade,
			},
			TipoVeiculo: struct {
				TipoVeiculo string `json:"tipo_veiculo"`
			}{
				TipoVeiculo: dataItem.TipoVeiculo,
			},
			Eixos: fmt.Sprintf("%d", dataItem.QuantidadeEixo),
			Extra: map[string]interface{}{
				"ano_fabricacao":    fmt.Sprintf("%d", dataItem.AnoFabricacao),
				"cap_maxima_tracao": fmt.Sprintf("%d", dataItem.CapacidadeMaxTracao),
				"chassi":            dataItem.Chassi,
			},
			Multas: struct {
				Dados []Multa `json:"dados"`
			}{
				Dados: []Multa{}, // API alternativa não retorna multas
			},
		},
	}
}

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

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "3.238.87.0:6379",
		Password: "",
		DB:       0,
	})
}

// consultarAPIAlternativa consulta a API alternativa para veiculos-dados-v1
func consultarAPIAlternativa(placa, bearer, device string, client *http.Client) (*FullAPIResponse, error) {
	fallbackURL := "https://gateway.apibrasil.io/api/v2/vehicles/base/001/consulta"
	fallbackPayload := fmt.Sprintf(`{"tipo":"agregados-basica","placa":"%s","homolog":false}`, placa)

	req, err := http.NewRequest("POST", fallbackURL, strings.NewReader(fallbackPayload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição da API alternativa: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição da API alternativa: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da API alternativa: %w", err)
	}

	// Decodifica a resposta da API alternativa
	var fallbackResp FallbackAPIResponse
	if err := json.Unmarshal(respBody, &fallbackResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON da API alternativa: %w", err)
	}

	// Verifica se a resposta indica erro
	if fallbackResp.Error {
		return nil, fmt.Errorf("API alternativa retornou erro: %s", fallbackResp.Message)
	}

	// Converte para o formato esperado
	fullResp := convertFallbackToFullResponse(fallbackResp)

	return fullResp, nil
}

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

	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN")

	client := &http.Client{
		Timeout: 60 * time.Second, // evita ficar travado muito tempo
	}

	// 1. Consulta dados do veículo (apenas se não estiver em cache)
	if !dadosPlacaCached {
		veiculoURL := "https://gateway.apibrasil.io/api/v2/vehicles/dados"
		body := fmt.Sprintf(`{"placa":"%s", "homolog":%t}`, placa, false)

		req, err := http.NewRequest("POST", veiculoURL, strings.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("erro ao criar requisição do veículo: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+bearer)
		req.Header.Set("DeviceToken", device)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erro ao enviar requisição do veículo: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta do veículo: %w", err)
		}

		// Verifica se precisa fazer fallback (status 400)
		if resp.StatusCode == 400 {

			// Chama a API alternativa
			fallbackResp, fallbackErr := consultarAPIAlternativa(placa, bearer, device, client)
			if fallbackErr != nil {
				return nil, fmt.Errorf("erro na API alternativa: %w", fallbackErr)
			}

			fullResp = *fallbackResp
		} else {
			// Processa resposta normal
			if err := json.Unmarshal(respBody, &fullResp); err != nil {
				return nil, fmt.Errorf("erro ao decodificar JSON do veículo: %w", err)
			}
		}

		// Cache apenas os dados da placa (sem multas)
		respBytes, _ := json.Marshal(fullResp)
		if err := rdb.Set(ctx, cacheKey, respBytes, 30*time.Minute).Err(); err != nil {
		}
	}

	// 2. Consulta multas (SEMPRE consulta, nunca usa cache)
	multasURL := "https://gateway.apibrasil.io/api/v2/vehicles/base/001/consulta"
	multaPayload := fmt.Sprintf(`{"placa":"%s", "tipo": "%s"}`, placa, "renainf")

	reqMultas, err := http.NewRequest("POST", multasURL, strings.NewReader(multaPayload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	reqMultas.Header.Set("Authorization", "Bearer "+bearer)
	reqMultas.Header.Set("DeviceToken", device)
	reqMultas.Header.Set("Content-Type", "application/json")

	respMultas, err := client.Do(reqMultas)
	if err != nil {
		fmt.Printf("erro ao buscar multas da placa: %s", placa)
		// Erro ao consultar multas - continua sem multas
	} else {
		defer respMultas.Body.Close()

		multasRespBody, _ := io.ReadAll(respMultas.Body)

		fmt.Printf("[DEBUG] Resposta API Brasil (multas) - Placa %s: %s\n------------------------------\n", placa, string(multasRespBody))

		// Verifica se é erro de saldo insuficiente
		if strings.Contains(string(multasRespBody), "Saldo insuficiente") {
			fmt.Printf("Saldo insuficiente")

			// Saldo insuficiente - retorna array vazio
			fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas
		} else {
			// Tenta decodificar como objeto com estrutura normal
			var multaAPIResponse struct {
				Error   bool   `json:"error"`
				Message string `json:"message"`
				Data    struct {
					Registros []Multa `json:"registros"`
				} `json:"data"`
			}

			if err := json.Unmarshal(multasRespBody, &multaAPIResponse); err == nil {
				if multaAPIResponse.Error {
					fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas
				} else {
					fullResp.Data.Multas.Dados = multaAPIResponse.Data.Registros
				}
			} else {
				fmt.Printf("[ERROR] Falha ao decodificar multas JSON (placa %s): %v\n", placa, err)
				// Se falhar, tenta decodificar como array vazio (caso raro)
				var dataArray []interface{}
				if err := json.Unmarshal(multasRespBody, &dataArray); err == nil {
					fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas
				} else {
					// Erro ao decodificar JSON - retorna array vazio
					fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas em caso de erro
				}
			}
		}
	}

	return &fullResp, nil
}

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

	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN")

	multasURL := "https://gateway.apibrasil.io/api/v2/vehicles/base/001/consulta"
	payload := []byte(fmt.Sprintf(`{"placa":"%s", "tipo":"renainf"}`, placa))

	req, err := http.NewRequest("POST", multasURL, bytes.NewBuffer(payload))
	if err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second} // evita travar
	resp, err := client.Do(req)
	if err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao enviar requisição de multas: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao ler resposta de multas: %w", err)
	}

	// Verifica se é erro de saldo insuficiente
	if strings.Contains(string(body), "Saldo insuficiente") {
		// Saldo insuficiente - retorna array vazio

		// Retorna resposta com array vazio de multas
		multaResp := MultasResponse{
			Data: struct {
				Alerta                   string   `json:"alerta"`
				Placa                    string   `json:"placa"`
				QuantidadeOcorrencias    string   `json:"quantidade_ocorrencias"`
				QuantidadeOcorrenciasTot string   `json:"quantidade_ocorrencias_total"`
				Registros                []MultaA `json:"registros"`
			}{
				Placa:     placa,
				Registros: []MultaA{}, // Array vazio
			},
		}

		return multaResp, nil
	}

	// Parse normal
	var multaResp MultasResponse
	if err := json.Unmarshal(body, &multaResp); err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao decodificar JSON de multas: %w", err)
	}

	return multaResp, nil
}

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
