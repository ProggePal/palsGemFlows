package gemini

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	client *genai.Client
}

// CallGemini matches the required MVP signature.
// It creates a short-lived client using GEMINI_API_KEY.
func CallGemini(model, systemPrompt, userPrompt string) (string, error) {
	ctx := context.Background()
	c, err := NewClientFromEnv(ctx)
	if err != nil {
		return "", err
	}
	defer c.Close()
	return c.Generate(ctx, model, systemPrompt, userPrompt)
}

func NewClientFromEnv(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not set")
	}
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Client{client: c}, nil
}

func (c *Client) Close() {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Close()
}

func (c *Client) Generate(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	if c == nil || c.client == nil {
		return "", errors.New("gemini client not initialized")
	}
	if model == "" {
		return "", errors.New("model is required")
	}

	m := c.client.GenerativeModel(model)
	if systemPrompt != "" {
		m.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}
	}

	resp, err := m.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return "", err
	}

	text, err := firstCandidateText(resp)
	if err != nil {
		return "", err
	}
	return text, nil
}

func firstCandidateText(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", errors.New("empty response")
	}

	var out string
	for _, part := range resp.Candidates[0].Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			out += string(v)
		case fmt.Stringer:
			out += v.String()
		default:
			// ignore non-text parts for MVP
		}
	}
	if out == "" {
		return "", errors.New("no text in response")
	}
	return out, nil
}
