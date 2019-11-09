package responseformat

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/lib/pq"

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

		out = formatResponseObject(fmt.Sprintf("%s", r), "", status)
		return
	}
	if reflect.ValueOf(Body).IsValid() {
		if _, e := Body.([]interface{}); e {
			if len(Body.([]interface{})) == 0 {
				Body = make(map[string]interface{})
			}
		}
		status = 200
		switch Body.(type) {
		case *json.UnmarshalTypeError, *json.UnmarshalFieldError, *pq.Error, *mgo.LastError:
			status = 500
		case string:
			Body, status = stringBeegoErrorCatch(Body.(string))
		}
		out = formatResponseObject(Body, "", status)
		return
	}

	beego.Error(Body)
	status = 500
	out = formatResponseObject(Body, "", status)
}

// CheckResponseError ... return true if response format type is an error.
func CheckResponseError(response Response) bool {
	if response.Type == "error" {
		return true
	}
	return false
}

func stringBeegoErrorCatch(err string) (body interface{}, status int) {
	if strings.Contains(err, "json:") || strings.Contains(err, "Error:") || strings.Contains(err, "wrong field/column name") || strings.Contains(err, "pq:") || strings.Contains(err, "unknown field/column name") {
		return err, 500
	}

	switch err {
	case "<QuerySeter> no row found":
		data := make(map[string]interface{})
		return data, 200
	}

	return err, 200
}
