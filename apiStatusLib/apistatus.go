package apistatus

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
)

var checkCount uint

func clearCheck() {
	checkCount = 0
}

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
	clearCheck()
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
				clearCheck()
				logs.Critical(defaultErrorString, responseError)
				response = formatErrorResponse(responseError)
				ctx.Output.SetStatus(http.StatusServiceUnavailable) // 503
			}
			response["checkCount"] = checkCount
			ctx.Output.JSON(response, true, true)
			if checkCount == 0 {
				logs.Warn("APP_JUST_STARTED (please compare against the previous logged checkCount value to dismiss uint overflow)")
			}

			checkCount++
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
