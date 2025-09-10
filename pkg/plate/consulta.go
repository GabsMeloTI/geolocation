package plate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// Estrutura para a resposta da API alternativa (veiculos-dados-v1)
type FallbackAPIResponse struct {
	User struct {
		FirstName    string `json:"first_name"`
		Email        string `json:"email"`
		Cellphone    string `json:"cellphone"`
		Notification string `json:"notification"`
	} `json:"user"`
	Balance string `json:"balance"`
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Homolog bool   `json:"homolog"`
	Data    []struct {
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
	} `json:"data"`
}

// Converte a resposta da API alternativa para o formato FullAPIResponse
func convertFallbackToFullResponse(fallbackResp FallbackAPIResponse) *FullAPIResponse {
	// Pega o primeiro item do array (se existir)
	var dataItem struct {
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

	if len(fallbackResp.Data) > 0 {
		dataItem = fallbackResp.Data[0]
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
			Extra: struct {
				AnoFabricacao   string `json:"ano_fabricacao"`
				CapMaximaTracao string `json:"cap_maxima_tracao"`
				Chassi          string `json:"chassi"`
			}{
				AnoFabricacao:   fmt.Sprintf("%d", dataItem.AnoFabricacao),
				CapMaximaTracao: fmt.Sprintf("%d", dataItem.CapacidadeMaxTracao),
				Chassi:          dataItem.Chassi,
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
	startFallback := time.Now()
	fallbackURL := "https://gateway.apibrasil.io/api/v2/vehicles/base/001/consulta"
	fallbackPayload := fmt.Sprintf(`{"tipo":"agregados-basica","placa":"%s","homolog":false}`, placa)

	// Log da requisição de fallback
	log.Printf("🚀 [API BRASIL - FALLBACK] Iniciando consulta alternativa para placa: %s", placa)
	log.Printf("🌐 [API BRASIL - FALLBACK] URL: %s", fallbackURL)
	log.Printf("📤 [API BRASIL - FALLBACK] Request Body: %s", fallbackPayload)
	log.Printf("🔑 [API BRASIL - FALLBACK] Headers - Authorization: Bearer %s", bearer)
	log.Printf("🔑 [API BRASIL - FALLBACK] Headers - DeviceToken: %s", device)

	req, err := http.NewRequest("POST", fallbackURL, strings.NewReader(fallbackPayload))
	if err != nil {
		log.Printf("❌ [API BRASIL - FALLBACK] Erro ao criar requisição: %v", err)
		return nil, fmt.Errorf("erro ao criar requisição da API alternativa: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ [API BRASIL - FALLBACK] Erro ao enviar requisição: %v", err)
		return nil, fmt.Errorf("erro ao enviar requisição da API alternativa: %w", err)
	}
	defer resp.Body.Close()

	// Log da resposta de fallback
	log.Printf("📊 [API BRASIL - FALLBACK] Status Code: %d", resp.StatusCode)
	log.Printf("📊 [API BRASIL - FALLBACK] Response Headers: %v", resp.Header)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [API BRASIL - FALLBACK] Erro ao ler resposta: %v", err)
		return nil, fmt.Errorf("erro ao ler resposta da API alternativa: %w", err)
	}

	// Log da resposta completa
	log.Printf("📄 [API BRASIL - FALLBACK] Response Body: %s", string(respBody))
	log.Printf("⏱️ [API BRASIL - FALLBACK] Tempo de resposta: %v", time.Since(startFallback))

	// Decodifica a resposta da API alternativa
	var fallbackResp FallbackAPIResponse
	if err := json.Unmarshal(respBody, &fallbackResp); err != nil {
		log.Printf("❌ [API BRASIL - FALLBACK] Erro ao decodificar JSON: %v", err)
		return nil, fmt.Errorf("erro ao decodificar JSON da API alternativa: %w", err)
	}

	// Verifica se a resposta indica erro
	if fallbackResp.Error {
		log.Printf("❌ [API BRASIL - FALLBACK] API alternativa retornou erro: %s", fallbackResp.Message)
		return nil, fmt.Errorf("API alternativa retornou erro: %s", fallbackResp.Message)
	}

	// Converte para o formato esperado
	fullResp := convertFallbackToFullResponse(fallbackResp)

	log.Printf("✅ [API BRASIL - FALLBACK] Conversão concluída com sucesso")
	return fullResp, nil
}

func ConsultarPlaca(placa string) (*FullAPIResponse, error) {
	startTotal := time.Now() // ⏱ início da função

	placa = strings.ToUpper(strings.TrimSpace(placa))
	placa = strings.ReplaceAll(placa, "-", "") // Remove hífens da placa
	cacheKey := "placa:" + placa

	// 🔹 Verifica cache Redis apenas para dados da placa (sem multas)
	var fullResp FullAPIResponse
	var dadosPlacaCached bool

	if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		var cachedResp FullAPIResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			fmt.Println("🔁 Cache Redis usado para dados da placa")
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
		startVeiculo := time.Now()
		veiculoURL := "https://gateway.apibrasil.io/api/v2/vehicles/dados"
		body := fmt.Sprintf(`{"placa":"%s", "homolog":%t}`, placa, false)

		// Log da requisição principal
		log.Printf("🚀 [API BRASIL - DADOS VEÍCULO] Iniciando consulta para placa: %s", placa)
		log.Printf("🌐 [API BRASIL - DADOS VEÍCULO] URL: %s", veiculoURL)
		log.Printf("📤 [API BRASIL - DADOS VEÍCULO] Request Body: %s", body)

		req, err := http.NewRequest("POST", veiculoURL, strings.NewReader(body))
		if err != nil {
			log.Printf("❌ [API BRASIL - DADOS VEÍCULO] Erro ao criar requisição: %v", err)
			return nil, fmt.Errorf("erro ao criar requisição do veículo: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+bearer)
		req.Header.Set("DeviceToken", device)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ [API BRASIL - DADOS VEÍCULO] Erro ao enviar requisição: %v", err)
			return nil, fmt.Errorf("erro ao enviar requisição do veículo: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ [API BRASIL - DADOS VEÍCULO] Erro ao ler resposta: %v", err)
			return nil, fmt.Errorf("erro ao ler resposta do veículo: %w", err)
		}

		// Log da resposta principal
		log.Printf("📊 [API BRASIL - DADOS VEÍCULO] Status Code: %d", resp.StatusCode)
		log.Printf("📄 [API BRASIL - DADOS VEÍCULO] Response Body: %s", string(respBody))

		fmt.Println("📄 Resposta bruta da API de dados do veículo:")
		fmt.Println(string(respBody))
		fmt.Printf("⏱ Tempo API veículos: %v\n", time.Since(startVeiculo))

		// Verifica se precisa fazer fallback (status 400)
		if resp.StatusCode == 400 {
			log.Printf("🔄 [API BRASIL - FALLBACK] Status 400 detectado, tentando API alternativa para semi-reboques/carrocerias")

			// Chama a API alternativa
			fallbackResp, fallbackErr := consultarAPIAlternativa(placa, bearer, device, client)
			if fallbackErr != nil {
				log.Printf("❌ [API BRASIL - FALLBACK] Erro na API alternativa: %v", fallbackErr)
				return nil, fmt.Errorf("erro na API alternativa: %w", fallbackErr)
			}

			log.Printf("✅ [API BRASIL - FALLBACK] API alternativa executada com sucesso")
			fullResp = *fallbackResp
		} else {
			// Processa resposta normal
			if err := json.Unmarshal(respBody, &fullResp); err != nil {
				log.Printf("❌ [API BRASIL - DADOS VEÍCULO] Erro ao decodificar JSON: %v", err)
				return nil, fmt.Errorf("erro ao decodificar JSON do veículo: %w", err)
			}
		}

		// Cache apenas os dados da placa (sem multas)
		respBytes, _ := json.Marshal(fullResp)
		if err := rdb.Set(ctx, cacheKey, respBytes, 30*time.Minute).Err(); err != nil {
			fmt.Println("❌ Falha ao salvar dados da placa no Redis:", err)
		}
	}

	// 2. Consulta multas (SEMPRE consulta, nunca usa cache)
	startMultas := time.Now()
	multasURL := "https://gateway.apibrasil.io/api/v2/vehicles/base/001/consulta"
	multaPayload := fmt.Sprintf(`{"placa":"%s", "tipo": "%s"}`, placa, "renainf")

	// Log da requisição de multas
	log.Printf("🚀 [API BRASIL - MULTAS] Iniciando consulta para placa: %s", placa)
	log.Printf("🌐 [API BRASIL - MULTAS] URL: %s", multasURL)
	log.Printf("📤 [API BRASIL - MULTAS] Request Body: %s", multaPayload)

	reqMultas, err := http.NewRequest("POST", multasURL, strings.NewReader(multaPayload))
	if err != nil {
		log.Printf("❌ [API BRASIL - MULTAS] Erro ao criar requisição: %v", err)
		return nil, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	reqMultas.Header.Set("Authorization", "Bearer "+bearer)
	reqMultas.Header.Set("DeviceToken", device)
	reqMultas.Header.Set("Content-Type", "application/json")

	respMultas, err := client.Do(reqMultas)
	if err != nil {
		log.Printf("❌ [API BRASIL - MULTAS] Erro ao enviar requisição: %v", err)
		fmt.Println("⚠️ Erro ao consultar multas:", err)
	} else {
		defer respMultas.Body.Close()

		// Log da resposta de multas
		log.Printf("📊 [API BRASIL - MULTAS] Status Code: %d", respMultas.StatusCode)

		multasRespBody, _ := io.ReadAll(respMultas.Body)

		// Log da resposta completa
		log.Printf("📄 [API BRASIL - MULTAS] Response Body: %s", string(multasRespBody))
		log.Printf("⏱️ [API BRASIL - MULTAS] Tempo de resposta: %v", time.Since(startMultas))

		fmt.Println("📄 Resposta bruta da nova API de multas:")
		fmt.Println(string(multasRespBody))
		fmt.Printf("⏱ Tempo API multas: %v\n", time.Since(startMultas))

		// Verifica se é erro de saldo insuficiente
		if strings.Contains(string(multasRespBody), "Saldo insuficiente") {
			log.Printf("⚠️ [API BRASIL - MULTAS] Saldo insuficiente detectado, retornando array de multas vazio")
			fmt.Println("⚠️ Saldo insuficiente para consulta de multas, retornando array vazio")
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
					log.Printf("📄 [API BRASIL - MULTAS] API retornou erro: %s", multaAPIResponse.Message)
					fmt.Printf("📄 API de multas retornou erro: %s (sem tarifação)\n", multaAPIResponse.Message)
					fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas
				} else {
					fullResp.Data.Multas.Dados = multaAPIResponse.Data.Registros
				}
			} else {
				// Se falhar, tenta decodificar como array vazio (caso raro)
				var dataArray []interface{}
				if err := json.Unmarshal(multasRespBody, &dataArray); err == nil {
					log.Printf("📄 [API BRASIL - MULTAS] API retornou array vazio")
					fmt.Println("📄 API retornou array vazio para multas")
					fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas
				} else {
					log.Printf("⚠️ [API BRASIL - MULTAS] Erro ao decodificar JSON: %v", err)
					fmt.Println("⚠️ Erro ao decodificar JSON da nova API de multas:", err)
					fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas em caso de erro
				}
			}
		}
	}

	fmt.Printf("✅ Tempo total da função ConsultarPlaca: %v\n", time.Since(startTotal))
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
	start := time.Now()
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
		log.Printf("⚠️ [API BRASIL - CONSULTAR MULTAS] Saldo insuficiente detectado, retornando array de multas vazio")
		fmt.Println("⚠️ Saldo insuficiente para consulta de multas, retornando array vazio")

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

		fmt.Printf("✅ Multas consultadas para %s em %v (total %d) - Saldo insuficiente\n",
			placa, time.Since(start), len(multaResp.Data.Registros))

		return multaResp, nil
	}

	// Parse normal
	var multaResp MultasResponse
	if err := json.Unmarshal(body, &multaResp); err != nil {
		return MultasResponse{}, fmt.Errorf("erro ao decodificar JSON de multas: %w", err)
	}

	fmt.Printf("📄 Resposta bruta da API de multas (placa %s):\n%s\n", placa, multaResp)

	fmt.Printf("✅ Multas consultadas para %s em %v (total %d)\n",
		placa, time.Since(start), len(multaResp.Data.Registros))

	return multaResp, nil
}

// ConsultarMultiplasPlacas consulta múltiplas placas e retorna um mapa com os resultados
func ConsultarMultiplasPlacas(placas []string) (map[string]*FullAPIResponse, error) {
	startTotal := time.Now()
	fmt.Printf("🚀 Iniciando consulta de %d placas\n", len(placas))

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
			fmt.Printf("🔍 Consultando placa: %s\n", p)
			resp, err := ConsultarPlaca(p)
			resultsChan <- result{placa: p, resp: resp, err: err}
		}(placa)
	}

	// Coleta os resultados
	for i := 0; i < len(placas); i++ {
		res := <-resultsChan
		if res.err != nil {
			fmt.Printf("❌ Erro ao consultar placa %s: %v\n", res.placa, res.err)
			// Cria uma resposta de erro para esta placa
			results[res.placa] = &FullAPIResponse{
				Error:   true,
				Message: res.err.Error(),
				Data:    Response{},
			}
		} else {
			fmt.Printf("✅ Placa %s consultada com sucesso\n", res.placa)
			results[res.placa] = res.resp
		}
	}

	fmt.Printf("🏁 Consulta de %d placas finalizada em %v\n", len(placas), time.Since(startTotal))
	return results, nil
}
