package auditoria

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

// withRunMode overrides beego's run mode for the duration of the test.
func withRunMode(t *testing.T, mode string) {
	t.Helper()

	prev := beego.BConfig.RunMode
	beego.BConfig.RunMode = mode
	t.Cleanup(func() { beego.BConfig.RunMode = prev })
}

// stubTransport stands in for the userinfo endpoint, recording what the
// package sent so tests can assert on the outgoing request.
type stubTransport struct {
	mu        sync.Mutex
	status    int
	body      string
	err       error
	calls     int
	gotAuth   string
	gotAccept string
	gotURL    string
}

func (s *stubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++
	s.gotAuth = r.Header.Get("Authorization")
	s.gotAccept = r.Header.Get("Accept")
	s.gotURL = r.URL.String()

	if s.err != nil {
		return nil, s.err
	}

	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
		Request:    r,
	}, nil
}

// installStub points the package at a fresh cache and the given transport,
// restoring both when the test ends.
func installStub(t *testing.T, stub *stubTransport) *stubTransport {
	t.Helper()

	newCache, err := cache.NewCache("memory", `{"interval":300}`)
	if err != nil {
		t.Fatalf("cache: %v", err)
	}

	prevCache, prevClient := c, authHTTPClient
	c = newCache
	authHTTPClient = &http.Client{Transport: stub}
	t.Cleanup(func() { c, authHTTPClient = prevCache, prevClient })

	return stub
}

func setup(t *testing.T, status int, body string) *stubTransport {
	t.Helper()
	return installStub(t, &stubTransport{status: status, body: body})
}

// newCtx builds a beego context around an inbound request to host, optionally
// carrying an Authorization header.
func newCtx(host, token string) *beegoCtx.Context {
	req := httptest.NewRequest(http.MethodGet, "http://"+host+"/v1/foo?page=2", nil)
	req.Host = host
	if token != "" {
		req.Header.Set(authorizationKey, token)
	}

	ctx := beegoCtx.NewContext()
	ctx.Reset(httptest.NewRecorder(), req)

	return ctx
}

// logCapture collects everything the package writes through beego's logger.
type logCapture struct {
	mu   sync.Mutex
	msgs []string
}

func (lc *logCapture) Init(string) error { return nil }
func (lc *logCapture) Destroy()          {}
func (lc *logCapture) Flush()            {}

func (lc *logCapture) SetFormatter(logs.LogFormatter) {}

func (lc *logCapture) WriteMsg(lm *logs.LogMsg) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.msgs = append(lc.msgs, lm.Msg)

	return nil
}

func (lc *logCapture) lines() []string {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	return append([]string(nil), lc.msgs...)
}

var (
	captured    = &logCapture{}
	captureOnce sync.Once
)

// captureLogs routes beego's global logger into a buffer for the test's duration.
func captureLogs(t *testing.T) *logCapture {
	t.Helper()

	captureOnce.Do(func() {
		logs.Register("testcapture", func() logs.Logger { return captured })
	})

	captured.mu.Lock()
	captured.msgs = nil
	captured.mu.Unlock()

	logs.Reset()
	if err := logs.SetLogger("testcapture"); err != nil {
		t.Fatalf("set logger: %v", err)
	}
	t.Cleanup(func() { logs.Reset() })

	return captured
}

func TestSanitizeInputData(t *testing.T) {
	t.Run("string keys pass through", func(t *testing.T) {
		in := map[string]any{"a": 1}

		got := sanitizeInputData(in)
		if len(got) != 1 || got["a"] != 1 {
			t.Errorf("got %v, want {a:1}", got)
		}
	})

	t.Run("any keys are stringified", func(t *testing.T) {
		got := sanitizeInputData(map[any]any{"message": "unauthorized", 7: "seven"})

		if got["message"] != "unauthorized" {
			t.Errorf("message = %v, want unauthorized", got["message"])
		}
		if got["7"] != "seven" {
			t.Errorf("key 7 = %v, want seven under string key %q", got["7"], "7")
		}
	})

	t.Run("unsupported types yield nil", func(t *testing.T) {
		for _, in := range []any{nil, "a string", 42, []string{"x"}} {
			if got := sanitizeInputData(in); got != nil {
				t.Errorf("sanitizeInputData(%v) = %v, want nil", in, got)
			}
		}
	})
}

