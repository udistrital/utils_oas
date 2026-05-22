package xray

import (
	"context"
	"errors"
	"fmt"
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

type contextKey string

const (
	segmentKey contextKey = "xray_seg"
	urlKey     contextKey = "xray_url"
	methodKey  contextKey = "xray_method"
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

// Actualiza el estado del segmento principal con la respuesta de la petición.
// En caso de ser un estado 5XX, adjunta el error de la petición al segmento y lo cierra.
//
// Parámetros:
// - status: el código de estado HTTP de la respuesta.
// - err: el error generado en la petición.
func UpdateState(status int, err error) {
	if globalSeg == nil {
		return
	}
	statusCode = status
	globalSeg.HTTP = &xray.HTTPData{
		Request:  &xray.RequestData{Method: method, URL: url},
		Response: &xray.ResponseData{Status: statusCode},
	}
	if status == http.StatusInternalServerError || status == http.StatusNotImplemented || status == http.StatusBadGateway || status == http.StatusServiceUnavailable {
		_ = globalSeg.AddError(fmt.Errorf("%v", err))
		globalSeg.Close(nil)
	}
}

// Función que maneja errores 5xx.
//
// Toma un error como parámetro y hace lo siguiente:
// - Establece el código de estado del segmento en 500.
// - Agrega metadatos al segmento para especificar el error.
// - Agrega un error al segmento.
// - Cierra el segmento.
func ErrorController5xx(err error) {
	if globalSeg == nil {
		return
	}

	statusCode = 500
	globalSeg.HTTP = &xray.HTTPData{
		Request:  &xray.RequestData{Method: method, URL: url},
		Response: &xray.ResponseData{Status: statusCode},
	}
	_ = globalSeg.AddMetadata("Error", err)
	_ = globalSeg.AddError(fmt.Errorf("%v", err))
	globalSeg.Close(nil)
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

// Función creada para la creación de segmentos secundarios desde el API principal, que realizan
// seguimiento a las peticiones realizadas a otras APIs.
// Transmite, a través del Header de la petición, el ID de Traza y del segmento principal.
// Inicializa el segmento secundario y lo asigna como hijo del segmento principal.
//
// Parámetros:
// - req: puntero a Request de la petición saliente.
//
// Devoluciones:
// - seg: puntero al segmento secundario recién creado.
func BeginSegmentSec(req *http.Request) *xray.Segment {
	if globalSeg == nil {
		return nil
	}

	req.Header.Set("X-Amzn-Trace-Id", globalSeg.DownstreamHeader().String())
	_, seg := xray.BeginSegment(globalCtx, req.Host)
	seg.Lock()
	seg.Origin = url
	seg.HTTP = &xray.HTTPData{
		Request:  &xray.RequestData{Method: req.Method, URL: req.URL.String()},
		Response: &xray.ResponseData{Status: 0},
	}

	seg.TraceID = globalSeg.TraceID
	seg.ParentID = globalSeg.ID
	seg.Unlock()
	return seg
}

// Actualiza y cierra el segmento secundario con la respuesta obtenida de la solicitud.
// Tambien detecta si la API a la que realizó la solicitud se encuentra tambien instrumentada con
// X-Ray. En caso de cumplir esta condición, elimina el segmento actual para evitar un duplicado.
//
// Esta función toma tres parámetros: resp, err y seg. El parámetro resp es de tipo *http.Response y representa
// la respuesta HTTP. El parámetro err es de tipo error y representa cualquier error que ocurrió durante la solicitud.
// El parámetro seg es de tipo *xray.Segment y representa el segmento de rayos X.
//
// No hay ningún valor de retorno para esta función.
func UpdateSegmentSec(resp *http.Response, err error, seg *xray.Segment) {
	if seg == nil {
		return
	}
	var status int
	if err != nil {
		status = 500
		_ = seg.AddError(err)
	} else {
		status = resp.StatusCode
		if resp.Header.Values("Resp-X-Amzn-Trace-Id") != nil {
			seg.Sampled = false
		}
	}
	seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: seg.HTTP.Request.Method,
			URL:    seg.HTTP.Request.URL,
		},
		Response: &xray.ResponseData{
			Status: status,
		},
	}
	seg.Close(nil)
}

// Realiza la actualización y cierre del segmento secundario y, actualiza en cascada el estado del
// segmento principal. En caso de haber algun error en la respuesta, este se propaga al segmento principal
// y se cierra.
//
// Parámetros:
// - resp: la respuesta HTTP de la petición.
// - err: error de la petición.
// - seg: puntero al segmento secundario.
func UpdateSegment(resp *http.Response, err error, seg *xray.Segment) {
	UpdateSegmentSec(resp, err, seg)
	if err != nil {
		ErrorController5xx(err)
	} else {
		UpdateState(resp.StatusCode, err)
	}
}
