package llms

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
)

var ErrGenerationFailed = errors.New("generation failed")

type LLM interface {
	Generate(prompt string) (string, error)
}

func New() LLM {
	return &OpenAI{}
}

type OpenAI struct {
	client openai.Client
}

func (o *OpenAI) Generate(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	res, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})

	if err != nil {
		slog.Error("openai error: chat completions failed", "error", err)
		return "", ErrGenerationFailed
	}

	return res.Choices[0].Message.Content, nil
}
