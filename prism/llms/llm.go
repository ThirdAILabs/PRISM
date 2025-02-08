package llms

import (
	"context"
	"errors"
	"log/slog"
	"time"

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
	return &OpenAI{}
}

const (
	GPT4oMini = openai.ChatModelGPT4oMini
)

type OpenAI struct {
	client openai.Client
}

func (o *OpenAI) Generate(prompt string, opts *Options) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	messages := make([]openai.ChatCompletionMessageParamUnion, 2)

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
