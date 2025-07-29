package plate

import (
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
	placa = strings.ToUpper(strings.TrimSpace(placa))
	cacheKey := "placa:" + placa

	if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		var cachedResp FullAPIResponse
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			fmt.Println("🔁 Cache Redis usado")
			return &cachedResp, nil
		}
	}

	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN")

	// 1. Consulta dados do veículo
	veiculoURL := "https://gateway.apibrasil.io/api/v2/vehicles/dados"
	body := fmt.Sprintf(`{"placa":"%s", "homolog":%t}`, placa, false)

	req, err := http.NewRequest("POST", veiculoURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição do veículo: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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
	fmt.Println(string(respBody)) // 🚨 Log do retorno bruto

	var fullResp FullAPIResponse
	if err := json.Unmarshal(respBody, &fullResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON do veículo: %w", err)
	}

	// 2. Consulta multas (nova API)
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

	return &fullResp, nil
}
