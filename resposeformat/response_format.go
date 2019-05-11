package resposeformat

import (
	"github.com/astaxie/beego"
)

type response struct {
	Code   string
	Status string
	Body   interface{}
}

// SetResponseFormat ... set the status format for service's response.
func SetResponseFormat(c *beego.Controller, data interface{}, code string, status int) {
	res := response{}
	c.Ctx.Output.SetStatus(status)
	if status >= 100 && status < 200 {
		res.Status = "information"
	} else if status == 200 && status < 300 {
		res.Status = "success"
	} else if status == 300 && status < 400 {
		res.Status = "redirection"
	} else if status == 404 {
		res.Status = "not found"
	} else {
		res.Status = "error"
	}

	res.Code = code
	res.Body = data

	c.Data["json"] = res
	c.ServeJSON()
}
