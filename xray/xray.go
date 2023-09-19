package xray

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	context2 "github.com/astaxie/beego/context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var GlobalContext context.Context
var SegmentName string
var StatusCode int
var Seg *xray.Segment
var URL string
var Method string
var Controller string

func InitXRay() error {
	os.Setenv("AWS_XRAY_NOOP_ID", "true")
	os.Setenv("AWS_XRAY_DEBUG_MODE", "TRUE")
	XraySess, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		return err
	}
	xray.Configure(xray.Config{
		//DaemonAddr: "ec2-54-162-219-111.compute-1.amazonaws.com:2000", // Dirección y puerto del demonio de X-Ray local
		DaemonAddr: "127.0.0.1:2000",
		LogLevel:   "debug", // Nivel de log deseado
		LogFormat:  "json",  // Formato de log deseado (text o json)
	})

	// S3 and ECS Clients
	ecrClient := ecr.New(XraySess)
	ecsClient := ecs.New(XraySess)

	// XRay Setup
	xray.AWS(ecrClient.Client)
	xray.AWS(ecsClient.Client)

	fmt.Println("Listed buckets successfully")
	beego.InsertFilter("*", beego.BeforeExec, BeginSegment)
	beego.InsertFilter("*", beego.AfterExec, EndSegment, false)
	return nil
}

func BeginSegment(ctx *context2.Context) {
	fmt.Println("CONTEXTO ENTRANTE: ", ctx.Request.Context())
	SegmentName = ctx.Input.Context.Request.Host
	URL = "http://" + SegmentName + ctx.Input.Context.Request.URL.String()
	Method = ctx.Request.Method
	ctx3, seg := BeginSegmentWithContextTP(ctx.Request.Context(), StatusCode, ctx.Request.Header.Values("X-Amzn-Trace-Id"))
	Seg = seg
	GlobalContext = ctx3
}

func BeginSubsegment(subsegment, method, URL string, status int) (context.Context, *xray.Segment) {

	globalContext, subseg := xray.BeginSubsegment(GlobalContext, subsegment)

	subseg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: method,
			URL:    URL,
		},
		Response: &xray.ResponseData{
			Status: status,
		},
	}
	return globalContext, subseg

}

func EndSegment(ctx *context2.Context) {
	// Obtener el valor de la clave "json" del contexto
	jsonValue := ctx.Input.GetData("json")
	// Convertir el valor a un mapa
	if jsonMap, ok := jsonValue.(map[string]interface{}); ok {
		// Obtener el valor de la clave "Status" del mapa
		status, ok := jsonMap["Status"].(string)
		if ok {
			num, err := strconv.Atoi(status)
			if err == nil {
				Seg.HTTP = &xray.HTTPData{
					Request: &xray.RequestData{
						Method: Method,
						URL:    URL,
					},
					Response: &xray.ResponseData{
						Status: num,
					},
				}
			}

		}
	}
	fmt.Println("----CONTEXT XRAY 5----:", xray.GetSegment(GlobalContext).Name)
	fmt.Println("----GLOBAL CONTEXT----:", GlobalContext)
	fmt.Println("----GLOBAL CONTEXT back----:", context.Background())
	fmt.Println("SEG NAME PRINCI: ", Seg.Name)
	Seg.Close(nil)
}

func UpdateState(status int, err error) {
	StatusCode = status
	Seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: Method,
			URL:    URL,
		},
		Response: &xray.ResponseData{
			Status: StatusCode,
		},
	}
	fmt.Println("----CONTEXT XRAY 4----:", xray.GetSegment(GlobalContext).Name)
	if status == 500 || status == 501 || status == 502 || status == 503 {
		Seg.AddError(fmt.Errorf("%v", err))
		Seg.Close(nil)
	}
}

func EvaluateState(status int) {
	if StatusCode != 500 && StatusCode != 501 && StatusCode != 502 && StatusCode != 503 {
		StatusCode = status
	}
}

func ErrorController5xx(err error) {
	StatusCode = 500
	Seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: Method,
			URL:    URL,
		},
		Response: &xray.ResponseData{
			Status: StatusCode,
		},
	}
	Seg.AddMetadata("Error", err)
	Seg.AddError(fmt.Errorf("%v", err))
	fmt.Println("----CONTEXT XRAY 3----:", xray.GetSegment(GlobalContext).Name)
	Seg.Close(nil)
	fmt.Println("context DONE D: ", Seg.ContextDone)
	fmt.Println("context DONE seg D2:", Seg.Emitted)
}

func EndSegmentErr(err interface{}) {
	Seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: Method,
			URL:    URL,
		},
		Response: &xray.ResponseData{
			Status: StatusCode,
		},
	}
	Seg.AddMetadata("Error", err)
	fmt.Println("----CONTEXT XRAY 2----:", xray.GetSegment(GlobalContext).Name)
	Seg.Close(nil)
}

func BeginSegmentWithContextTP(ctx context.Context, code int, traceID []string) (context.Context, *xray.Segment) {

	ctx, seg := xray.BeginSegment(ctx, SegmentName)

	seg.Origin = URL

	seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: Method,
			URL:    URL,
		},
		Response: &xray.ResponseData{
			Status: code,
		},
	}
	//fmt.Println("----CONTEXT XRAY 1----:", xray.GetSegment(globalContext).Name)
	if traceID != nil {
		traceID := strings.Trim(traceID[0], "[]")
		id, parent := GetTraceIDAndParentID(traceID)
		seg.TraceID = id
		seg.ParentID = parent
	}
	return ctx, seg
}

func GetTraceIDAndParentID(traceID string) (trace string, parent string) {
	if traceID != "" {
		traceIDParts := strings.Split(traceID, ";")
		Id := ""
		IdParent := ""
		for _, part := range traceIDParts {
			if strings.HasPrefix(part, "Root=") {
				Id = strings.TrimPrefix(part, "Root=")
			}
			if strings.HasPrefix(part, "Parent=") {
				IdParent = strings.TrimPrefix(part, "Parent=")
			}
		}
		return Id, IdParent
	} else {
		return
	}
}

func GetStatusCode() int {
	return StatusCode
}

func SetStatusCode(statusCode int) {
	StatusCode = statusCode
}
