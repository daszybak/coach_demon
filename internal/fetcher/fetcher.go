package fetcher

import "context"

// Service fetches the full HTML statement for a Codeforces problem.
type Service interface {
	Fetch(ctx context.Context, problemID string) (string, error)
}
