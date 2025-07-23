package pkg

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

	return result.Choices[0].Message.Content, nil
}
