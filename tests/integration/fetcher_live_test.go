//go:build integration

package integration

import (
	"coach_demon/internal/fetcher"
	"coach_demon/tests/helpers"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLiveFetcher(t *testing.T) {
	helpers.LoadConfig(t)

	tracer := helpers.New(http.DefaultTransport)

	service := fetcher.NewBrowserless(
		viper.GetString("FETCHER_ENDPOINT"),
		viper.GetString("FETCHER_TOKEN"),
	)

	ctx, cancel := helpers.TimeoutContext(t, 120*time.Second)
	defer cancel()

	html, err := service.Fetch(ctx, "1873G2")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if html == "" {
		t.Fatal("Fetched empty HTML!")
	}

	t.Logf("Fetched %d characters", len(html))

	_ = os.MkdirAll("tests", 0o755)
	_ = tracer.DumpHTML(filepath.Join("tests", "fetcher_trace.html"))
}
