package customerror

import (
	"github.com/astaxie/beego"
)

// UsuarioController operations for Usuario
type CustomErrorController struct {
	beego.Controller
}

func (c *CustomErrorController) Error400() {
	outputError := map[string]interface{}{"Status": "400", "Message": "The request contains incorrect syntax", "System": c.Data["system"], "Development": c.Data["development"]}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error404() {
	outputError := map[string]interface{}{"Status": "404", "Message": "Not found resource", "System": c.Data["system"], "Development": c.Data["development"]}

	c.Data["json"] = outputError
	c.ServeJSON()
}
