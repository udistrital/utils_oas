package customerror

import (
	"github.com/astaxie/beego"
)

// UsuarioController operations for Usuario
type CustomErrorController struct {
	beego.Controller
}

func (c *CustomErrorController) Error400() {
	outputError := map[string]interface{}{"status": "400", "message": "The request contains incorrect syntax", "type": "error", "development": c.Data["development"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error404() {
	outputError := map[string]interface{}{"status": "400", "message": "Not found resource", "type": "error", "development": c.Data["development"]}

	c.Data["json"] = outputError
	c.ServeJSON()
}
