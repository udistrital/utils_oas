package errorhandler

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/requestresponse"
)

// Agregar en main o router beego.ErrorController(& ruta de utils_oas errorhandler.ErrorHandlerController{})
type ErrorHandlerController struct {
	beego.Controller
}

// Captura de error cuando se consulta a endpoint inexistente
func (c *ErrorHandlerController) Error404() {
	metodo := c.Ctx.Request.Method
	ruta := c.Ctx.Request.URL.Path
	statusCode := http.StatusNotFound
	message := fmt.Sprintf("nomatch|%s|%s", metodo, ruta)
	c.Ctx.Output.SetStatus(statusCode)
	c.Data["json"] = requestresponse.APIResponseDTO(false, statusCode, nil, message)
	c.ServeJSON()
}

// Captura de error cuando Mid entra en pánico, se debe colocar al inicio de cada controlador: defer HandlePanic(&c.Controller)
// - Por consola indica donde estuvo el fallo
// - Formatea respuesta cuando mid falla enviando un Internal Server Error.
func HandlePanic(c *beego.Controller) {
	if r := recover(); r != nil {
		logs.Error("Panic: ", r)
		debug.PrintStack()
		message := fmt.Sprintf("Error service %s: An internal server error occurred.", beego.AppConfig.String("appname"))
		message += fmt.Sprintf(" Request Info: URL: %s, Method: %s", c.Ctx.Request.URL, c.Ctx.Request.Method)
		message += " Time: " + time.Now().Format(time.RFC3339)
		statusCode := http.StatusInternalServerError
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = requestresponse.APIResponseDTO(false, statusCode, nil, message)
		c.ServeJSON()
	}
}
