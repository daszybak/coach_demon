//go:build integration

package integration

import (
	"coach_demon/internal/openai"
	"coach_demon/tests/helpers"
	"github.com/spf13/viper"
	"net/http" //  ← missing import
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLiveOpenAI(t *testing.T) {
	helpers.LoadConfig(t)

	tracer := helpers.New(http.DefaultTransport) // use ctor
	httpClient := &http.Client{Transport: tracer, Timeout: 20 * time.Second}

	cli, err := openai.NewClient(openai.Config{
		APIKey:       viper.GetString("OPENAI_API_KEY"),
		Model:        viper.GetString("OPENAI_MODEL"),
		SystemPrompt: viper.GetString("OPENAI_SYSTEM_PROMPT"),
		Timeout:      20 * time.Second,
	}, httpClient)
	if err != nil {
		t.Fatalf("cannot init client: %v", err)
	}

	_, err = cli.GetFeedback("int a;", "stub", "A+B")
	if err != nil {
		t.Fatalf("GetFeedback: %v", err)
	}

	_ = os.MkdirAll("tests", 0o755)
	_ = tracer.DumpHTML(filepath.Join("tests", "openai_trace.html"))
}
