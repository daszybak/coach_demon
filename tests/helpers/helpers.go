package helpers

import (
	"bytes"
	"context"
	"github.com/spf13/viper"
	"html/template"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func LoadConfig(t *testing.T) {
	t.Helper()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")

	_ = viper.ReadInConfig()
}

type span struct {
	Time     string
	Method   string
	URL      string
	Status   int
	Request  string
	Response string
}

type TraceRoundTripper struct {
	next http.RoundTripper
	mu   sync.Mutex
	log  []span
}

func New(base http.RoundTripper) *TraceRoundTripper {
	return &TraceRoundTripper{next: base}
}

func (t *TraceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// copy request
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	res, err := t.next.RoundTrip(req)
	if err != nil {
		return res, err
	}

	// copy response
	var resBody []byte
	if res.Body != nil {
		resBody, _ = io.ReadAll(res.Body)
		res.Body = io.NopCloser(bytes.NewReader(resBody))
	}

	t.mu.Lock()
	t.log = append(t.log, span{
		Time:     time.Now().Format(time.RFC3339),
		Method:   req.Method,
		URL:      req.URL.String(),
		Status:   res.StatusCode,
		Request:  string(reqBody),
		Response: string(resBody),
	})
	t.mu.Unlock()
	return res, nil
}

func (t *TraceRoundTripper) DumpHTML(path string) error {
	const tpl = `
	<!doctype html><html><head><meta charset="utf-8">
	<title>OpenAI Trace</title><style>
	body{font-family:sans-serif;margin:0 2rem}
	h2{margin-top:2rem}
	pre{background:#f7f7f7;border:1px solid #ddd;padding:8px;white-space:pre-wrap;color:gray}
	</style></head><body>
	<h1>HTTP Trace</h1>{{range .}}
	<h2>{{.Time}} — {{.Method}} {{.URL}} ({{.Status}})</h2>
	<b>Request</b><pre>{{.Request}}</pre>
	<b>Response</b><pre>{{.Response}}</pre>{{end}}
	</body></html>`
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return template.Must(template.New("t").Parse(tpl)).Execute(f, t.log)
}

func TimeoutContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx, cancel
}
