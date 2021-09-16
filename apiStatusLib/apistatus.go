package apistatus

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
)

func statusResponse(status string) map[string]interface{} {
	return map[string]interface{}{
		"Status": status,
	}
}

var defaultStatusResponse map[string]interface{} = statusResponse("Ok")

const defaultErrorString string = "UNHEALTHY_STATE"

func formatErrorResponse(errorMsg interface{}) map[string]interface{} {
	return statusResponse(fmt.Sprintf("%s -  %v", defaultErrorString, errorMsg))
}

// InitWithHandler accepts a (handler) function that, once performs the
// healthcheck, returns "nil" when everything is OK.
func InitWithHandler(statusCheckHandler func() (statusCheckError interface{})) {
	beego.Any("/", func(ctx *context.Context) {
		var responseError interface{}

		defer func() {

			// "catch"
			if err := recover(); err != nil {
				responseError = err
			}

			// "finally"
			response := defaultStatusResponse
			if responseError != nil {
				logs.Critical(defaultErrorString, responseError)
				response = formatErrorResponse(responseError)
				ctx.Output.SetStatus(http.StatusServiceUnavailable) // 503
			}
			ctx.Output.JSON(response, true, true)
		}()

		// "try"
		if statusCheckHandler != nil {
			if err := statusCheckHandler(); err != nil {
				responseError = err
			}
		}
	})
}

func Init() {
	InitWithHandler(nil)
}
