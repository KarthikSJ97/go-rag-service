package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	EmbModel   string
	LLMModel   string
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingResponse struct {
    Embeddings [][]float32 `json:"embeddings"`
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func NewOllamaClient(baseURL string, embModel string, llmModel string) *OllamaClient {
	return &OllamaClient{
		baseURL:  baseURL,
		EmbModel: embModel,
		LLMModel: llmModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *OllamaClient) GetEmbedding(text string) ([]float32, error) {
    reqBody, err := json.Marshal(EmbeddingRequest{
        Model: c.EmbModel,
        Input: text,
    })
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Post(c.baseURL+"/api/embed", "application/json", bytes.NewBuffer(reqBody))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("ollama returned non-200 code: %d", resp.StatusCode)
    }

    var result EmbeddingResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if len(result.Embeddings) == 0 {
        return nil, fmt.Errorf("no embeddings returned from ollama")
    }

    return result.Embeddings[0], nil
}

func (c *OllamaClient) GenerateAnswer(prompt string) (string, error) {
	reqBody, err := json.Marshal(GenerateRequest{
		Model:  c.LLMModel,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned non-200 code: %d", resp.StatusCode)
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Response, nil
}