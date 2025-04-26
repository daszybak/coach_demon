package codeforces

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"unicode"
)

// ParseID splits a Codeforces problem ID (e.g. “1873G2”) into contest-number (“1873”) and index (“G2”).
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

func FetchStatement(ctx context.Context, baseURL, token, id string) (string, error) {
	contest, index, err := ParseID(id)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("https://codeforces.com/problemset/problem/%s/%s", contest, index)

	jsCodeTemplate := `
module.exports = async ({ page }) => {
  await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) ' +
    'AppleWebKit/537.36 (KHTML, like Gecko) ' +
    'Chrome/122.0.0.0 Safari/537.36');
  await page.setViewport({
    width: 1280 + Math.floor(Math.random() * 50),
    height: 800 + Math.floor(Math.random() * 50),
    deviceScaleFactor: 1,
  });
  await page.goto("%s", { waitUntil: "domcontentloaded", timeout: 60000 });
  await page.waitForTimeout(1000 + Math.random() * 2000);
  await page.evaluate(() => window.scrollBy(0, Math.floor(Math.random() * 400)));
  const isChallenge = await page.evaluate(() => {
    return !!document.querySelector('#cf-spinner-please-wait') ||
           !!document.querySelector('div[class*="Challenge"]') ||
           !!document.querySelector('input[name="cf_captcha_kind"]');
  });
  if (isChallenge) {
    await page.waitForTimeout(15000);
    await page.reload({ waitUntil: "domcontentloaded" });
  }
  await page.waitForSelector('.problem-statement', { timeout: 30000 });
  const html = await page.$eval('.problem-statement', el => el.innerHTML.trim());
  return html;
};
`

	jsCode := fmt.Sprintf(jsCodeTemplate, target)
	url := fmt.Sprintf("%s/function?token=%s", baseURL, token)

	var lastErr error
	const maxRetries = 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(jsCode)))
		httpReq.Header.Set("Content-Type", "application/javascript")

		res, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("cannot call fetcher: %w", err)
		} else {
			if res.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(res.Body)
				res.Body.Close()
				lastErr = fmt.Errorf("fetcher %d: %s", res.StatusCode, b)
				continue
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			return string(body), nil
		}

		wait := time.Second * time.Duration(1<<uint(attempt-1)) // 1s, 2s, 4s, etc.
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(wait):
		}
	}
	return "", lastErr
}
