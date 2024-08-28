package requestresponse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type APIResponse struct {
	Success bool        `json:"Success"`
	Status  int         `json:"Status"`
	Message interface{} `json:"Message"`
	Data    interface{} `json:"Data"`
}

// Formato de respuesta generalizado para entrega de respuesta de MID
//   - success: proceso exitoso (true) o fallido  (false)
//   - statusCode: Códigos de estado de respuesta HTTP
//   - data: información principal a entregar
//   - customMessage: mensaje informativo de estado de respuesta (variádica: acepta n messages o incluso ninguno)
//
// Retorna:
//   - respuesta formateada
func APIResponseDTO(success bool, statusCode int, data interface{}, customMessage ...interface{}) APIResponse {
	var message interface{}

	if len(customMessage) > 0 {
		if len(customMessage) == 1 {
			message = customMessage[0]
		} else {
			message = customMessage
		}
	} else {
		message = getHttpStatusMessage(success, statusCode)
	}

	return APIResponse{
		Success: success,
		Status:  statusCode,
		Message: message,
		Data:    data,
	}
}

type APIResponseMeta struct {
	Success  bool        `json:"Success"`
	Status   int         `json:"Status"`
	Message  interface{} `json:"Message"`
	Data     interface{} `json:"Data"`
	Metadata interface{} `json:"Metadata"`
}

// Formato de respuesta generalizado para entrega de respuesta de MID
//   - success: proceso exitoso (true) o fallido  (false)
//   - statusCode: Códigos de estado de respuesta HTTP
//   - data: información principal a entregar
//   - metadata: informacion adicional a la informacion principal a entregar
//   - customMessage: mensaje informativo de estado de respuesta (variádica: acepta n messages o incluso ninguno)
//
// Retorna:
//   - respuesta formateada
func APIResponseMetadataDTO(success bool, statusCode int, data interface{}, metadata interface{}, customMessage ...interface{}) APIResponseMeta {
	var message interface{}

	if len(customMessage) > 0 {
		if len(customMessage) == 1 {
			message = customMessage[0]
		} else {
			message = customMessage
		}
	} else {
		message = getHttpStatusMessage(success, statusCode)
	}

	return APIResponseMeta{
		Success:  success,
		Status:   statusCode,
		Message:  message,
		Data:     data,
		Metadata: metadata,
	}
}

func getHttpStatusMessage(success bool, statusCode int) string {
	switch statusCode {
	case http.StatusContinue: // 100
		return "Continuar"
	case http.StatusSwitchingProtocols: // 101
		return "Cambiando protocolos"
	case http.StatusProcessing: // 102
		return "Procesando"
	case http.StatusEarlyHints: // 103
		return "Pistas tempranas"

	case http.StatusOK: // 200
		return "Solicitud exitosa"
	case http.StatusCreated: // 201
		return "Recurso creado con éxito"
	case http.StatusAccepted: // 202
		return "Solicitud aceptada"
	case http.StatusNonAuthoritativeInfo: // 203
		return "Información no autorizada"
	case http.StatusNoContent: // 204
		return "Sin contenido"
	case http.StatusResetContent: // 205
		return "Restablecer contenido"
	case http.StatusPartialContent: // 206
		return "Contenido parcial"
	case http.StatusMultiStatus: // 207
		return "Multi-Status"
	case http.StatusAlreadyReported: // 208
		return "Ya reportado"
	case http.StatusIMUsed: // 226
		return "IM Used"

	case http.StatusMultipleChoices: // 300
		return "Múltiples opciones"
	case http.StatusMovedPermanently: // 301
		return "Movido permanentemente"
	case http.StatusFound: // 302
		return "Encontrado"
	case http.StatusSeeOther: // 303
		return "Ver otros"
	case http.StatusNotModified: // 304
		return "No modificado"
	case http.StatusUseProxy: // 305
		return "Usar proxy"
	// 306 es unused
	case http.StatusTemporaryRedirect: // 307
		return "Redirección temporal"
	case http.StatusPermanentRedirect: // 308
		return "Redirección permanente"

	case http.StatusBadRequest: // 400
		return "Solicitud incorrecta"
	case http.StatusUnauthorized: // 401
		return "No autorizado"
	case http.StatusPaymentRequired: // 402
		return "Pago requerido"
	case http.StatusForbidden: // 403
		return "Prohibido"
	case http.StatusNotFound: // 404
		return "Recurso no encontrado"
	case http.StatusMethodNotAllowed: // 405
		return "Método no permitido"
	case http.StatusNotAcceptable: // 406
		return "No aceptable"
	case http.StatusProxyAuthRequired: // 407
		return "Autenticación de proxy requerida"
	case http.StatusRequestTimeout: // 408
		return "Tiempo de espera de solicitud agotado"
	case http.StatusConflict: // 409
		return "Conflicto"
	case http.StatusGone: // 410
		return "Recurso ya no disponible"
	case http.StatusLengthRequired: // 411
		return "Longitud requerida"
	case http.StatusPreconditionFailed: // 412
		return "Precondición fallida"
	case http.StatusRequestEntityTooLarge: // 413
		return "Entidad de solicitud demasiado grande"
	case http.StatusRequestURITooLong: // 414
		return "URI de solicitud demasiado largo"
	case http.StatusUnsupportedMediaType: // 415
		return "Tipo de medio no soportado"
	case http.StatusRequestedRangeNotSatisfiable: // 416
		return "Rango solicitado no satisfactorio"
	case http.StatusExpectationFailed: // 417
		return "Expectativa fallida"
	case http.StatusTeapot: // 418
		return "Soy una tetera"
	case http.StatusMisdirectedRequest: // 421
		return "Solicitud mal dirigida"
	case http.StatusUnprocessableEntity: // 422
		return "Entidad no procesable"
	case http.StatusLocked: // 423
		return "Bloqueado"
	case http.StatusFailedDependency: // 424
		return "Dependencia fallida"
	case http.StatusTooEarly: // 425
		return "Demasiado pronto"
	case http.StatusUpgradeRequired: // 426
		return "Actualización requerida"
	case http.StatusPreconditionRequired: // 428
		return "Precondición requerida"
	case http.StatusTooManyRequests: // 429
		return "Demasiadas solicitudes"
	case http.StatusRequestHeaderFieldsTooLarge: // 431
		return "Campos de cabecera de solicitud demasiado grandes"
	case http.StatusUnavailableForLegalReasons: // 451
		return "No disponible por razones legales"

	case http.StatusInternalServerError: // 500
		return "Error interno del servidor"
	case http.StatusNotImplemented: // 501
		return "No implementado"
	case http.StatusBadGateway: // 502
		return "Mal gateway"
	case http.StatusServiceUnavailable: // 503
		return "Servicio no disponible"
	case http.StatusGatewayTimeout: // 504
		return "Tiempo de espera del gateway agotado"
	case http.StatusHTTPVersionNotSupported: // 505
		return "Versión HTTP no soportada"
	case http.StatusVariantAlsoNegotiates: // 506
		return "Variante también negocia"
	case http.StatusInsufficientStorage: // 507
		return "Almacenamiento insuficiente"
	case http.StatusLoopDetected: // 508
		return "Bucle detectado"
	case http.StatusNotExtended: // 510
		return "No extendido"
	case http.StatusNetworkAuthenticationRequired: // 511
		return "Autenticación de red requerida"
	default:
		if success {
			return "Operación exitosa"
		} else {
			return "Error desconocido"
		}
	}
}

