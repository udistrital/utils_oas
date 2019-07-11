package responseformat

import (
	"reflect"

	"github.com/astaxie/beego/context"

	"github.com/astaxie/beego"
)

// Response struct ... Response format JSON
type Response struct {
	Code string
	Type string
	Body interface{}
}

// formatResponseObject ... format to response structure.
func formatResponseObject(data interface{}, code string, status int) Response {
	res := Response{}

	if status >= 200 && status < 300 {
		res.Type = "success"
	} else {
		res.Type = "error"
	}

	res.Code = code
	res.Body = data
	return res
}

// SetResponseFormat ... set the status format for service's response.
func SetResponseFormat(c *beego.Controller, data interface{}, code string, status int) {
	c.Ctx.Output.SetStatus(status)

	res := formatResponseObject(data, code, status)
	c.Data["json"] = res
	c.ServeJSON()
}

// GlobalResponseHandler ... Global defer for any go panic in the Beego API.
func GlobalResponseHandler(ctx *context.Context) {
	var out interface{}
	var status int
	Body := ctx.Input.Data()["json"]

	defer func() {
		ctx.ResponseWriter.WriteHeader(status)
		ctx.Output.JSON(out, true, false)

	}()

	if r := recover(); r != nil {
		beego.Error(r)
		status = 500

		out = formatResponseObject(r, "", status)
	} else {
		if reflect.ValueOf(Body).IsValid() {

			status = 200
			out = formatResponseObject(Body, "", status)

		} else {
			beego.Error("Unknow error")
			status = 500
			out = formatResponseObject("Unknow error", "", status)
		}
	}
}

// CheckResponseError ... return true if response format type is an error.
func CheckResponseError(response Response) bool {
	if response.Type == "error" {
		return true
	}
	return false
}
