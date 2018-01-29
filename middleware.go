package auditoria

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	amqp "github.com/streadway/amqp"
)

//Variables para la conexión y el canal
var connection *amqp.Connection
var chl *amqp.Channel

func failOnError(err error, msg string) {
	if err != nil {
		beego.Info("%s: %s", msg, err)
		beego.Info(fmt.Sprintf("%s: %s", msg, err))
	}
}

func FunctionBeforeStatic(ctx *context.Context) {
	beego.Info("beego.BeforeStatic: Before finding the static file")
}
func FunctionBeforeRouter(ctx *context.Context) {
	beego.Info("beego.BeforeRouter: Executing Before finding router")
}
func FunctionBeforeExec(ctx *context.Context) {

	beego.Info("beego.BeforeExec: After finding router and before executing the matched Controller")
}

func FunctionAfterExec(ctx *context.Context) {
	//Variable que contiene la hora de la operación
	now := time.Now().String()
	//Variable que contiene la IP del usuario
	ip_user := ctx.Input.IP()
	//Variable que contiene el servicio al que se le hace la petición
	url := ctx.Request.URL.String()
	//Variable que contiene el método de la petición
	metodo := ctx.Request.Method
	//Host del API
	host := ctx.Request.Host
	//Variable que contiene el cuerpo del JSON que el usuario envia
	data_user := string(ctx.Input.RequestBody)
	//Variable que contiene el response body del servicio
	data_response := ctx.Input.Data()
	//Variable que contiene el nombre del API al que se le hace la petición
	app := beego.AppConfig.Strings("appname")

	/*fmt.Println("Nombre API: " + app[0])              //Nombre del API al que se le hace la petición
	fmt.Println("La fecha de la petición es: " + now) //Fecha de transacción
	fmt.Println("Este es el query ", ctx.Request.URL.Query().Get("auth"))---> Usuario quien hace la petición WSO2
	fmt.Println("Este es la IP del usuario que hace la petición: " + ctx.Input.IP())
	fmt.Println("Este es la URL del servicio a la que se le hace la petición: " + ctx.Request.URL.String()) //URL de la petición
	fmt.Println("Este es el método de la petición: " + ctx.Request.Method)                                  //Método de la petición
	fmt.Println("Este es el host del api: " + ctx.Request.Host)                                             //Host desde el que se hace la petición
	fmt.Println("Data enviada por el usuario:" + data_user)                                                 //Data enviada por el usuario
	fmt.Println(data_response["json"])                                                                      //En las peticiones get y post se ve la data, devuelve OK cuando se hace un post o un delete
	*/

	var mensaje = fmt.Sprintf(`{"FechaOperacion": %s, "User": "userWSO2", "IpUser": %s, "UrlService": %s, "Método": %s, "HostApi": %s ,"DataUser":%s, "DataResponse":%s, "ApiName":%s}`, now, ip_user, url, metodo, host, data_user, data_response["json"], app[0])

	sentToRabbit(mensaje)

	beego.Info("Petición auditada")

}

func FunctionFinishRouter(ctx *context.Context) {
	beego.Info("beego.FinishRouter: After finishing router")
}

func InitMiddleware() {
	beego.InsertFilter("*", beego.AfterExec, FunctionAfterExec, false)
}
