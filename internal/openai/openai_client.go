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
	Feedback             string `json:"feedback" jsonschema_description:"Feedback of the quality of my thinking process"`
	Proof                string `json:"proof" jsonschema_description:"Mathematical proofs or logical proofs for every step of this problem"`
	OptimalMetaCognition string `json:"optima_meta_cognition" jsonschema_description:"What a top competitive programmer would be thinking in this situation"`
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

type Summary struct {
	Feedback             string `json:"feedback" jsonschema_description:"Summarize the feedback that I received from AI."`
	Proof                string `json:"summary" jsonschema_description:"Summarize the most important proofs."`
	OptimalMetaCognition string `json:"optimal_meta_cognition" jsonschema_description:"Summarize the optimal meta cognition that a top competitive programmer should have based on the AI inputs."`
}

var SummarySchema = GenerateSchema[Summary]()

func (c *Client) SummarizeFeedback(statement string, feedbacks, proofs, optimalMetaCognitions []string) (Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Compose the full history text
	history := "Problem statement:\n" + statement + "\n\n"

	if len(feedbacks) > 0 {
		history += "Feedbacks:\n"
		for i, feedback := range feedbacks {
			history += fmt.Sprintf("Feedback #%d:\n%s\n\n", i+1, feedback)
		}
	}

	if len(proofs) > 0 {
		history += "Proofs:\n"
		for i, proof := range proofs {
			history += fmt.Sprintf("Proof #%d:\n%s\n\n", i+1, proof)
		}
	}

	if len(optimalMetaCognitions) > 0 {
		history += "Optimal Meta Cognitions:\n"
		for i, oMC := range optimalMetaCognitions {
			history += fmt.Sprintf("Optimal Meta Cognition #%d:\n%s\n\n", i+1, oMC)
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
					Schema:      SummarySchema,
					Description: openai.String("Structured history summary"),
					Strict:      openai.Bool(true),
					Type:        "json_schema",
				},
			},
		},
	})
	if err != nil {
		return Summary{}, fmt.Errorf("failed to call OpenAI API for summary: %w", err)
	}

	log.Printf("%v", resp.OutputText())
	raw := resp.OutputText()
	var summary Summary

	err = json.Unmarshal([]byte(raw), &summary)
	if err != nil {
		summary.Feedback = raw
		return summary, fmt.Errorf("failed to unmarshal OpenAI JSON summary: %w", err)
	}

	return summary, nil
}
