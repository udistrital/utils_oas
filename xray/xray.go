package xray

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	context2 "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/udistrital/utils_oas/ssm"
)

var GlobalContext context.Context
var SegmentName string
var AppName = beego.AppConfig.String("appname")
var StatusCode int
var Seg *xray.Segment
var URL string
var Method string
var Controller string
var capturar bool

// InitXRay inicializa la configuración de X-Ray y configura los clientes necesarios.
// Devuelve un error si ocurre algún error durante la inicialización.
func InitXRay() error {
	parameterStore, exists := os.LookupEnv("PARAMETER_STORE")
	if !exists {
		parameterStore = "preprod"
	}

	daemonAddr, err := ssm.GetParameterFromParameterStore("/" + parameterStore + "/utils/xray/DaemonAddr")
	if err != nil {
		logs.Critical("Error retrieving daemon address: %v", err)
	}

	// Establecer variables de entorno para X-Ray
	os.Setenv("AWS_XRAY_NOOP_ID", "true")
	os.Setenv("AWS_XRAY_DEBUG_MODE", "TRUE")
	// Crea una nueva sesión con configuración compartida
	XraySess, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		return err
	}
	//Activación de logs y modo Debug
	//xray.SetLogger(xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelDebug))
	//Configuración de X-Ray

	// Configuración X-Ray
	xray.Configure(xray.Config{
		DaemonAddr: daemonAddr, // Dirección dinamica del demonio de X-ray
		// DaemonAddr: "127.0.0.1:2000", // Establece la dirección y el puerto del demonio en local
		LogLevel:  "debug", // Nivel de log deseado
		LogFormat: "json",  // Formato de log deseado (text o json)
	})

	// Crea clliente para ECS
	ecsClient := ecs.New(XraySess)

	// Habilita el seguimiento de X-Ray para los clientes
	xray.AWS(ecsClient.Client)

	fmt.Println("Listed buckets successfully")

	//Filtros X-Ray al inicio y fin de la ejecución de la API.
	beego.InsertFilter("*", beego.BeforeExec, BeginSegment)
	beego.InsertFilter("*", beego.AfterExec, EndSegment, false)
	return nil
}

// Función que Crea el segmento principal asociado a la API, tomando en cuenta si
// es la API principal (a la cual se realizó la petición inicial) o una secundaria.
//
// Parámetros:
// - ctx: objeto context de Beego
//
// Variables:
// - SegmentName: nombre del segmento, equivalente al Host de la API.
// - URL: URL de la petición.
// - Method: Método de la petición.
// - GlobalContext: Inicialización de un contexto vacío para almacenar los segmentos que se generen.
// - Seg: Segmento principal de la Traza.
func BeginSegment(ctx *context2.Context) {
	if ctx.Input.Context.Request.URL.String() == "/" {
		capturar = false
	} else {
		capturar = true
	}
	SegmentName = ctx.Input.Context.Request.Host
	URL = "http://" + SegmentName + ctx.Input.Context.Request.URL.String()
	Method = ctx.Request.Method
	GlobalContext = context.Background()
	seg := BeginSegmentWithContextTP(StatusCode, ctx.Request.Header.Values("X-Amzn-Trace-Id"), ctx)
	Seg = seg
}

// Crea un nuevo subsegmento para seguimiento y lo completa con datos HTTP.
//
// Parámetros:
// - subsegment: el nombre del subsegmento.
// - method: el método HTTP utilizado en la solicitud.
// - URL: la URL de la solicitud.
// - status: el código de estado HTTP de la respuesta.
//
// Devoluciones:
// - globalContext: el contexto global.
// - subseg: el subsegmento recién creado.
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

// Actualiza y cierra el segmento principal y envía los datos del segmento y la traza a AWS X-Ray.
//
// Parámetros:
// - ctx: puntero a objeto context de Beego
func EndSegment(ctx *context2.Context) {
	// Obtener el valor de la clave "json" del contexto
	jsonValue := ctx.Input.GetData("json")
	// Convertir el valor a un mapa
	if jsonMap, ok := jsonValue.(map[string]interface{}); ok {
		// Obtiene el valor de la clave "Status" del mapa
		status, ok := jsonMap["Status"].(string)
		// Evalua si no hay errores al obtener el estado y actualiza los valores del segmento principal.
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
	Seg.Close(nil)
}

// Actualiza el estado del segmento principal con la respuesta de la petición.
// En caso de ser un estado 5XX, adjunta el error de la petición al segmento y lo cierra.
//
// Parámetros:
// - status: el código de estado HTTP de la respuesta.
// - err: el error generado en la petición.
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
	if status == 500 || status == 501 || status == 502 || status == 503 {
		Seg.AddError(fmt.Errorf("%v", err))
		Seg.Close(nil)
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
	Seg.Close(nil)
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
	if StatusCode != 500 && StatusCode != 501 && StatusCode != 502 && StatusCode != 503 {
		StatusCode = status
	}
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

// Crea el segmento principal de la API con base en los parámetros de entrada y algunas variables
// inicializadas previamente.
// Con el parámetro "traceID" evalua si se trata de el segmento principal de la traza o de un segmento
// secundario. En caso de serlo, lo relaciona con el segmento principal.
//
// Parámetros:
// - code: un número entero que representa el código de estado de la respuesta del segmento.
// - traceID: un segmento de cadenas que representa el ID de seguimiento del segmento.
// - ctx: puntero a objeto context de Beego.
//
// Devoluciones:
// - seg: puntero al segmento principal recién creado.
func BeginSegmentWithContextTP(code int, traceID []string, ctx *context2.Context) *xray.Segment {
	suffix := ""
	if strings.HasPrefix(SegmentName, "pruebas") {
		suffix = "_test"
	}

	ctx2, seg := xray.BeginSegment(GlobalContext, AppName+suffix)
	seg.Sampled = !capturar
	GlobalContext = ctx2
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
	if traceID != nil {
		traceID := strings.Trim(traceID[0], "[]")
		id, parent := GetTraceIDAndParentID(traceID)
		seg.TraceID = id
		seg.ParentID = parent
		ctx.ResponseWriter.Header().Set("Resp-X-Amzn-Trace-Id", "true")
	}
	return seg
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
	req.Header.Set("X-Amzn-Trace-Id", Seg.DownstreamHeader().String())
	_, seg := xray.BeginSegment(GlobalContext, req.Host)
	seg.Lock()
	seg.Origin = URL
	seg.HTTP = &xray.HTTPData{
		Request: &xray.RequestData{
			Method: req.Method,
			URL:    req.URL.String(),
		},
		Response: &xray.ResponseData{
			Status: 200,
		},
	}
	seg.TraceID = Seg.TraceID
	seg.ParentID = Seg.ID
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
	var status int
	if err != nil {
		status = 500
		seg.AddError(err)
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

// Función creada para obtener el ID de la traza y del segmento principal
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
