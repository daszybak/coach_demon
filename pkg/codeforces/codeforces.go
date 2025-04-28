package codeforces

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unicode"
)

func ParseID(id string) (contest, index string, err error) {
	for i, r := range id {
		if !unicode.IsDigit(r) {
			if i == 0 {
				return "", "", errors.New("problem id starts with a letter, expected digits")
			}
			return id[:i], id[i:], nil
		}
	}
	return "", "", fmt.Errorf("cannot split %q into contest/index", id)
}

func FetchStatement(ctx context.Context, fetcherURL, id string) (string, error) {
	contest, index, err := ParseID(id)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("https://codeforces.com/problemset/problem/%s/%s", contest, index)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fetcherURL, bytes.NewBufferString(target))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusOK {
		return string(body), nil
	}
	return "", fmt.Errorf("fetch error: %s", string(body))
}
