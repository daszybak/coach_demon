package fetcher

import (
	"coach_demon/pkg/codeforces"
	"context"
)

// Browserless wraps the /content endpoint of browserless/chrome.
type Browserless struct {
	BaseURL string // http://fetcher:3000
	Token   string // "cf-fun"
}

func (b *Browserless) Fetch(ctx context.Context, id string) (string, error) {
	return codeforces.FetchStatement(ctx, b.BaseURL, id)
}

// Constructor, so callers never new() directly.
func NewBrowserless(baseURL, token string) *Browserless {
	return &Browserless{BaseURL: baseURL, Token: token}
}
