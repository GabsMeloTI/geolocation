package plate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "3.238.112.146:6379",
		Password: "",
		DB:       0,
	})
}

func ConsultarPlaca(placa string) (*FullAPIResponse, error) {
	startTotal := time.Now() // ⏱ início da função

	placa = strings.ToUpper(strings.TrimSpace(placa))
	cacheKey := "placa:" + placa

	// 🔹 Verifica cache Redis
	if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		var cachedResp FullAPIResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			fmt.Println("🔁 Cache Redis usado")
			fmt.Printf("⏱ Tempo total (CACHE): %v\n", time.Since(startTotal))
			return &cachedResp, nil
		}
	}

	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN")

	client := &http.Client{
		Timeout: 20 * time.Second, // evita ficar travado muito tempo
	}

	// 1. Consulta dados do veículo
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

	var fullResp FullAPIResponse
	if err := json.Unmarshal(respBody, &fullResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON do veículo: %w", err)
	}

	// 2. Consulta multas
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

		var multaAPIResponse struct {
			Data struct {
				Registros []Multa `json:"registros"`
			} `json:"data"`
		}

		if err := json.Unmarshal(multasRespBody, &multaAPIResponse); err == nil {
			fullResp.Data.Multas.Dados = multaAPIResponse.Data.Registros
		} else {
			fmt.Println("⚠️ Erro ao decodificar JSON da nova API de multas:", err)
		}
	}

	// 3. Cache final com tudo
	respBytes, _ := json.Marshal(fullResp)
	if err := rdb.Set(ctx, cacheKey, respBytes, 30*time.Minute).Err(); err != nil {
		fmt.Println("❌ Falha ao salvar no Redis:", err)
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

func ConsultarMultas(placa string) ([]MultaA, error) {
	start := time.Now()
	placa = strings.ToUpper(strings.TrimSpace(placa))

	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN")

	multasURL := "https://gateway.apibrasil.io/api/v2/vehicles/base/001/consulta"
	payload := []byte(fmt.Sprintf(`{"placa":"%s", "tipo":"renainf"}`, placa))

	req, err := http.NewRequest("POST", multasURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de multas: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second} // evita travar
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição de multas: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta de multas: %w", err)
	}

	// Debug opcional
	fmt.Printf("📄 Resposta bruta da API de multas (placa %s):\n%s\n", placa, string(body))

	// Parse
	var multaResp MultasResponse
	if err := json.Unmarshal(body, &multaResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON de multas: %w", err)
	}

	fmt.Printf("✅ Multas consultadas para %s em %v (total %d)\n",
		placa, time.Since(start), len(multaResp.Data.Registros))

	return multaResp.Data.Registros, nil
}
