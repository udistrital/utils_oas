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

var globalContext context.Context
var SegmentName string
var StatusCode int
var Seg *xray.Segment
var Controller string

func InitXRay(segmentName string) error {
	os.Setenv("AWS_XRAY_NOOP_ID", "false")
	XraySess, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		return err
	}

	xray.Configure(xray.Config{
		DaemonAddr: "ec2-54-162-219-111.compute-1.amazonaws.com:2000", // Dirección y puerto del demonio de X-Ray local
		//DaemonAddr: "127.0.0.1:2000",
		LogLevel:  "info", // Nivel de log deseado
		LogFormat: "json", // Formato de log deseado (text o json)
	})

	// S3 and ECS Clients
	ecrClient := ecr.New(XraySess)
	ecsClient := ecs.New(XraySess)

	// XRay Setup
	xray.AWS(ecrClient.Client)
	xray.AWS(ecsClient.Client)

	fmt.Println("Listed buckets successfully")
	SegmentName = segmentName
	beego.InsertFilter("*", beego.BeforeExec, BeginSegment)
	beego.InsertFilter("*", beego.AfterExec, EndSegment, false)
	return nil
}

func BeginSegment(ctx *context2.Context) {
	ctx2 := ctx.Request.Context()
	ctx2, seg := BeginSegmentWithContextTP(ctx2, SegmentName, ctx.Request.Method, ctx.Request.URL.String(), StatusCode, ctx.Request.URL.String(), ctx.Request.Header.Values("X-Amzn-Trace-Id"))
	Seg = seg
	SetContext(ctx2)
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
						Method: ctx.Request.Method,
						URL:    ctx.Request.URL.String(),
					},
					Response: &xray.ResponseData{
						Status: num,
					},
				}
			}

		}
	}
	Seg.Close(nil)
}

func UpdateState(Method, URL string, status int, err error) {
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

func ErrorController5xx(method, url string, err error) {
	StatusCode = 500
	Seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: method,
			URL:    url,
		},
		Response: &xray.ResponseData{
			Status: StatusCode,
		},
	}
	Seg.AddError(fmt.Errorf("%v", err))
	Seg.AddMetadata("Error", err)
	Seg.Close(nil)
}

func EndSegmentErr(Method, URL string, err interface{}) {
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
	Seg.Close(nil)
}

func BeginSegmentWithContextTP(ctx context.Context, segmentName string, method string, url string, code int, origin string, traceID []string) (context.Context, *xray.Segment) {
	if traceID != nil {
		traceID := strings.Trim(traceID[0], "[]")
		id, parent := GetTraceIDAndParentID(traceID)
		ctx, seg := xray.BeginSegment(ctx, segmentName)
		seg.Origin = origin
		seg.HTTP = &xray.HTTPData{
			Request: &xray.RequestData{
				Method: method,
				URL:    url,
			},
			Response: &xray.ResponseData{
				Status: code,
			},
		}
		seg.TraceID = id
		seg.ParentID = parent

		return ctx, seg
	} else {
		ctx, seg := xray.BeginSegment(ctx, segmentName)
		seg.Origin = origin
		seg.HTTP = &xray.HTTPData{
			Request: &xray.RequestData{
				Method: method,
				URL:    url,
			},
			Response: &xray.ResponseData{
				Status: code,
			},
		}
		return ctx, seg
	}

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

func GetContext() context.Context {
	return globalContext
}

func SetContext(ctx context.Context) {
	globalContext = ctx
}

func GetStatusCode() int {
	return StatusCode
}

func SetStatusCode(statusCode int) {
	StatusCode = statusCode
}
