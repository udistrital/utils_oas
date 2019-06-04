package auditoria

import (
	"fmt"
	"time"
	
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	//amqp "github.com/streadway/amqp"
)

/*Variables para la conexión y el canal
var connection *amqp.Connection
var chl *amqp.Channel

func failOnError(err error, msg string) {
	if err != nil {
		beego.Info("%s: %s", msg, err)
		beego.Info(fmt.Sprintf("%s: %s", msg, err))
	}
}
*/

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

    //Variable que contiene el nombre del API al que se le hace la petición
    app_name := beego.AppConfig.Strings("appname")
	//Host del API
	host := ctx.Request.Host
	//Variable que contiene el end point al que se le realiza la petición
	end_point := ctx.Request.URL.String()
    //Variable que contiene el método REST de la petición
	method := ctx.Request.Method
	//Variable que contiene la fecha y hora de la operación
	date := time.Now().String()

	
	//Variable que contiene la IP del usuario   <----- pendiente
	ip_user := "MyIP"
	//Variable que contiene el access token de la peticion  <--- Cuando viene de WSO2, se puede obtener. Cuando se realiza por postman, por ejemplo, no
	access_token := ctx.Request.Header["Authorization"][0]
	//Variable que define el tipo de aplicación, sistema operativo, provedor del software o laversión del software de la petición del agente de usuario
	user_agent := ctx.Request.Header["User-Agent"][0]
	//Variable que contiene el usuario    <----- pendiente
	user := "MyUser"

	//Variable que contiene el response body del servicio
	data_response := ctx.Input.Data()
		
	fmt.Println("Nombre API: " ,app_name)    
	fmt.Println("Host de la petición: " ,host)            
	fmt.Println("Endpoint ",  end_point) 
	fmt.Println("method " ,method)  
	fmt.Println("date  " ,date)  
	fmt.Println("ip_user" ,ip_user)  
//	fmt.Println("acces_token: " ,access_token)  
	fmt.Println("user_agent" ,user_agent)  
	                                            
	//fmt.Println(data_response["json"])                                                                      //En las peticiones get y post se ve la data, devuelve OK cuando se hace un post o un delete
	
	var log = fmt.Sprintf(`%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@&%s@$`, app_name, host,end_point,method,date,ip_user,access_token,user_agent,user,data_response["json"])
	beego.Info(log)
	

}

func FunctionFinishRouter(ctx *context.Context) {
	beego.Info("beego.FinishRouter: After finishing router")
}

func InitMiddleware() {
	beego.InsertFilter("*", beego.AfterExec, FunctionAfterExec, false)
}
