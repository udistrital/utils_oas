package auditoria

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

const (
	authorizationKey = "Authorization"
	userKey          = "user"
	userInfoURL      = "https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo"
)

var appName, _ = beego.AppConfig.String("appname")
var globalLogger = &customSQLLogger{}
var authHTTPClient = &http.Client{Timeout: 30 * time.Second}
var c cache.Cache

type usuario struct {
	Documento          string `json:"documento"`
	DocumentoCompuesto string `json:"documento_compuesto"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	Sub                string `json:"sub"`
}

type requestLog struct {
	AppName      string         `json:"app_name"`
	Agent        string         `json:"agent,omitempty"`
	Data         map[string]any `json:"data"`
	Date         string         `json:"date"`
	Host         string         `json:"host"`
	IPUser       string         `json:"ip_user"`
	Method       string         `json:"method"`
	Path         string         `json:"path"`
	Query        string         `json:"query,omitempty"`
	Schema       string         `json:"schema,omitempty"`
	SQLStatement string         `json:"sql_statement,omitempty"`
	Status       int            `json:"status"`
	TraceID      string         `json:"trace_id,omitempty"`
	User         string         `json:"user"`
}

// customSQLLogger intercepts beego ORM debug output to capture the last executed
// SQL statement for audit logging.
//
// NOTE: beego v1's ORM logger is global — there is no way to associate a query
// with a specific request. As a result this implementation captures only the
// last query written globally, which is inaccurate under concurrent load.
// A mutex makes reads/writes race-safe, but the value may still belong to a
// different request.
//
// With beego v2 this is properly solvable: use orm.AddGlobalFilterChain with
// an orm.Filter that receives the context.Context, and store queries per-request
// using context.WithValue. Handlers must call o.WithContext(ctx) for the
// association to work.
type customSQLLogger struct {
	mu        sync.Mutex
	lastQuery string
}

func InitMiddleware() {
	var err error
	c, err = cache.NewCache("memory", `{"interval":300}`)
	if err != nil {
		logs.Error("error al inicializar el cache:", err)
		return
	}

	orm.DebugLog = orm.NewLog(globalLogger)
	logs.Info("middleware inicializado correctamente.")

	beego.InsertFilter("/:version/*", beego.BeforeExec, validateAndSetAuth)
	beego.InsertFilter("/:version/*", beego.AfterExec, LogRequest, beego.WithReturnOnOutput(false))
}

func validateAndSetAuth(ctx *beegoCtx.Context) {
	token := ctx.Request.Header.Get(authorizationKey)
	if token == "" {
		if strings.HasPrefix(ctx.Input.Context.Request.Host, "localhost") {
			return
		}
		// debería retornar 401
		logs.Warn("missing access token")
		// ctx.Abort(401, "unauthorized")
		return
	}

	reqCtx := context.WithValue(ctx.Request.Context(), authorizationKey, token)

	if strings.HasPrefix(ctx.Input.Context.Request.Host, "localhost") {
		return
	}

	cachedUser, err := c.Get(reqCtx, token)
	if err == nil {
		if user, ok := cachedUser.(string); ok && user != "" {
			ctx.Request = ctx.Request.WithContext(context.WithValue(reqCtx, userKey, user))
			return
		}
	}

	var user usuario
	if status, err := getWithContext(reqCtx, &user); err != nil {
		logs.Error("error al validar el token: %v, status %d", err, status)
		// debería retornar 401
		// ctx.Abort(401, "unauthorized")
		return
	}

	if err := c.Put(reqCtx, token, user.Sub, 60*time.Minute); err != nil {
		logs.Error("error al guardar el token el cache:", err)
		return
	}
	ctx.Request = ctx.Request.WithContext(context.WithValue(reqCtx, userKey, user.Sub))
}

func LogRequest(ctx *beegoCtx.Context) {
	logRequestWithLogger(ctx, globalLogger)
}

func logRequestWithLogger(ctx *beegoCtx.Context, logger *customSQLLogger) {
	user, _ := ctx.Request.Context().Value(userKey).(string)

	status := ctx.ResponseWriter.Status
	if status == 0 {
		status = http.StatusOK
	}

	entry := requestLog{
		AppName:      appName,
		Agent:        ctx.Input.UserAgent(),
		Data:         sanitizeInputData(ctx.Input.Data()),
		Date:         time.Now().Format(time.RFC3339),
		Host:         ctx.Request.Host,
		IPUser:       ctx.Input.IP(),
		Method:       ctx.Request.Method,
		Path:         ctx.Request.URL.Path,
		Query:        ctx.Request.URL.RawQuery,
		Schema:       ctx.Input.Scheme(),
		SQLStatement: logger.GetLastQuery(),
		Status:       status,
		TraceID:      xray.TraceID(ctx.Request.Context()),
		User:         user,
	}

	if jsonData, err := json.Marshal(entry); err != nil {
		logs.Error("error al serializar log a JSON:", err)
	} else {
		logs.Info(string(jsonData))
	}
}

func sanitizeInputData(input any) map[string]any {
	switch data := input.(type) {
	case map[string]any:
		return data
	case map[any]any:
		converted := make(map[string]any, len(data))
		for k, v := range data {
			converted[fmt.Sprintf("%v", k)] = v
		}
		return converted
	}
	return nil
}

func getWithContext(ctx context.Context, target any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if token, ok := ctx.Value(authorizationKey).(string); ok && token != "" {
		req.Header.Set(authorizationKey, token)
	}

	resp, err := authHTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		return resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, fmt.Errorf("response body could not be decoded into target: %w", err)
	}

	return resp.StatusCode, nil
}

func (l *customSQLLogger) Write(p []byte) (int, error) {
	logMessage := string(p)

	re := regexp.MustCompile(`\[(SELECT|INSERT|UPDATE|DELETE).*`)
	match := re.FindString(logMessage)

	l.mu.Lock()
	l.lastQuery = match
	l.mu.Unlock()

	return len(p), nil
}

func (l *customSQLLogger) GetLastQuery() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.TrimSpace(l.lastQuery)
}
