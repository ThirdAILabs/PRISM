package llms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
)

var ErrGenerationFailed = errors.New("generation failed")

type Options struct {
	Model        string
	ZeroTemp     bool
	SystemPrompt string
}

type LLM interface {
	Generate(prompt string, opts *Options) (string, error)
}

func New() LLM {
	return NewOpenAI()
}

const (
	GPT4oMini = openai.ChatModelGPT4oMini
	GPT4o     = openai.ChatModelGPT4o
)

type OpenAI struct {
	client *openai.Client
}

func NewOpenAI() LLM {
	return &OpenAI{client: openai.NewClient()}
}

func (o *OpenAI) Generate(prompt string, opts *Options) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, 2)

	if opts != nil && len(opts.SystemPrompt) > 0 {
		messages = append(messages, openai.SystemMessage(opts.SystemPrompt))
	}
	messages = append(messages, openai.UserMessage(prompt))

	chatOpts := openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		Model:    openai.String(openai.ChatModelGPT4o),
	}

	if opts != nil && len(opts.Model) > 0 {
		chatOpts.Model = openai.String(opts.Model)
	}

	if opts != nil && opts.ZeroTemp {
		chatOpts.Temperature = openai.Float(0)
	}

	res, err := o.client.Chat.Completions.New(ctx, chatOpts)
	if err != nil {
		slog.Error("openai error: chat completions failed", "error", err)
		return "", ErrGenerationFailed
	}

	return res.Choices[0].Message.Content, nil
}

type PerplexityAI struct {
	client *resty.Client
}

func NewPerplexityAI(apiKey string) *PerplexityAI {
	return &PerplexityAI{
		client: resty.New().
			SetBaseURL("https://api.perplexity.ai").
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json").
			SetTimeout(120 * time.Second),
	}
}

type PerplexityOptions struct {
	Temperature      float32                `json:"temperature,omitempty"`
	ResponseFormat   map[string]interface{} `json:"response_format,omitempty"`
	WebSearchOptions map[string]interface{} `json:"web_search_options,omitempty"`
	Model            string                 `json:"model,omitempty"`
}

type PerplexityResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message"`
	} `json:"choices"`
	Citations []string `json:"citations"`
}

func (p *PerplexityAI) createOptionsPayload(opt *PerplexityOptions) map[string]interface{} {
	payload := make(map[string]interface{})
	if opt != nil {
		if opt.Temperature != 0 {
			payload["temperature"] = opt.Temperature
		}
		if opt.Model != "" {
			payload["model"] = opt.Model
		}
		if opt.ResponseFormat != nil {
			payload["response_format"] = opt.ResponseFormat
		}
		if opt.WebSearchOptions != nil {
			payload["web_search_options"] = opt.WebSearchOptions
		}
	}
	if _, ok := payload["model"]; !ok {
		payload["model"] = "sonar-pro"
	}
	return payload
}

func (p *PerplexityAI) Generate(userPrompt, systemPrompt string, opt *PerplexityOptions) (string, []string, error) {
	messages := []map[string]string{}
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	messages = append(messages, map[string]string{
		"role":    "user",
		"content": userPrompt,
	})

	payload := map[string]interface{}{
		"messages": messages,
		"stream":   false,
	}

	updatedOpts := p.createOptionsPayload(opt)
	for k, v := range updatedOpts {
		payload[k] = v
	}

	resp, err := p.client.R().
		SetBody(payload).
		Post("/chat/completions")

	if err != nil {
		return "", make([]string, 0), fmt.Errorf("failed to send request: %w", err)
	}

	if resp.IsError() {
		return "", make([]string, 0), fmt.Errorf("request failed with status %s: %s", resp.Status(), resp.Body())
	}

	var result PerplexityResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", make([]string, 0), fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", make([]string, 0), fmt.Errorf("no choices in response")
	}

	content := result.Choices[0].Message.Content
	if content == "" {
		return "", make([]string, 0), fmt.Errorf("empty content in message")
	}

	return content, result.Citations, nil
}
