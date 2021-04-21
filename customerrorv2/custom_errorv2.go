package customerrorv2

import (
	"github.com/astaxie/beego"
)

// UsuarioController operations for Usuario
type CustomErrorController struct {
	beego.Controller
}

func (c *CustomErrorController) Error400() {
	outputError := map[string]interface{}{"Success": false, "Status": "400", "Message": c.Data["mesaage"], "Data": c.Data["data"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error404() {
	outputError := map[string]interface{}{"Success": false, "Status": "404", "Message": c.Data["mesaage"], "Data": c.Data["data"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error500() {
	outputError := map[string]interface{}{"Success": false, "Status": "500", "Message": c.Data["mesaage"], "Data": c.Data["data"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error502() {
	outputError := map[string]interface{}{"Success": false, "Status": "502", "Message": c.Data["mesaage"], "Data": c.Data["data"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}
