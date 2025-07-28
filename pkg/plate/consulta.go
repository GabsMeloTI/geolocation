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

	url := "https://gateway.apibrasil.io/api/v2/vehicles/base/000/dados"
	body := fmt.Sprintf(`{"placa":"%s", "homolog":%t}`, placa, false)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var fullResp FullAPIResponse
	if err := json.Unmarshal(respBody, &fullResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}

	respBytes, _ := json.Marshal(fullResp)
	if err := rdb.Set(ctx, cacheKey, respBytes, 30*time.Minute).Err(); err != nil {
		fmt.Println("❌ Falha ao salvar no Redis:", err)
	}

	return &fullResp, nil
}
