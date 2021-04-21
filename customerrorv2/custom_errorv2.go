package customerrorv2

import (
	"github.com/astaxie/beego"
)

// UsuarioController operations for Usuario
type CustomErrorController struct {
	beego.Controller
}

func genericError(c *CustomErrorController, status string) {
	outputError := map[string]interface{}{"Success": false, "Status": status, "Message": c.Data["mesaage"], "Data": c.Data["data"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error400() {
	genericError(c, "400")
}

func (c *CustomErrorController) Error404() {
	genericError(c, "404")
}

func (c *CustomErrorController) Error500() {
	genericError(c, "500")
}

func (c *CustomErrorController) Error501() {
	genericError(c, "501")
}

func (c *CustomErrorController) Error502() {
	genericError(c, "502")
}

func (c *CustomErrorController) Error509() {
	genericError(c, "509")
}