func TestCustomSQLLogger(t *testing.T) {
	t.Run("captures a SQL statement", func(t *testing.T) {
		l := &customSQLLogger{}

		msg := "[ORM] 2026/09/11 - [SELECT * FROM usuario WHERE id = 1] - 1.2ms  \n"
		n, err := l.Write([]byte(msg))
		if err != nil || n != len(msg) {
			t.Fatalf("Write = (%d, %v), want (%d, nil)", n, err, len(msg))
		}

		got := l.GetLastQuery()
		if !strings.HasPrefix(got, "[SELECT * FROM usuario") {
			t.Errorf("GetLastQuery = %q, want it to start with the SELECT", got)
		}
		if got != strings.TrimSpace(got) {
			t.Errorf("GetLastQuery = %q, want surrounding space trimmed", got)
		}
	})

	t.Run("each verb is recognised", func(t *testing.T) {
		for _, verb := range []string{"SELECT", "INSERT", "UPDATE", "DELETE"} {
			l := &customSQLLogger{}
			l.Write([]byte("[ORM] [" + verb + " something]"))

			if !strings.Contains(l.GetLastQuery(), verb) {
				t.Errorf("%s not captured, got %q", verb, l.GetLastQuery())
			}
		}
	})

	t.Run("a non-SQL write clears the previous query", func(t *testing.T) {
		l := &customSQLLogger{}
		l.Write([]byte("[ORM] [SELECT 1]"))
		if l.GetLastQuery() == "" {
			t.Fatal("precondition: expected a captured query")
		}

		l.Write([]byte("[ORM] connection established"))
		if got := l.GetLastQuery(); got != "" {
			t.Errorf("GetLastQuery = %q, want empty after a non-SQL write", got)
		}
	})

	t.Run("concurrent access is race free", func(t *testing.T) {
		l := &customSQLLogger{}

		var wg sync.WaitGroup
		for range 50 {
			wg.Add(2)
			go func() { defer wg.Done(); l.Write([]byte("[ORM] [SELECT 1]")) }()
			go func() { defer wg.Done(); _ = l.GetLastQuery() }()
		}
		wg.Wait()
	})
}

func TestGetUserInfo(t *testing.T) {
	t.Run("decodes the user and reports the status", func(t *testing.T) {
		setup(t, http.StatusOK, `{"sub":"abc","email":"e@u.edu.co","role":"admin","documento":"123","documento_compuesto":"CC123"}`)

		var user usuario
		status, err := getUserInfo(context.Background(), &user)
		if err != nil || status != http.StatusOK {
			t.Fatalf("getUserInfo = (%d, %v), want (200, nil)", status, err)
		}

		want := usuario{Documento: "123", DocumentoCompuesto: "CC123", Email: "e@u.edu.co", Role: "admin", Sub: "abc"}
		if user != want {
			t.Errorf("user = %+v, want %+v", user, want)
		}
	})

	t.Run("forwards the token from the context", func(t *testing.T) {
		stub := setup(t, http.StatusOK, `{"sub":"abc"}`)

		ctx := context.WithValue(context.Background(), authorizationKey, "Bearer tok")
		if _, err := getUserInfo(ctx, &usuario{}); err != nil {
			t.Fatalf("getUserInfo: %v", err)
		}

		if stub.gotAuth != "Bearer tok" {
			t.Errorf("Authorization = %q, want %q", stub.gotAuth, "Bearer tok")
		}
		if stub.gotAccept != "application/json" {
			t.Errorf("Accept = %q, want application/json", stub.gotAccept)
		}
		if stub.gotURL != userInfoURL {
			t.Errorf("URL = %q, want %q", stub.gotURL, userInfoURL)
		}
	})

	t.Run("omits the header when the context has no token", func(t *testing.T) {
		stub := setup(t, http.StatusOK, `{"sub":"abc"}`)

		if _, err := getUserInfo(context.Background(), &usuario{}); err != nil {
			t.Fatalf("getUserInfo: %v", err)
		}
		if stub.gotAuth != "" {
			t.Errorf("Authorization = %q, want it unset", stub.gotAuth)
		}
	})

	t.Run("reports non-2xx responses", func(t *testing.T) {
		setup(t, http.StatusUnauthorized, `{}`)

		status, err := getUserInfo(context.Background(), &usuario{})
		if err == nil {
			t.Fatal("want an error for a 401")
		}
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", status)
		}
	})

	t.Run("226 is still treated as success", func(t *testing.T) {
		setup(t, http.StatusIMUsed, `{"sub":"abc"}`)

		status, err := getUserInfo(context.Background(), &usuario{})
		if err != nil {
			t.Errorf("getUserInfo = (%d, %v), want no error at the 226 boundary", status, err)
		}
	})

	t.Run("reports an undecodable body", func(t *testing.T) {
		setup(t, http.StatusOK, `not json`)

		status, err := getUserInfo(context.Background(), &usuario{})
		if err == nil {
			t.Fatal("want a decode error")
		}
		if status != http.StatusOK {
			t.Errorf("status = %d, want the response status alongside the error", status)
		}
	})

	t.Run("reports a transport failure", func(t *testing.T) {
		installStub(t, &stubTransport{err: errors.New("dial refused")})

		status, err := getUserInfo(context.Background(), &usuario{})
		if err == nil {
			t.Fatal("want a transport error")
		}
		if status != 0 {
			t.Errorf("status = %d, want 0 when no response arrived", status)
		}
	})
}

