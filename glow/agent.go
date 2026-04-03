// Package glow implements the Glow chat agent — an ERP data analyst powered by
// Claude with live tool access to supplier, incident, and insights data.
package glow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	anthropicAPI   = "https://api.anthropic.com/v1/messages"
	anthropicModel = "claude-sonnet-4-6"
	maxTokens      = 4096
	maxLoops       = 8 // safety limit on tool-use rounds
)

var systemPrompt = `You are Glow, an intelligent ERP procurement data analyst. You help teams analyse supplier data, detect data incidents, and draft professional responses.

You have live access to the ERP database through your tools. Always call a tool before answering data questions — never guess numbers.

Behaviour guidelines:
- Be concise and professional. Lead with the key finding, then the detail.
- When presenting lists or tables, use markdown formatting (**, bullet points, tables).
- Always suggest a concrete next action at the end of your response.
- If multiple incidents exist, prioritise Critical and High severity ones.
- When drafting questionnaire responses, present each section clearly with the heading bolded.`

// ChatMessage is a single turn in the conversation.
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string (user) or []contentBlock (assistant)
}

// ChatRequest is the payload sent to /api/glow/chat.
type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse is returned to the frontend.
type ChatResponse struct {
	Content   string           `json:"content"`
	ToolUses  []ToolUseDisplay `json:"tool_uses"`
}

// contentBlock is a single block within an assistant or tool-result message.
type contentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

type anthropicRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []ChatMessage   `json:"messages"`
	Tools     []ToolDefinition `json:"tools"`
}

type anthropicResponse struct {
	ID         string         `json:"id"`
	StopReason string         `json:"stop_reason"`
	Content    []contentBlock `json:"content"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Run executes the agentic loop: calls Claude, handles tool use, and returns
// the final text response plus a list of tools that were invoked.
func Run(ctx context.Context, userMessages []ChatMessage) (ChatResponse, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return ChatResponse{}, fmt.Errorf("ANTHROPIC_API_KEY is not set — please export it before starting the server")
	}

	messages := make([]ChatMessage, len(userMessages))
	copy(messages, userMessages)

	var allToolUses []ToolUseDisplay

	for loop := 0; loop < maxLoops; loop++ {
		resp, err := callClaude(ctx, apiKey, messages)
		if err != nil {
			return ChatResponse{}, err
		}

		switch resp.StopReason {

		case "end_turn":
			// Extract final text
			for _, block := range resp.Content {
				if block.Type == "text" {
					return ChatResponse{
						Content:  block.Text,
						ToolUses: allToolUses,
					}, nil
				}
			}
			return ChatResponse{Content: "(no response)", ToolUses: allToolUses}, nil

		case "tool_use":
			// Add assistant message with all content blocks
			messages = append(messages, ChatMessage{
				Role:    "assistant",
				Content: resp.Content,
			})

			// Execute every tool_use block and collect results
			var resultBlocks []contentBlock
			for _, block := range resp.Content {
				if block.Type != "tool_use" {
					continue
				}
				result, display, err := executeTool(block.Name, block.Input)
				if err != nil {
					resultBlocks = append(resultBlocks, contentBlock{
						Type:      "tool_result",
						ToolUseID: block.ID,
						Content:   fmt.Sprintf("Error: %v", err),
						IsError:   true,
					})
				} else {
					resultBlocks = append(resultBlocks, contentBlock{
						Type:      "tool_result",
						ToolUseID: block.ID,
						Content:   result,
					})
					allToolUses = append(allToolUses, ToolUseDisplay{
						Name:    block.Name,
						Display: display,
					})
				}
			}

			// Add user message with tool results
			messages = append(messages, ChatMessage{
				Role:    "user",
				Content: resultBlocks,
			})

		default:
			return ChatResponse{}, fmt.Errorf("unexpected stop_reason: %s", resp.StopReason)
		}
	}

	return ChatResponse{}, fmt.Errorf("exceeded maximum tool-use rounds (%d)", maxLoops)
}

func callClaude(ctx context.Context, apiKey string, messages []ChatMessage) (*anthropicResponse, error) {
	payload := anthropicRequest{
		Model:     anthropicModel,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     tools,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPI, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var ar anthropicResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if ar.Error != nil {
		return nil, fmt.Errorf("claude API error [%s]: %s", ar.Error.Type, ar.Error.Message)
	}

	return &ar, nil
}
