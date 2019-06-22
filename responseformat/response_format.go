package responseformat

import (
	"reflect"

	"github.com/astaxie/beego/context"

	"github.com/astaxie/beego"
)

type response struct {
	Code string
	Type string
	Body interface{}
}

// SetResponseFormat ... set the status format for service's response.
func SetResponseFormat(c *beego.Controller, data interface{}, code string, status int) {
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
		res.Type = "error"
	}

	res.Code = code
	res.Body = data

	c.Data["json"] = res
	c.ServeJSON()
}

func ModifyBeegoDefaultResponseFormat(ctx *context.Context, data interface{}, status int) {
	res := response{}
	ctx.Output.SetStatus(status)
	if status >= 100 && status < 200 {
		res.Type = "information"
	} else if status == 200 && status < 300 {
		res.Type = "success"
	} else if status == 300 && status < 400 {
		res.Type = "redirection"
	} else if status == 404 {
		res.Type = "not found"
	} else {
		res.Type = "error"
	}

	res.Body = data

	ctx.Output.JSON(res, false, false)
}

// GlobalResponseHandler ... Global defer for any go panic at the API.
func GlobalResponseHandler(ctx *context.Context) {
	out := response{}
	Body := ctx.Input.Data()["json"]

	defer func() {
		ctx.Output.JSON(out, true, false)

	}()

	if r := recover(); r != nil {
		beego.Error(r)
		ctx.ResponseWriter.WriteHeader(500)
		out.Body = r
		out.Code = ""
		out.Type = "error"
	} else {
		if reflect.ValueOf(Body).IsValid() {

			defer func() {
				if r := recover(); r != nil {
					// beego.Error(r)
					out.Body = Body
					out.Type = "success"
					ctx.ResponseWriter.WriteHeader(200)
				}
			}()

			if reflect.ValueOf(Body).IsNil() {
				out.Body = nil
				out.Type = "No Data Found"
				ctx.ResponseWriter.WriteHeader(201)
			} else {
				out.Body = Body
				out.Type = "success"
				ctx.ResponseWriter.WriteHeader(200)
			}

		} else {
			beego.Error("Unknow error")
			ctx.ResponseWriter.WriteHeader(500)
			out.Body = "Unknow error"
			out.Code = ""
			out.Type = "error"
		}
	}
}
