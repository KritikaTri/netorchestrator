package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient handles interactions with OpenAI API
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// OpenAIRequest represents a request to OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// OpenAIResponse represents OpenAI API response
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIChoice represents a choice in OpenAI response
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage represents token usage information
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an error from OpenAI API
type OpenAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Code    interface{} `json:"code"`
}

// EmbeddingRequest represents a request to OpenAI embeddings API
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents OpenAI embeddings API response
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  OpenAIUsage     `json:"usage"`
	Error  *OpenAIError    `json:"error,omitempty"`
}

// EmbeddingData represents embedding data
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion sends a chat completion request to OpenAI
func (c *OpenAIClient) ChatCompletion(ctx context.Context, req OpenAIRequest) (*OpenAIResponse, error) {
	// Set default model if not specified
	if req.Model == "" {
		req.Model = "gpt-4"
	}

	// Set default temperature if not specified
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	// Set default max tokens if not specified
	if req.MaxTokens == 0 {
		req.MaxTokens = 2000
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", openaiResp.Error.Message, openaiResp.Error.Type)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	return &openaiResp, nil
}

// CreateEmbedding creates embeddings using OpenAI API
func (c *OpenAIClient) CreateEmbedding(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error) {
	if model == "" {
		model = "text-embedding-ada-002"
	}

	req := EmbeddingRequest{
		Model: model,
		Input: texts,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if embeddingResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", embeddingResp.Error.Message, embeddingResp.Error.Type)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	return &embeddingResp, nil
}

// BuildNetworkAnalysisPrompt creates a system prompt for network analysis
func BuildNetworkAnalysisPrompt() string {
	return `You are an AI network analysis expert for NetOrchestrator. Analyze network performance data and provide insights.

Your capabilities include:
- Identifying performance bottlenecks and anomalies
- Recommending optimization strategies
- Detecting security vulnerabilities
- Predicting capacity needs
- Suggesting cost optimizations

Supported network components:
- Node types: router, switch, host, firewall, load_balancer, gateway
- Topologies: star, mesh, tree, custom
- Policies: traffic, security, firewall, qos, routing, access
- Vendors: cisco, juniper, arista

Always provide:
1. Clear analysis of the issue
2. Specific actionable recommendations  
3. Risk assessment (low/medium/high)
4. Expected impact/improvement

Keep responses concise and technical.`
}

// BuildNLPValidationPrompt creates a prompt for validating NLP requests
func BuildNLPValidationPrompt() string {
	return `You are a network configuration validator for NetOrchestrator. Parse natural language network requests and extract valid configurations.

STRICTLY SUPPORTED COMPONENTS:
- Node Types: router, switch, host, firewall, load_balancer, gateway
- Topologies: star, mesh, tree, custom  
- Policy Types: traffic, security, firewall, qos, routing, access
- Firewall Actions: allow, deny, drop
- Device Vendors: cisco, juniper, arista

VALIDATION RULES:
1. Only extract supported node types, topologies, and policies
2. Reject requests for unsupported components
3. Validate topology constraints (e.g., mesh needs 3+ nodes)
4. Ensure realistic network configurations

Return JSON format:
{
  "valid": true/false,
  "network_name": "extracted name",
  "topology": "star|mesh|tree|custom", 
  "nodes": [{"name": "node1", "type": "router"}, ...],
  "policies": [{"type": "firewall", "rules": [...]}],
  "validation_errors": ["error1", "error2"],
  "confidence": 0.0-1.0
}

If invalid, explain why in validation_errors array.`
}
