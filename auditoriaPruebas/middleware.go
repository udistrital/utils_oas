package auditoria

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/patrickmn/go-cache"
	"encoding/json"
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

/*func ListenRequest(ctx *context.Context) {

	/*---- Declaración de variables ---- 

	/*---- Información relacionada a la aplicación y la petición ---- 
	var app_name string  //Nombre del API al que se le hace la petición
	var host string      //Host del API
	var end_point string //End point al que se le realiza la petición
	var method string    //Método REST de la petición
	var date string      //Fecha y hora de la operación

	/*---- Información relacionada con el usuario ---- 
	var ip_user string    //IP del usuario   <----- pendiente
	var user_agent string //Tipo de aplicación, sistema operativo, provedor del software o laversión del software de la petición del agente de usuario
	var user string       //Nombre de usuario en WSO2 que realiza la petición    <----- pendiente
	// var access_token string //Access token asignado al usuario que realiza peticion

	/*---- Información relacionada con el cuerpo de la petición ---- 
	var data_response interface{} //Payload del servicio

	/*---- Asignación de variables ----
	app_name = beego.AppConfig.String("appname")
	host = ctx.Request.Host
	end_point = ctx.Request.URL.String()
	method = ctx.Request.Method
	date = time.Now().String()
	ip_user = ctx.Input.IP()
	if len(ctx.Request.Header["User-Agent"]) > 0 {
		user_agent = ctx.Request.Header["User-Agent"][0]
	}
	data_response = ctx.Input.Data()
	//data_response = "ejemplo"

	// *--------- Se implementa try y catch para cuando la petición NO viene de WSO2 y no se tiene access_token
	//
	go func() {
		defer func() {

			//Catch
			if r := recover(); r != nil {

				// access_token = "NO WSO2"
				user = "NO WSO2 - No user"
				var log = fmt.Sprintf(`@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@$`, app_name, host, end_point, method, date, ip_user, user_agent, user, data_response)
				if end_point != "/" {
					beego.Info(log)
				}

			}
		}()

		// try
		// access_token = ctx.Request.Header["Authorization"][0]

		/*---- Obtención del usuario ---- 
		defer func() {
			if r := recover(); r != nil {
				// access_token = "Error WSO2"
				user = "Error wso2"
				var log = fmt.Sprintf(`@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@$`, app_name, host, end_point, method, date, ip_user, user_agent, user, data_response)
				if end_point != "/" {
					beego.Info(log)
				}
			}
		}()

		user = getUserInfo(ctx)
		var log = fmt.Sprintf(`@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@$`, app_name, host, end_point, method, date, ip_user, user_agent, user, data_response)
		if end_point != "/" {
			beego.Info(log)
		}
	}()
}*/


func ListenRequest(ctx *context.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logData := map[string]interface{}{
					"app_name":   beego.AppConfig.String("appname"),
					"host":       ctx.Request.Host,
					"end_point":  ctx.Request.URL.String(),
					"method":     ctx.Request.Method,
					"date":       time.Now().Format(time.RFC3339),
					"ip_user":    ctx.Input.IP(),
					"user_agent": getUserAgent(ctx),
					"user":       "Error WSO2 - Sin usuario"
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
			"ip_user":    ctx.Input.IP(),
			"user_agent": getUserAgent(ctx),
			"user":       user,
			"data":       sanitizeInputData(ctx.Input.Data()),
		}

		logAsJSON(logData)
	}()
}

func logAsJSON(data map[string]interface{}) {
	jsonData, err := json.Marshal(data["data"])
	if err != nil {
		beego.Error("Error al serializar el campo 'data' a JSON:", err)
		jsonData = []byte("{}")
	}  else {
		var pruebaLog = "{app_name: " + data["app_name"].(string) + 
			", host: " + data["host"].(string) +
			", end_point: " + data["end_point"].(string) +
			", method: " + data["method"].(string) +
			", date: " + data["date"].(string) +
			", ip_user: " + data["ip_user"].(string) +
			", user_agent: " + data["user_agent"].(string) +
			", user: " + data["user"].(string) +
			", data: " + string(jsonData) +
		"}"
		var pruebaLogLog = fmt.Sprintf(`%s`, pruebaLog)
		beego.Info(pruebaLogLog)
	}
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
	beego.InsertFilter("*", beego.AfterExec, ListenRequest, false)
}
