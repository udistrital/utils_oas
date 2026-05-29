package xray

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-xray-sdk-go/v2/header"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/aws/aws-xray-sdk-go/v2/xraylog"
	"github.com/udistrital/utils_oas/ssm"
)

const (
	segmentKey string = "xray_seg"
	urlKey     string = "xray_url"
	methodKey  string = "xray_method"
	traceIDKey string = "X-Amzn-Trace-Id"
)

var appName = beego.AppConfig.String("appname")
var globalCtx context.Context
var globalSeg *xray.Segment
var statusCode int
var url string
var method string

func Init() {
	if err := InitXRay(); err != nil {
		logs.Error(err.Error())
	}
}

// Deprecated. Use Init instead. This function is kept for backward compatibility
func InitXRay() error {
	if err := configureXRay(); err != nil {
		return err
	}

	beego.InsertFilter("/:version/*", beego.BeforeExec, beginSegment)
	beego.InsertFilter("/:version/*", beego.AfterExec, endSegment, false)
	return nil
}

// configureXRay performs the core X-Ray initialization and configuration.
// Returns an error if configuration fails, or nil if X-Ray is not configured or succeeds.
func configureXRay() error {
	parameterStore := beego.AppConfig.String("parameterStore")
	if parameterStore == "" {
		return fmt.Errorf("no se puede consultar daemon address: %v", errors.New("parameterStore no configurado"))
	}

	daemonAddr, err := ssm.GetParameterFromParameterStore(context.Background(), "/"+parameterStore+"/utils/xray/DaemonAddr")
	if err != nil {
		return fmt.Errorf("error consultando daemon address: %v", err)
	}

	config := xray.Config{DaemonAddr: daemonAddr}
	if err := xray.Configure(config); err != nil {
		return fmt.Errorf("error configurando xray: %v", err)
	}

	xray.SetLogger(xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelInfo))
	logs.Info("X-Ray inicializado correctamente")

	return nil
}

// Función que Crea el segmento principal asociado a la API, tomando en cuenta si
// es la API principal (a la cual se realizó la petición inicial) o una secundaria.
//
// Parámetros:
// - ctx: objeto context de Beego
//
// Variables:
// - URL: URL de la petición.
// - method: Método de la petición.
// - globalCtx: Inicialización de un contexto vacío para almacenar los segmentos que se generen.
// - globalSeg: Segmento principal de la Traza.
func beginSegment(ctx *beegoCtx.Context) {
	host := ctx.Input.Context.Request.Host
	if strings.HasPrefix(host, "localhost") {
		globalSeg = nil
		return
	}

	reqURL := ctx.Input.Scheme() + "://" + host + ctx.Input.Context.Request.URL.String()
	reqMethod := ctx.Request.Method
	env := ""
	if strings.HasPrefix(host, "pruebas") {
		env = "_test"
	}

	reqCtx, reqSeg := xray.BeginSegment(ctx.Request.Context(), appName+env)
	reqSeg.HTTP = &xray.HTTPData{
		Request:  &xray.RequestData{Method: reqMethod, URL: reqURL},
		Response: &xray.ResponseData{Status: 0},
	}

	traceID := ctx.Request.Header.Values("X-Amzn-Trace-Id")
	if traceID != nil {
		h := header.FromString(strings.Trim(traceID[0], "[]"))
		reqSeg.TraceID = h.TraceID
		reqSeg.ParentID = h.ParentID
		ctx.ResponseWriter.Header().Set("Resp-X-Amzn-Trace-Id", "true")
	}

	reqCtx = context.WithValue(reqCtx, segmentKey, reqSeg)
	reqCtx = context.WithValue(reqCtx, urlKey, reqURL)
	reqCtx = context.WithValue(reqCtx, methodKey, reqMethod)
	ctx.Request = ctx.Request.WithContext(reqCtx)

	url = reqURL
	method = reqMethod
	globalCtx = reqCtx
	globalSeg = reqSeg
}

// Actualiza y cierra el segmento principal y envía los datos del segmento y la traza a AWS X-Ray.
//
// Parámetros:
// - ctx: puntero a objeto context de Beego
func endSegment(ctx *beegoCtx.Context) {
	seg, ok := ctx.Request.Context().Value(segmentKey).(*xray.Segment)
	if !ok || seg == nil {
		return
	}

	url, _ := ctx.Request.Context().Value(urlKey).(string)
	method, _ := ctx.Request.Context().Value(methodKey).(string)

	status := ctx.ResponseWriter.Status
	if jsonMap, ok := ctx.Input.GetData("json").(map[string]interface{}); ok {
		if s, ok := jsonMap["Status"].(string); ok {
			if num, err := strconv.Atoi(s); err == nil {
				status = num
			}
		}
	}

	if status == 0 {
		status = http.StatusOK
	}

	seg.HTTP = &xray.HTTPData{
		Request:  &xray.RequestData{Method: method, URL: url},
		Response: &xray.ResponseData{Status: status},
	}
	if status >= http.StatusInternalServerError {
		_ = seg.AddError(fmt.Errorf("response status %d", status))
	}
	seg.Close(nil)
}

// BeginSubsegment starts an X-Ray subsegment for an outgoing HTTP request.
// It sets the trace propagation header on req and returns the updated context and subsegment
func BeginSubsegment(ctx context.Context, req *http.Request) (context.Context, *xray.Segment) {
	if xray.GetSegment(ctx) == nil || req == nil {
		return ctx, nil
	}

	ctx, subseg := xray.BeginSubsegment(ctx, req.Host)
	subseg.Namespace = "remote"
	subseg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{Method: req.Method, URL: req.URL.String()},
	}
	req.Header.Set(traceIDKey, subseg.DownstreamHeader().String())

	return ctx, subseg
}

// CloseSubsegment records the HTTP response status on subseg and closes it.
func CloseSubsegment(subseg *xray.Segment, resp *http.Response, err error) {
	if subseg == nil {
		return
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	} else if err != nil {
		// added to show non instrumented services in xray
		var netErr net.Error
		if (errors.As(err, &netErr) && netErr.Timeout()) || errors.Is(err, context.DeadlineExceeded) {
			statusCode = http.StatusGatewayTimeout
		} else {
			statusCode = http.StatusBadGateway
		}
	}

	xray.HttpCaptureResponse(subseg, statusCode)
	subseg.Close(err)
}

func BeginSegmentSec(req *http.Request) (context.Context, *xray.Segment) {
	return BeginSubsegment(globalCtx, req)
}

// Instrumenta la actualización del estado del segmento principal, cuando se obtiene una respuesta
// con un error.
// Establece los datos HTTP para el segmento con el método de la solicitud, la URL y el código de
// estado de respuesta. Agrega metadatos con la información del error y, finalmente, cierra el segmento.
//
// Parámetros:
// - status: el código de estado a actualizar.
// - err: la información del error para agregar como metadatos.
func EndSegmentErr(status int, err interface{}) {
	if globalSeg == nil {
		return
	}

	if statusCode != 500 && statusCode != 501 && statusCode != 502 && statusCode != 503 {
		statusCode = status
	}

	globalSeg.HTTP = &xray.HTTPData{
		Request:  &xray.RequestData{Method: method, URL: url},
		Response: &xray.ResponseData{Status: statusCode},
	}
	_ = globalSeg.AddMetadata("Error", err)
	globalSeg.Close(nil)
}
