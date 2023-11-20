package request

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/formatdata"
	"github.com/udistrital/utils_oas/xray"
)

const (
	rethrow_panic   string = "_____rethrow"
	JSON_error      string = "Error en el archivo JSON"
	ErrorParametros string = "Error en los parametros de ingreso"
	ErrorBody       string = "Cuerpo de la peticion invalido"
	AppJson         string = "application/json"
)

type Excep struct {
	error interface{}
}

// Envia una petición con datos al endpoint indicado y extrae la respuesta del campo Data para retornarla
func SendRequestNew(endpoint string, route string, trequest string, target interface{}, datajson interface{}) error {
	url := beego.AppConfig.String("ProtocolAdmin") + beego.AppConfig.String(endpoint) + route
	var response map[string]interface{}
	var err error
	err = SendJson(url, trequest, &response, &datajson)
	err = ExtractData(response, target, err)
	return err
}

// Envia una petición con datos a endpoints que responden con el body sin encapsular
func SendRequestLegacy(endpoint string, route string, trequest string, target interface{}, datajson interface{}) error {
	url := beego.AppConfig.String("ProtocolAdmin") + beego.AppConfig.String(endpoint) + route
	if err := SendJson(url, trequest, &target, &datajson); err != nil {
		return err
	}
	return nil
}

// Envia una petición al endpoint indicado y extrae la respuesta del campo Data para retornarla
func GetRequestNew(endpoint string, route string, target interface{}) error {
	url := beego.AppConfig.String("ProtocolAdmin") + beego.AppConfig.String(endpoint) + route
	var response map[string]interface{}
	var err error
	err = GetJson(url, &response)
	err = ExtractData(response, &target, err)
	return err
}

// Envia una petición a endpoints que responden con el body sin encapsular
func GetRequestLegacy(endpoint string, route string, target interface{}) error {
	url := beego.AppConfig.String("ProtocolAdmin") + beego.AppConfig.String(endpoint) + route
	if err := GetJson(url, target); err != nil {
		return err
	}
	return nil
}

// Esta función extrae la información cuando se recibe encapsulada en una estructura
// y da manejo a las respuestas que contienen arreglos de objetos vacíos
func ExtractData(respuesta map[string]interface{}, v interface{}, err2 error) error {
	var err error
	if err2 != nil {
		return err2
	}
	if respuesta["Success"] == false {
		err = errors.New(fmt.Sprint(respuesta["Data"], respuesta["Message"]))
		panic(err)
	}
	datatype := fmt.Sprintf("%v", respuesta["Data"])
	switch datatype {
	case "map[]", "[map[]]": // response vacio
		break
	default:
		err = formatdata.FillStruct(respuesta["Data"], &v)
		respuesta = nil
	}
	return err
}

func Commit(f func()) (err Excep) {
	defer func() {
		err.error = recover()
	}()
	f()
	return
}

func (err Excep) Rollback(f func(response interface{}, excep interface{}), params interface{}) {
	if err.error != nil {
		defer func() {
			if excep := recover(); excep != nil {
				if excep == rethrow_panic {
					excep = err.error
				}
				panic(excep)
			}
		}()
		f(params, err.error)
	}
}

func diff(a, b time.Time) (year, month, day int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)

	// Normalize negative values

	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

