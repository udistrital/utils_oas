package auditoria

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/udistrital/utils_oas/request"
)

var c cache.Cache

type usuario struct {
	Documento          string `json:"documento"`
	DocumentoCompuesto string `json:"documento_compuesto"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	Sub                string `json:"sub"`
}

func InitMiddleware() {
	var err error
	c, err = cache.NewCache("memory", `{"interval":300}`)
	if err != nil {
		logs.Error("Error al inicializar el cache:", err)
		return
	}

	customLogger := &customSQLLogger{}
	orm.DebugLog = orm.NewLog(customLogger)

	beego.InsertFilter("/*", beego.BeforeExec, func(ctx *beegoCtx.Context) {
		if auth := ctx.Request.Header.Get("Authorization"); auth != "" {
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(ctx.Request.Context(), request.AuthorizationKey, auth),
			)
		}
	}, false)

	logs.Info("middleware inicializado correctamente.")
	beego.InsertFilter("/:version/*", beego.AfterExec, auditRequest(customLogger), false)
}

func getUserInfo(ctx *beegoCtx.Context) string {
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "No user"
	}

	if x := c.Get(authHeader); x != nil {
		return x.(string)
	}

	var user usuario
	if _, err := request.GetWithContext(ctx.Request.Context(), "https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo", &user); err == nil {
		c.Put(authHeader, user.Sub, 60*time.Minute)
		logs.Info("Usuario obtenido y almacenado en cache:", user)
		return user.Sub
	} else {
		logs.Error("Error al obtener información del usuario:", err)
	}

	c.Put(authHeader, "No user", 60*time.Minute)
	return "No user"
}

func auditRequest(logger *customSQLLogger) func(ctx *beegoCtx.Context) {
	return func(ctx *beegoCtx.Context) {
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
			"end_point":  ctx.Request.URL.Path,
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

func getUserAgent(ctx *beegoCtx.Context) string {
	if userAgent := ctx.Request.Header.Get("User-Agent"); userAgent != "" {
		return userAgent
	}

	return "Desconocido"
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
