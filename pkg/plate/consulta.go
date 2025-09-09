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

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "3.238.87.0:6379",
		Password: "",
		DB:       0,
	})
}

func ConsultarPlaca(placa string) (*FullAPIResponse, error) {
	startTotal := time.Now() // ⏱ início da função

	placa = strings.ToUpper(strings.TrimSpace(placa))
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

		fmt.Println("📄 Resposta bruta da API de dados do veículo:")
		fmt.Println(string(respBody))
		fmt.Printf("⏱ Tempo API veículos: %v\n", time.Since(startVeiculo))

		if err := json.Unmarshal(respBody, &fullResp); err != nil {
			return nil, fmt.Errorf("erro ao decodificar JSON do veículo: %w", err)
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

	reqMultas, err := http.NewRequest("POST", multasURL, strings.NewReader(multaPayload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	reqMultas.Header.Set("Authorization", "Bearer "+bearer)
	reqMultas.Header.Set("DeviceToken", device)
	reqMultas.Header.Set("Content-Type", "application/json")

	respMultas, err := client.Do(reqMultas)
	if err != nil {
		fmt.Println("⚠️ Erro ao consultar multas:", err)
	} else {
		defer respMultas.Body.Close()
		multasRespBody, _ := io.ReadAll(respMultas.Body)

		fmt.Println("📄 Resposta bruta da nova API de multas:")
		fmt.Println(string(multasRespBody))
		fmt.Printf("⏱ Tempo API multas: %v\n", time.Since(startMultas))

		// Tenta decodificar como objeto primeiro
		var multaAPIResponse struct {
			Data struct {
				Registros []Multa `json:"registros"`
			} `json:"data"`
		}

		if err := json.Unmarshal(multasRespBody, &multaAPIResponse); err == nil {
			fullResp.Data.Multas.Dados = multaAPIResponse.Data.Registros
		} else {
			// Se falhar, tenta decodificar como array vazio
			var dataArray []interface{}
			if err := json.Unmarshal(multasRespBody, &dataArray); err == nil {
				fmt.Println("📄 API retornou array vazio para multas")
				fullResp.Data.Multas.Dados = []Multa{} // Array vazio de multas
			} else {
				fmt.Println("⚠️ Erro ao decodificar JSON da nova API de multas:", err)
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

	// Debug opcional

	// Parse
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
