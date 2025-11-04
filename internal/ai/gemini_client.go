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

// GeminiClient handles interactions with Google Gemini API
type GeminiClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// GeminiRequest represents a request to Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents content in Gemini format
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse represents Gemini API response
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

// GeminiCandidate represents a response candidate
type GeminiCandidate struct {
	Content       GeminiContent  `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

// SafetyRating represents safety assessment
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiError represents an error from Gemini API
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateContent sends a content generation request to Gemini
func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string, systemPrompt string) (*GeminiResponse, error) {
	// Combine system prompt and user prompt
	fullPrompt := fmt.Sprintf("%s\n\nUser Request: %s", systemPrompt, prompt)

	req := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: fullPrompt,
					},
				},
			},
		},
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 2000,
			TopP:            0.8,
			TopK:            40,
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use Gemini Pro model
	url := fmt.Sprintf("%s/models/gemini-pro:generateContent?key=%s", c.baseURL, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	return &geminiResp, nil
}

// GenerateEmbedding creates embeddings using Gemini API (basic implementation)
func (c *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Note: Gemini doesn't have a direct embedding API like OpenAI
	// This is a mock implementation that returns a basic embedding
	// In production, you might use Google's embedding models separately

	embedding := make([]float64, 1536) // Match OpenAI dimensions for compatibility
	textLen := float64(len(text))

	// Create a simple but consistent embedding based on text characteristics
	for i := range embedding {
		embedding[i] = (textLen / float64(i+1)) * 0.001
		if i%2 == 0 {
			embedding[i] *= -1 // Add some variation
		}
	}

	return embedding, nil
}

// BuildGeminiNetworkAnalysisPrompt creates a system prompt optimized for Gemini
func BuildGeminiNetworkAnalysisPrompt() string {
	return `You are an expert network analysis AI for NetOrchestrator, a network orchestration platform.

CONTEXT: You analyze network performance data and provide actionable insights for network infrastructure.

CAPABILITIES:
- Identify performance bottlenecks and anomalies
- Recommend optimization strategies  
- Detect security vulnerabilities
- Predict capacity needs
- Suggest cost optimizations

SUPPORTED COMPONENTS:
- Node types: router, switch, host, firewall, load_balancer, gateway
- Topologies: star, mesh, tree, custom
- Policies: traffic, security, firewall, qos, routing, access
- Vendors: cisco, juniper, arista

OUTPUT FORMAT:
Provide clear, technical analysis with:
1. Issue identification and severity assessment
2. Specific actionable recommendations
3. Expected impact and improvement metrics
4. Risk assessment (low/medium/high)

Keep responses concise, technical, and actionable.`
}

// BuildGeminiNLPValidationPrompt creates a prompt for NLP validation optimized for Gemini
func BuildGeminiNLPValidationPrompt() string {
	return `You are a network configuration validator for NetOrchestrator. Parse natural language network requests and extract valid configurations.

STRICTLY SUPPORTED COMPONENTS:
- Node Types: router, switch, host, firewall, load_balancer, gateway
- Topologies: star, mesh, tree, custom
- Policy Types: traffic, security, firewall, qos, routing, access  
- Firewall Actions: allow, deny, drop
- Device Vendors: cisco, juniper, arista

VALIDATION RULES:
1. Extract only supported node types, topologies, and policies
2. Reject requests for unsupported components
3. Validate topology constraints (mesh needs 3+ nodes, star needs 2+)
4. Ensure realistic network configurations

RESPONSE FORMAT - Return valid JSON only:
{
  "valid": true/false,
  "network_name": "extracted name",
  "topology": "star|mesh|tree|custom",
  "nodes": [{"name": "node1", "type": "router"}, ...],
  "policies": [{"type": "firewall", "rules": [...]}],
  "validation_errors": ["error1", "error2"],
  "confidence": 0.0-1.0
}

If invalid, explain why in validation_errors array. Only return the JSON object.`
}
