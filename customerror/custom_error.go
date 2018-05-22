package customerror

import (
	"github.com/astaxie/beego"
)

// UsuarioController operations for Usuario
type CustomErrorController struct {
	beego.Controller
}

func (c *CustomErrorController) Error400() {
	outputError := map[string]string{"Code": "400", "Body": "The request contains incorrect syntax", "Type": "error"}
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error404() {
	outputError := map[string]string{"Code": "404", "Body": "Not found resource", "Type": "error"}
	c.Data["json"] = outputError
	c.ServeJSON()
}
