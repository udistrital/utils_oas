package xray

import (
	"context"
	"fmt"
	"os"
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

func InitXRay(segmentName string) error {
	os.Setenv("AWS_XRAY_NOOP_ID", "false")
	XraySess, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		return err
	}

	xray.Configure(xray.Config{
		DaemonAddr: "127.0.0.1:2000", // Dirección y puerto del demonio de X-Ray local
		LogLevel:   "info",           // Nivel de log deseado
		LogFormat:  "json",           // Formato de log deseado (text o json)
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

	return nil
}

func BeginSegment(ctx *context2.Context) {
	ctx2 := ctx.Request.Context()
	ctx2, seg := BeginSegmentWithContextTP(ctx2, SegmentName, ctx.Request.Method, ctx.Request.URL.String(), 200, ctx.Request.URL.String(), ctx.Request.Header.Values("X-Amzn-Trace-Id"))
	defer seg.Close(nil)
	SetContext(ctx2)
	ctx.Input.SetData("XRaySegment", seg)
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