// logEntry unmarshals the single JSON line LogRequest is expected to emit.
func logEntry(t *testing.T, lc *logCapture) map[string]any {
	t.Helper()

	lines := lc.lines()
	if len(lines) != 1 {
		t.Fatalf("got %d log lines, want exactly 1: %q", len(lines), lines)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("log line is not JSON (%v): %q", err, lines[0])
	}

	return entry
}

func TestLogRequest(t *testing.T) {
	t.Run("emits the request fields", func(t *testing.T) {
		lc := captureLogs(t)

		ctx := newCtx("api.example.com", "")
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), userKey, "user-abc"))
		ctx.Input.SetData("message", "unauthorized")
		ctx.ResponseWriter.Status = http.StatusCreated

		LogRequest(ctx)

		entry := logEntry(t, lc)
		checks := map[string]any{
			"method":  http.MethodGet,
			"path":    "/v1/foo",
			"query":   "page=2",
			"host":    "api.example.com",
			"user":    "user-abc",
			"status":  float64(http.StatusCreated),
			"ip_user": "192.0.2.1",
		}
		for field, want := range checks {
			if got := entry[field]; got != want {
				t.Errorf("%s = %v, want %v", field, got, want)
			}
		}

		data, ok := entry["data"].(map[string]any)
		if !ok || data["message"] != "unauthorized" {
			t.Errorf("data = %v, want it to carry message=unauthorized", entry["data"])
		}
	})

	t.Run("status falls back through the response writer, output, then 200", func(t *testing.T) {
		cases := []struct {
			name       string
			writer     int
			output     int
			wantStatus float64
		}{
			{"response writer wins", http.StatusCreated, http.StatusTeapot, float64(http.StatusCreated)},
			{"output is the fallback", 0, http.StatusNotFound, float64(http.StatusNotFound)},
			{"200 when neither is set", 0, 0, float64(http.StatusOK)},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				lc := captureLogs(t)

				ctx := newCtx("api.example.com", "")
				ctx.ResponseWriter.Status = tc.writer
				if tc.output != 0 {
					ctx.Output.SetStatus(tc.output)
				}

				LogRequest(ctx)

				if got := logEntry(t, lc)["status"]; got != tc.wantStatus {
					t.Errorf("status = %v, want %v", got, tc.wantStatus)
				}
			})
		}
	})

	t.Run("includes the last SQL statement and omits an absent trace", func(t *testing.T) {
		lc := captureLogs(t)

		prev := globalLogger
		globalLogger = &customSQLLogger{}
		t.Cleanup(func() { globalLogger = prev })
		globalLogger.Write([]byte("[ORM] [SELECT * FROM usuario]"))

		LogRequest(newCtx("api.example.com", ""))

		entry := logEntry(t, lc)
		if sql, _ := entry["sql_statement"].(string); !strings.Contains(sql, "SELECT * FROM usuario") {
			t.Errorf("sql_statement = %v, want the captured query", entry["sql_statement"])
		}
		if _, present := entry["trace_id"]; present {
			t.Error("trace_id should be omitted when there is no X-Ray segment")
		}
	})

	t.Run("an anonymous request logs an empty user", func(t *testing.T) {
		lc := captureLogs(t)

		LogRequest(newCtx("api.example.com", ""))

		if got := logEntry(t, lc)["user"]; got != "" {
			t.Errorf("user = %v, want an empty string", got)
		}
	})
}

func TestInitInitialisesTheCache(t *testing.T) {
	for _, tc := range []struct {
		name string
		init func()
	}{
		{"InitMiddleware", InitMiddleware},
		{"InitWithAuthEnforcer", InitWithAuthEnforcer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev := c
			c = nil
			t.Cleanup(func() { c = prev })

			tc.init()

			if c == nil {
				t.Fatal("cache was not initialised")
			}
			if err := c.Put(context.Background(), "k", "v", 0); err != nil {
				t.Errorf("cache unusable: %v", err)
			}
		})
	}
}
