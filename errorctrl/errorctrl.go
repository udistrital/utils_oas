package errorctrl

import (
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// ErrorControlController recovers structured errors raised by Error.
func ErrorControlController(c beego.Controller, controller string) {
	if err := recover(); err != nil {
		logs.Error(err)

		appname, _ := beego.AppConfig.String("appname")
		message := appname + "/" + controller
		data := normalizeError(err)
		statusCode := http.StatusInternalServerError

		switch localError := err.(type) {
		case map[string]any:
			if funcion, ok := localError["funcion"].(string); ok {
				message += "/" + funcion
			}

			if localData, ok := localError["err"]; ok {
				data = normalizeError(localData)
			}

			switch status := localError["status"].(type) {
			case int:
				statusCode = status
			case string:
				if parsedStatus, err := strconv.Atoi(status); err == nil {
					statusCode = parsedStatus
				}
			}
		case string:
			data = localError
		case error:
			data = localError.Error()
		}

		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]any{
			"Data":    data,
			"Message": message,
			"Status":  strconv.Itoa(statusCode),
			"Success": false,
		}

		if err := c.ServeJSON(); err != nil {
			logs.Error("error al serializar la respuesta de error: %v", err)
		}
	}
}

// ErrorControlFunction recovers an error and repanics it using the standard structure.
func ErrorControlFunction(funcion string, status string) {
	if err := recover(); err != nil {
		panic(Error(funcion, err, status))
	}
}

// Error returns an error using the standard structure.
func Error(funcion string, err any, status string) (outputError map[string]any) {
	switch localError := err.(type) {
	case map[string]any:
		if fun, ok := localError["funcion"].(string); ok {
			funcion += "/" + fun
		}
		if internalError, ok := localError["err"]; ok {
			err = internalError
		}
	case string:
		err = localError
	case error:
		err = localError.Error()
	}

	return map[string]any{
		"funcion": funcion,
		"err":     normalizeError(err),
		"status":  status,
	}
}

func normalizeError(err any) any {
	switch localError := err.(type) {
	case error:
		return localError.Error()
	case string:
		return localError
	}

	return err
}
