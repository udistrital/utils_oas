package errorctrl

import (
	"net/http"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// Error control for controller
func ErrorControlController(c beego.Controller, controller string) {
	if err := recover(); err != nil {
		logs.Error(err)
		localError := err.(map[string]interface{})
		c.Data["message"] = (beego.AppConfig.String("appname") + "/" + controller + "/" + (localError["funcion"]).(string))
		c.Data["data"] = (localError["err"])
		if status, ok := localError["status"]; ok && status != nil {
			if statusCode, ok := status.(int); ok {
				c.Ctx.Output.SetStatus(statusCode)
			} else {
				c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			}
		} else {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		}
		c.Data["json"] = map[string]interface{}{
			"Data":    localError["err"],
			"Message": c.Data["message"],
			"Status":  strconv.Itoa(c.Ctx.Output.Status),
			"Success": false,
		}
		c.ServeJSON()
	}
}

// Error control for functions
func ErrorControlFunction(funcion string, status string) {
	if err := recover(); err != nil {
		panic(Error(funcion, err, status))
	}
}

// Get a error with standard struct
func Error(funcion string, err interface{}, status string) (outputError map[string]interface{}) {
	switch localError := err.(type) {
	case map[string]interface{}:
		if fun, ok := localError["funcion"]; ok && fun != nil {
			funcion = funcion + "/" + fun.(string)
		}
		if er, ok := localError["err"]; ok && er != nil {
			err = er
		}
	}
	outputError = map[string]interface{}{"funcion": funcion, "err": err, "status": status}
	return outputError
}