func diff2(a, b time.Time) (year, month, day int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	oneDay := time.Hour * 5
	a = a.Add(oneDay)
	b = b.Add(oneDay)
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)

	// Normalize negative values

	if day < 0 {
		// days in month: p
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

func Diff3(a, b time.Time) (year, month, day int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	oneDay := time.Hour * 5
	a = a.Add(oneDay)
	b = b.Add(oneDay)
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year = y2 - y1
	month = int(M2 - M1)
	day = d2 - d1

	if day < 0 {

		day = (30 - d1) + d2
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

// Valida que el body recibido en la petición tenga contenido válido
func ValidarBody(body []byte) (valid bool, err error) {
	var test interface{}
	if err = json.Unmarshal(body, &test); err != nil {
		return false, err
	} else {
		content := fmt.Sprintf("%v", test)
		fmt.Println(content)
		switch content {
		case "map[]", "[map[]]": // body vacio
			return false, nil
		}
	}
	return true, nil
}

// Quita el formato de moneda a un string y lo convierte en valor flotante
func DeformatNumber(formatted string) (number float64) {
	formatted = strings.ReplaceAll(formatted, ",", "")
	formatted = strings.Trim(formatted, "$")
	number, _ = strconv.ParseFloat(formatted, 64)
	return
}

// Obtiene los datos del usuario autenticado
func GetUsuario(usuario string) (nombreUsuario map[string]interface{}, err error) {
	if len(usuario) > 0 {
		var decData map[string]interface{}
		if data, err6 := base64.StdEncoding.DecodeString(usuario); err6 != nil {
			return nombreUsuario, err6
		} else {
			if err7 := json.Unmarshal(data, &decData); err7 != nil {
				return nombreUsuario, err7
			}
		}
		nombreUsuario = decData["user"].(map[string]interface{})
	}
	return nombreUsuario, err
}

// Manejo único de errores para controladores sin repetir código
func ErrorController(c beego.Controller, controller string) {
	if err := recover(); err != nil {
		logs.Error(err)
		localError := err.(map[string]interface{})
		c.Data["mesaage"] = (beego.AppConfig.String("appname") + "/" + controller + "/" + (localError["funcion"]).(string))
		c.Data["data"] = (localError["err"])
		xray.EndSegmentErr(http.StatusBadRequest, localError["err"])
		if status, ok := localError["status"]; ok {
			c.Abort(status.(string))
		} else {
			c.Abort("500")
		}
	}
}

func JsonDebug(i interface{}) {
	formatdata.JsonPrint(i)
	fmt.Println()
}

func Iguales(a interface{}, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func LimpiezaRespuestaRefactor(respuesta map[string]interface{}, v interface{}) {
	fmt.Println("---------Entrada 2--------")
	b, err := json.Marshal(respuesta["Data"])
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &v)
}

func CalcularDias(FechaInicio time.Time, FechaFin time.Time) (diasLaborados float64, meses float64) {
	fmt.Println("---------Entrada--------")
	var a, m, d int
	var mesesContrato float64
	var diasContrato float64
	if FechaFin.IsZero() {
		FechaFin2 := time.Now()
		a, m, d = Diff3(FechaInicio, FechaFin2)
		mesesContrato = (float64(a * 12)) + float64(m) + (float64(d) / 30)
		diasContrato = mesesContrato * 30
	} else {
		a, m, d = Diff3(FechaInicio, FechaFin)
		mesesContrato = (float64(a * 12)) + float64(m) + (float64(d) / 30)
		diasContrato = mesesContrato * 30
	}
	return diasContrato, mesesContrato

}

func CalcularSemanas(diasLiquidados float64) (semanas int) {
	aux := diasLiquidados / 7

	if aux <= 1 {
		return 1
	} else if aux <= 2 {
		return 2
	} else if aux <= 3 {
		return 3
	} else {
		return 4
	}
}

func Roundf(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Funcionalidad para saber la cantidad de dias de un mes
func DaysInMonth(month, year int) int {
	switch time.Month(month) {
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) { // leap year
			return 29
		}
		return 28
	default:
		return 31
	}
}

func SortSlice(slice *[]map[string]interface{}, parameter string) {
	sort.SliceStable(*slice, func(i, j int) bool {
		var a int
		var b int
		if reflect.TypeOf((*slice)[j][parameter]).String() == "string" {
			b, _ = strconv.Atoi((*slice)[j][parameter].(string))
		} else {
			b = int((*slice)[j][parameter].(float64))
		}

		if reflect.TypeOf((*slice)[i][parameter]).String() == "string" {
			a, _ = strconv.Atoi((*slice)[i][parameter].(string))
		} else {
			a = int((*slice)[i][parameter].(float64))
		}
		return a < b
	})
}
