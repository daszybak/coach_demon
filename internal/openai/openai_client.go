package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go/responses"
	"log"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// --------------------------------------------------------------------
// Config + Client wrapper
// --------------------------------------------------------------------

type Config struct {
	APIKey       string        // OpenAI API key
	Model        string        // e.g. "gpt-4o-mini"
	SystemPrompt string        // role instruction
	Temperature  float64       // 0.0 – 1.0
	Timeout      time.Duration // per-request timeout
}

type Client struct {
	api          openai.Client
	model        string
	systemPrompt string
	temperature  float64
	timeout      time.Duration
}

func NewClient(cfg Config, client option.HTTPClient) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key missing")
	}
	if cfg.Model == "" {
		cfg.Model = "o3"
	}
	if cfg.Temperature <= 0 || cfg.Temperature > 1 {
		cfg.Temperature = 0.2
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}

	api := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithHTTPClient(client),
	)

	return &Client{
		api:          api,
		model:        cfg.Model,
		systemPrompt: cfg.SystemPrompt,
		temperature:  cfg.Temperature,
		timeout:      cfg.Timeout,
	}, nil
}

// --------------------------------------------------------------------
// Feedback
// --------------------------------------------------------------------

type Feedback struct {
	Feedback             string   `json:"feedback" jsonschema_description:"Feedback of the quality of my thinking process"`
	Suggestions          []string `json:"suggestions" jsonschema_description:"Suggestions of how I should improve my code and thinking in this situation"`
	Proofs               []string `json:"proofs" jsonschema_description:"Mathematical proofs or logical proofs for every step of this problem"`
	OptimalMetaCognition []string `json:"optima_meta_cognition" jsonschema_description:"What a top competitive programmer would be thinking in this situation"`
}

var FeedbackResponseSchema = GenerateSchema[Feedback]()

func (c *Client) GetFeedback(code, thoughts, problem string) (Feedback, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Construct the user message content
	userMessageContent := fmt.Sprintf(
		"Problem statement:\n%s\n\nMy code:\n%s\n\nMy thoughts:\n%s\n\n",
		problem, code, thoughts,
	)

	resp, err := c.api.Responses.New(ctx, responses.ResponseNewParams{
		Model:        c.model, // helper
		Instructions: openai.String(c.systemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(userMessageContent),
		},
		// Temperature: openai.Float(c.temperature), // helper

		Text: responses.ResponseTextConfigParam{ // value
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:        "coach_feedback", // helper → passes regex
					Schema:      FeedbackResponseSchema,
					Description: openai.String("Structured coach feedback"),
					Strict:      openai.Bool(true),
					Type:        "json_schema",
				},
			},
		},
	})

	if err != nil {
		return Feedback{}, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	log.Printf("%v", resp.OutputText())
	raw := resp.OutputText()
	var fb Feedback

	err = json.Unmarshal([]byte(raw), &fb)
	if err != nil {
		fb.Feedback = raw // Assign raw content to the fallback field
		return fb, fmt.Errorf("failed to unmarshall OpenAI JSON feedback: %w", err)
	}

	// Successfully unmarshalled JSON
	return fb, nil
}

type HistorySummary struct {
	Summary       string   `json:"summary" jsonschema_description:"Summary of my overall code and thinking evolution"`
	MainMistakes  []string `json:"main_mistakes" jsonschema_description:"Biggest mistakes I made across time"`
	GoodPractices []string `json:"good_practices" jsonschema_description:"Good practices I demonstrated"`
}

var HistorySummarySchema = GenerateSchema[HistorySummary]()

func (c *Client) SummarizeHistory(statement string, codes []string, thoughts []string) (HistorySummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Compose the full history text
	history := "Problem statement:\n" + statement + "\n\n"

	if len(codes) > 0 {
		history += "Code attempts:\n"
		for i, code := range codes {
			history += fmt.Sprintf("Attempt #%d:\n%s\n\n", i+1, code)
		}
	}

	if len(thoughts) > 0 {
		history += "Thoughts history:\n"
		for i, thought := range thoughts {
			history += fmt.Sprintf("Thought #%d:\n%s\n\n", i+1, thought)
		}
	}

	// Send to OpenAI
	resp, err := c.api.Responses.New(ctx, responses.ResponseNewParams{
		Model:        c.model,
		Instructions: openai.String(c.systemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(history),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:        "coach_summary",
					Schema:      HistorySummarySchema,
					Description: openai.String("Structured history summary"),
					Strict:      openai.Bool(true),
					Type:        "json_schema",
				},
			},
		},
	})
	if err != nil {
		return HistorySummary{}, fmt.Errorf("failed to call OpenAI API for summary: %w", err)
	}

	log.Printf("%v", resp.OutputText())
	raw := resp.OutputText()
	var summary HistorySummary

	err = json.Unmarshal([]byte(raw), &summary)
	if err != nil {
		summary.Summary = raw
		return summary, fmt.Errorf("failed to unmarshal OpenAI JSON summary: %w", err)
	}

	return summary, nil
}
