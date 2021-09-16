package apistatus

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func statusResponse(status string) map[string]interface{} {
	return map[string]interface{}{
		"Status": status,
	}
}

var defaultStatusResponse map[string]interface{} = statusResponse("Ok")

func formatErrorResponse(errorMsg interface{}) map[string]interface{} {
	return statusResponse(fmt.Sprintf("ERROR: %v", errorMsg))
}

// InitWithHandler accepts a (handler) function that, once performs the
// healthcheck, returns "nil" when everything is OK.
func InitWithHandler(statusCheckHandler func() (statusCheckError interface{})) {
	beego.Any("/", func(ctx *context.Context) {
		var responseError interface{}

		// "catch"
		defer func() {
			if err := recover(); err != nil {
				responseError = err
			}

			response := defaultStatusResponse
			if responseError != nil {
				response = formatErrorResponse(responseError)
			}

			// "finally"
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
