package auditoria

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/patrickmn/go-cache"

	"regexp"
	"strings"
)

type Usuario struct {
	Sub  string `json:"sub"`
	Date time.Time
}

var userMap = make(map[string]string)
var c = cache.New(60*time.Minute, 10*time.Minute)

func getUserInfo2(ctx *context.Context) (u string) {
	var usuario Usuario

	if val, ok := userMap[ctx.Request.Header["Authorization"][0]]; ok {
		return val
	} else {
		if err := GetJsonWithHeader("https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo", &usuario, ctx); err == nil {
			userMap[ctx.Request.Header["Authorization"][0]] = usuario.Sub
			return usuario.Sub
		} else {
			userMap[ctx.Request.Header["Authorization"][0]] = "No user"
			return "No user"
		}
	}
}

func getUserInfo(ctx *context.Context) (u string) {
	var usuario Usuario
	if x, found := c.Get(ctx.Request.Header["Authorization"][0]); found {
		foo := x.(string)
		return foo
	} else {
		if err := GetJsonWithHeader("https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo", &usuario, ctx); err == nil {
			c.Set(ctx.Request.Header["Authorization"][0], usuario.Sub, cache.DefaultExpiration)
			return usuario.Sub

		} else {
			c.Set(ctx.Request.Header["Authorization"][0], "No user", cache.DefaultExpiration)
			return "No user"
		}
	}
}

func ListenRequest(logger *customSQLLogger) func(ctx *context.Context) {
	return func(ctx *context.Context) {
		if ctx.Request.URL.String() == "/" {
			return
		}

		sqlQuery := logger.GetLastQuery()
		if sqlQuery == "" {
			sqlQuery = "No se registró sentencia SQL"
		}

		defer func() {
			if r := recover(); r != nil {
				logData := map[string]interface{}{
					"app_name":   beego.AppConfig.String("appname"),
					"host":       ctx.Request.Host,
					"end_point":  ctx.Request.URL.String(),
					"method":     ctx.Request.Method,
					"date":       time.Now().Format(time.RFC3339),
					"sql_orm":    sqlQuery,
					"ip_user":    ctx.Input.IP(),
					"user_agent": getUserAgent(ctx),
					"user":       "Error WSO2",
					"data":       sanitizeInputData(ctx.Input.Data()),
				}
				logAsJSON(logData)
			}
		}()

		user := getUserInfo(ctx)

		logData := map[string]interface{}{
			"app_name":   beego.AppConfig.String("appname"),
			"host":       ctx.Request.Host,
			"end_point":  ctx.Request.URL.String(),
			"method":     ctx.Request.Method,
			"date":       time.Now().Format(time.RFC3339),
			"sql_orm":    sqlQuery,
			"ip_user":    ctx.Input.IP(),
			"user_agent": getUserAgent(ctx),
			"user":       user,
			"data":       sanitizeInputData(ctx.Input.Data()),
		}

		logAsJSON(logData)
	}
}

func logAsJSON(data map[string]interface{}) {

	jsonData, err := json.Marshal(data["data"])
	if err != nil {
		beego.Error("Error al serializar el campo 'data' a JSON:", err)
		jsonData = []byte("{}")
	}

	var pruebaLog = "{app_name: " + data["app_name"].(string) +
		", host: " + data["host"].(string) +
		", end_point: " + data["end_point"].(string) +
		", method: " + data["method"].(string) +
		", date: " + data["date"].(string) +
		", sql_orm: {" + data["sql_orm"].(string) +
		"}, ip_user: " + data["ip_user"].(string) +
		", user_agent: " + data["user_agent"].(string) +
		", user: " + data["user"].(string) +
		", data: " + string(jsonData) +
		"}"

	logs.Info(pruebaLog)
}

func sanitizeInputData(input interface{}) interface{} {
	if data, ok := input.(map[interface{}]interface{}); ok {
		converted := make(map[string]interface{})
		for key, value := range data {
			converted[fmt.Sprintf("%v", key)] = value
		}
		return converted
	}
	return input
}

func getUserAgent(ctx *context.Context) string {
	if len(ctx.Request.Header["User-Agent"]) > 0 {
		return ctx.Request.Header["User-Agent"][0]
	}
	return "Desconocido"
}

func InitMiddleware() {
	customLogger := &customSQLLogger{}
	orm.DebugLog = orm.NewLog(customLogger)

	logs.Info("middleware inicializado correctamente.")
	beego.InsertFilter("*", beego.AfterExec, ListenRequest(customLogger), false)
}

type customSQLLogger struct {
	lastQuery string
}

func (l *customSQLLogger) Write(p []byte) (n int, err error) {
	logMessage := string(p)

	re := regexp.MustCompile(`\[(SELECT|INSERT|UPDATE|DELETE).*`)
	match := re.FindString(logMessage)
	l.lastQuery = match

	return len(p), nil
}

func (l *customSQLLogger) GetLastQuery() string {
	query := strings.TrimSpace(l.lastQuery)
	return query
}