// Formatea respuesta de api sin formato; en realidad solo valida que haya información
//   - dataIs: data de cualquier tipo de formato
//
// Retorna:
//   - data si existe o no si es array vacío
//   - error si existe
func ParseResonseNoFormat(dataIs interface{}) (interface{}, error) {
	data := dataIs
	switch dataIs.(type) {
	case []interface{}:
		if len(data.([]interface{})) == 0 {
			return nil, fmt.Errorf("data array is pure empty")
		}
		if len(data.([]interface{})[0].(map[string]interface{})) == 0 {
			return nil, fmt.Errorf("data array is dirty empty")
		}
	case map[string]interface{}:
		if len(data.(map[string]interface{})) == 0 {
			return nil, fmt.Errorf("data is empty")
		}
	}
	return data, nil
}

type expectedResponseFormato1 struct {
	Success bool        `json:"Success"`
	Status  string      `json:"Status"`
	Message string      `json:"Message"`
	Data    interface{} `json:"Data"`
}

// Formatea respuesta de api con formato; verifica el status y que haya información
//   - dataIs: data de cualquier tipo de formato
//
// Retorna:
//   - data si existe o no si es array vacío
//   - error si existe
func ParseResponseFormato1(resp interface{}) (interface{}, error) {
	// ? se prepara y convierte la respuesta en una estructura esperada
	expRespV1 := expectedResponseFormato1{}
	jsonString, err := json.Marshal(resp)
	if err != nil {
		return expRespV1, err
	}
	json.Unmarshal(jsonString, &expRespV1)
	// ? se corrobora nuevamente el estatus de la respuesta, por si las dudas (ha pasado que la petición retorna ok con Success false)
	_status, _ := strconv.Atoi(expRespV1.Status)
	if _status < 200 || _status > 299 || !expRespV1.Success {
		return expRespV1, fmt.Errorf("not successful response")
	}
	// ? checkeo si hay data, en querys puede retornar array vacío
	_, err = ParseResonseNoFormat(expRespV1.Data)
	if err != nil {
		return nil, err
	}

	return expRespV1.Data, nil
}

type expectedResponseFormato2 struct {
	Success bool        `json:"Success"`
	Status  int16       `json:"Status"`
	Message string      `json:"Message"`
	Data    interface{} `json:"Data"`
}

// Formatea respuesta de api con formato; verifica el status y que haya información
//   - dataIs: data de cualquier tipo de formato
//
// Retorna:
//   - data si existe o no si es array vacío
//   - error si existe
func ParseResponseFormato2(resp interface{}) (interface{}, error) {
	// ? se prepara y convierte la respuesta en una estructura esperada
	expRespV1 := expectedResponseFormato2{}
	jsonString, err := json.Marshal(resp)
	if err != nil {
		return expRespV1, err
	}
	json.Unmarshal(jsonString, &expRespV1)
	// ? se corrobora nuevamente el estatus de la respuesta, por si las dudas (ha pasado que la petición retorna ok con Success false)
	if expRespV1.Status < 200 || expRespV1.Status > 299 || !expRespV1.Success {
		return expRespV1, fmt.Errorf("not successful response")
	}
	// ? checkeo si hay data, en querys puede retornar array vacío
	_, err = ParseResonseNoFormat(expRespV1.Data)
	if err != nil {
		return nil, err
	}

	return expRespV1.Data, nil
}
