package auditoria

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/udistrital/utils_oas/request"
)

const (
	authorizationKey = "Authorization"
	userKey          = "user"
)

var appName = beego.AppConfig.String("appname")
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

	customLogger := &customSQLLogger{}
	orm.DebugLog = orm.NewLog(customLogger)
	logs.Info("middleware inicializado correctamente.")

	beego.InsertFilter("/:version/*", beego.BeforeExec, validateAndSetAuth, false)
	beego.InsertFilter("/:version/*", beego.AfterExec, logRequest(customLogger), false)
}

func validateAndSetAuth(ctx *beegoCtx.Context) {
	token := ctx.Request.Header.Get("Authorization")
	if token == "" {
		return
	}

	reqCtx := context.WithValue(ctx.Request.Context(), authorizationKey, token)
	if sub, ok := c.Get(token).(string); ok && sub != "" {
		ctx.Request = ctx.Request.WithContext(context.WithValue(reqCtx, userKey, sub))
		return
	}

	var user usuario
	if _, err := request.GetWithContext(reqCtx, "https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo", &user); err != nil {
		logs.Error("error al validar el token:", err)
		return
	}

	c.Put(token, user.Sub, 60*time.Minute)
	ctx.Request = ctx.Request.WithContext(context.WithValue(reqCtx, userKey, user.Sub))
}

func logRequest(logger *customSQLLogger) func(ctx *beegoCtx.Context) {
	return func(ctx *beegoCtx.Context) {
		user, _ := ctx.Request.Context().Value(userKey).(string)

		entry := requestLog{
			AppName:      appName,
			Agent:        ctx.Request.Header.Get("User-Agent"),
			Data:         sanitizeInputData(ctx.Input.Data()),
			Date:         time.Now().Format(time.RFC3339),
			Host:         ctx.Request.Host,
			IPUser:       ctx.Input.IP(),
			Method:       ctx.Request.Method,
			Path:         ctx.Request.URL.Path,
			Query:        ctx.Request.URL.RawQuery,
			Schema:       ctx.Input.Scheme(),
			SQLStatement: logger.GetLastQuery(),
			User:         user,
		}

		if jsonData, err := json.Marshal(entry); err != nil {
			logs.Error("error al serializar log a JSON:", err)
		} else {
			logs.Info(string(jsonData))
		}
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
