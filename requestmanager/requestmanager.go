package requestmanager

import (
	"encoding/json"

	"github.com/udistrital/utils_oas/formatdata"
	"github.com/udistrital/utils_oas/responseformat"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

// FillRequestWithPanic ... unmarshal body request to an interface . panic if some error hapen. Only objects no Arrays.
func FillRequestWithPanic(c *beego.Controller, output interface{}) {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, output); err != nil {
		logs.Error(err.Error())
		panic(err.Error())
	}

	errMes := formatdata.StructValidation(output)
	if errMes != nil {
		responseformat.SetResponseFormat(c, errMes, "", 422)
	}
}
