package gpt

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequisicaoCaminhao struct {
	Modelo    string             `json:"modelo"`
	PesoTotal float64            `json:"pesoTotal,omitempty"`
	Fatores   map[string]float64 `json:"fatores,omitempty"`
}

func PerguntarAoGptEstruturado(data RequisicaoCaminhao) (map[string]float64, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY não encontrado")
	}

	client := resty.New()

	// Prompt dinâmico: obter dados médios do modelo
	systemPrompt := fmt.Sprintf(`Você é um especialista técnico em caminhões pesados. O combustível utilizado é diesel. 
Sempre responda com um JSON no formato:
{
  "km_por_litro": float,
  "combustivel_gasto_por_km": float,
  "carbono_emitido_por_km": float
}

Baseie-se em dados médios reais do caminhão %s. 
Estime o consumo em km/l (exemplo: entre 2.8 e 3.2 km/l). 
Depois calcule:
- combustivel_gasto_por_km = 1 / km_por_litro
- carbono_emitido_por_km = combustivel_gasto_por_km * 2.63
Não forneça explicações, apenas o JSON.`, data.Modelo)

	userPrompt := fmt.Sprintf("Forneça os dados médios de consumo e emissão para o caminhão modelo %s.", data.Modelo)

	requestBody := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"temperature": 0.3,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post("https://api.openai.com/v1/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, fmt.Errorf("falha ao fazer unmarshal da resposta: %w", err)
	}

	choices, ok := raw["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("resposta inválida da API")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("mensagem inválida da API")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("conteúdo da resposta não é string")
	}

	// Base do GPT
	var resultado map[string]float64
	if err := json.Unmarshal([]byte(content), &resultado); err != nil {
		return nil, fmt.Errorf("erro ao converter resposta em JSON estruturado: %s", content)
	}

	// ---- Valores padrão se não vier pesoTotal ou fatores ----
	if data.PesoTotal <= 0 {
		data.PesoTotal = 20000 // valor médio padrão em kg
	}
	if data.Fatores == nil || len(data.Fatores) == 0 {
		data.Fatores = map[string]float64{
			"serra_pesada":   1.25,
			"urbano_intenso": 1.30,
			"chuva_forte":    1.15,
		}
	}

	// ---- Ajustes com peso e fatores ambientais ----
	// Exemplo de coeficiente base: cada 10.000 kg = +5% de consumo
	coefPeso := 1.0 + ((data.PesoTotal / 10000.0) * 0.05)

	// Produto dos fatores ambientais enviados
	coefAmbiental := 1.0
	for _, v := range data.Fatores {
		coefAmbiental *= v
	}

	// Cálculo final ajustado
	ajustado := resultado["carbono_emitido_por_km"] * coefPeso * coefAmbiental
	resultado["coeficiente_peso"] = coefPeso
	resultado["fator_ambiental_total"] = coefAmbiental
	resultado["carbono_emitido_ajustado"] = ajustado

	return resultado, nil
}
