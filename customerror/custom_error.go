package customerror

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/auditoria"
	"github.com/udistrital/utils_oas/xray"
)

type CustomErrorController struct {
	beego.Controller
}

func (c *CustomErrorController) Error400() {
	outputError := map[string]any{"Status": "400", "Message": "The request contains incorrect syntax", "System": c.Data["system"], "Development": c.Data["development"]}
	xray.EndSegment(c.Ctx)
	auditoria.LogRequest(c.Ctx)
	c.Data["json"] = outputError
	c.ServeJSON()
}

func (c *CustomErrorController) Error404() {
	outputError := map[string]any{"Status": "404", "Message": "Not found resource", "System": c.Data["system"], "Development": c.Data["development"]}
	xray.EndSegment(c.Ctx)
	auditoria.LogRequest(c.Ctx)
	c.Data["json"] = outputError
	c.ServeJSON()
}
