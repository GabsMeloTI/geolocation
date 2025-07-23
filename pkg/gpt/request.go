package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Messages   []Message   `json:"messages"`
	Model      string      `json:"model"`
	Tools      []string    `json:"tools"`
	ToolChoice interface{} `json:"tool_choice"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ResponseBody struct {
	Choices []Choice `json:"choices"`
}

// CallGPTAgent envia uma mensagem para o agente personalizado e retorna a resposta.
func CallGPTAgent(assistantID string, userMessage string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY não definido")
	}

	body := RequestBody{
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
		Model:      assistantID,
		Tools:      []string{},
		ToolChoice: nil,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response ResponseBody
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta retornada pelo agente")
	}

	return response.Choices[0].Message.Content, nil
}

type ChatRequest struct {
	Model      string      `json:"model"`
	Messages   []Message   `json:"messages"`
	ToolChoice interface{} `json:"tool_choice,omitempty"`
	Tools      []string    `json:"tools,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func Agent(caminhao string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY não encontrado")
	}

	assistantID := "g-688125bf3d8c8191914b360d197d0bb7"

	client := resty.New()

	message := fmt.Sprintf("Por favor, calcule o consumo de gasolina e a emissão de carbono para um caminhão modelo %s.", caminhao)
	requestBody := ChatRequest{
		Model: assistantID,
		Messages: []Message{
			{Role: "user", Content: message},
		},
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		SetResult(&ChatResponse{}).
		Post("https://api.openai.com/v1/chat/completions")

	if err != nil {
		return "", fmt.Errorf("erro na requisição: %w", err)
	}

	result := resp.Result().(*ChatResponse)
	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta foi retornada pela API")
	}

	return result.Choices[0].Message.Content, nil
}

func PerguntarAoGptEstruturado(modelo string) (map[string]float64, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY não encontrado")
	}

	client := resty.New()

	// Prompt dinâmico com instruções claras
	systemPrompt := fmt.Sprintf(`Você é um especialista técnico em caminhões pesados. O combustível utilizado é diesel. Sempre responda com um JSON no formato:
{
  "combustivel_gasto_por_km": float,
  "carbono_emitido_por_km": float
}

Baseie-se em dados médios reais do caminhão %s. Estime o consumo em km/l (por exemplo, entre 2.8 e 3.2 km/l), depois calcule o valor de combustivel_gasto_por_km como 1 / km_por_litro. Em seguida, calcule carbono_emitido_por_km = combustivel_gasto_por_km * 2.63. Não forneça explicações, comentários ou texto adicional. Responda apenas com o JSON.`, modelo)

	userPrompt := fmt.Sprintf("Forneça os dados de consumo e emissão para o caminhão modelo %s.", modelo)

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

	var resultado map[string]float64
	if err := json.Unmarshal([]byte(content), &resultado); err != nil {
		return nil, fmt.Errorf("erro ao converter resposta em JSON estruturado: %s", content)
	}

	return resultado, nil
}
