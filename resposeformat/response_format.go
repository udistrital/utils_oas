package resposeformat

import (
	"github.com/astaxie/beego"
)

type response struct {
	Code string
	Type string
	Body interface{}
}

// ResponseFormat ... set the status format for service's response.
func ResponseFormat(c *beego.Controller, data interface{}, code string, status int) {
	res := response{}
	c.Ctx.Output.SetStatus(status)
	if status >= 100 && status < 200 {
		res.Type = "information"
	} else if status == 200 && status < 300 {
		res.Type = "success"
	} else if status == 300 && status < 400 {
		res.Type = "redirection"
	} else if status == 404 {
		res.Type = "not found"
	} else {
		res.Type = "internal error"
	}

	res.Code = code
	res.Body = data

	c.Data["json"] = res
	c.ServeJSON()
}
