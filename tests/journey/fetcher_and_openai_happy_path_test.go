package journey

import (
	"coach_demon/internal/fetcher"
	"coach_demon/internal/openai"
	"coach_demon/tests/helpers"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJourney_FetcherAndOpenAI(t *testing.T) {
	helpers.LoadConfig(t)

	tracer := helpers.New(http.DefaultTransport)
	httpClient := &http.Client{Transport: tracer, Timeout: 30 * time.Second}

	fetcherSvc := fetcher.NewBrowserless(
		viper.GetString("FETCHER_ENDPOINT"),
		viper.GetString("FETCHER_TOKEN"),
	)

	aiClient, err := openai.NewClient(openai.Config{
		APIKey:       viper.GetString("OPENAI_API_KEY"),
		Model:        viper.GetString("OPENAI_MODEL"),
		SystemPrompt: viper.GetString("OPENAI_SYSTEM_PROMPT"),
		Timeout:      30 * time.Second,
	}, httpClient)
	if err != nil {
		t.Fatalf("openai: %v", err)
	}

	ctx, cancel := helpers.TimeoutContext(t, 360*time.Second)
	defer cancel()

	problemHTML, err := fetcherSvc.Fetch(ctx, "2A")
	if err != nil {
		t.Fatalf("fetch problem: %v", err)
	}
	if problemHTML == "" {
		t.Fatal("fetched empty problem statement")
	}

	feedback, err := aiClient.GetFeedback("int a;", "thinking hard...", problemHTML)
	if err != nil {
		t.Fatalf("openai feedback: %v", err)
	}

	if feedback.Feedback == "" {
		t.Fatal("got empty OpenAI feedback")
	}

	t.Logf("AI Feedback: %s", feedback.Feedback)

	_ = os.MkdirAll("tests", 0o755)
	_ = tracer.DumpHTML(filepath.Join("tests", "journey_trace.html"))
}
